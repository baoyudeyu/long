package dragon

import (
	"dragon-alert-bot/db"
	"dragon-alert-bot/lottery"
)

type Analyzer struct {
	monitor *lottery.Monitor
}

func NewAnalyzer(monitor *lottery.Monitor) *Analyzer {
	return &Analyzer{
		monitor: monitor,
	}
}

// Analyze 分析长龙
func (a *Analyzer) Analyze(newData *lottery.LotteryData) []*PatternResult {
	// 获取历史数据（最近500期，足够检测长龙）
	historyData, err := a.monitor.GetHistoryData(500)
	if err != nil {
		return nil
	}

	if len(historyData) == 0 {
		return nil
	}

	// 转换为属性列表（注意：数据库返回的是从新到旧，需要反转为从旧到新）
	var attrs []lottery.Attributes
	for i := len(historyData) - 1; i >= 0; i-- {
		attrs = append(attrs, historyData[i].CalculateAttributes())
	}

	// 现在attrs是从旧到新排列，attrs[0]是最老的，attrs[len-1]是最新的

	// 检测所有模式（使用最小阈值进行检测）
	var results []*PatternResult

	// 对每个属性类型，按优先级检测（优先长龙和复杂模式）
	for _, attrType := range []string{"size", "parity", "sum"} {
		var attrResults []*PatternResult

		// 检测所有格式
		if result := CheckPatternA(attrs, attrType, 2); result.Matched {
			attrResults = append(attrResults, result)
		}
		if result := CheckPatternAB(attrs, attrType, 2); result.Matched {
			attrResults = append(attrResults, result)
		}
		if result := CheckPatternABB(attrs, attrType, 3); result.Matched {
			attrResults = append(attrResults, result)
		}

		// 选择最长的结果（避免重叠）
		if len(attrResults) > 0 {
			longest := attrResults[0]
			for _, r := range attrResults {
				if r.Count > longest.Count {
					longest = r
				}
			}
			results = append(results, longest)
		}
	}

	// ab,ac 格式检测（最小2次）
	if result := CheckPatternABAC(attrs, 2); result.Matched {
		results = append(results, result)
	}

	// ab,cd 格式检测（最小2次）
	if result := CheckPatternABCD(attrs, 2); result.Matched {
		results = append(results, result)
	}

	// abab 格式检测（最小2次：组合重复）
	if result := CheckPatternABAB(attrs, 2); result.Matched {
		results = append(results, result)
	}

	return results
}

// GetActiveChats 获取所有启用的群组
func (a *Analyzer) GetActiveChats() ([]int64, error) {
	rows, err := db.WriteDB.Query("SELECT chat_id FROM chat_configs WHERE enabled = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatIDs []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			continue
		}
		chatIDs = append(chatIDs, chatID)
	}

	return chatIDs, nil
}

// GetChatRules 获取群组的规则配置
func (a *Analyzer) GetChatRules(chatID int64) ([]db.DragonRule, error) {
	rows, err := db.WriteDB.Query(`
		SELECT id, chat_id, pattern_type, attribute_type, threshold, enabled, created_at, updated_at 
		FROM dragon_rules 
		WHERE chat_id = ? AND enabled = TRUE
	`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []db.DragonRule
	for rows.Next() {
		var rule db.DragonRule
		err := rows.Scan(&rule.ID, &rule.ChatID, &rule.PatternType, &rule.AttributeType,
			&rule.Threshold, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
		if err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// FilterResultsByRules 根据规则过滤结果
func (a *Analyzer) FilterResultsByRules(results []*PatternResult, rules []db.DragonRule) []*PatternResult {
	var filtered []*PatternResult

	for _, result := range results {
		for _, rule := range rules {
			if result.PatternType == rule.PatternType &&
				result.AttributeType == rule.AttributeType &&
				result.Count >= rule.Threshold {
				filtered = append(filtered, result)
				break
			}
		}
	}

	return filtered
}
