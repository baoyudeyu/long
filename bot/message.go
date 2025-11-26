package bot

import (
	"dragon-alert-bot/dragon"
	"fmt"
	"strings"
)

// FormatAlertMessage 格式化提醒消息
func FormatAlertMessage(results []*dragon.PatternResult, currentData *dragon.CurrentLotteryInfo) string {
	if len(results) == 0 {
		return ""
	}

	var text strings.Builder
	text.WriteString("🔥 <b>长龙提醒</b>\n")

	if currentData != nil {
		text.WriteString(fmt.Sprintf("<code>%s</code>期 开奖号码: <b>%s=%d</b> %s%s\n",
			currentData.Qihao,
			currentData.OpenNum,
			currentData.SumValue,
			currentData.Size,
			currentData.Parity,
		))
	} else {
		text.WriteString(fmt.Sprintf("当前期号: <code>%s</code>期\n", results[0].CurrentQihao))
	}

	// 按属性类型分组
	sizeResults := []*dragon.PatternResult{}
	parityResults := []*dragon.PatternResult{}
	sumResults := []*dragon.PatternResult{}
	comboResults := []*dragon.PatternResult{}

	for _, r := range results {
		switch r.AttributeType {
		case "size":
			sizeResults = append(sizeResults, r)
		case "parity":
			parityResults = append(parityResults, r)
		case "sum":
			sumResults = append(sumResults, r)
		case "size_parity":
			comboResults = append(comboResults, r)
		}
	}

	// 排序函数：按Count降序排序
	sortResults := func(results []*dragon.PatternResult) {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].Count < results[j].Count {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}

	// 大小长龙
	if len(sizeResults) > 0 {
		sortResults(sizeResults)
		text.WriteString("<blockquote>📊 <b>【大小长龙】</b></blockquote>\n")
		for _, r := range sizeResults {
			text.WriteString(formatSingleResult(r))
		}
	}

	// 单双长龙
	if len(parityResults) > 0 {
		sortResults(parityResults)
		text.WriteString("<blockquote>🎯 <b>【单双长龙】</b></blockquote>\n")
		for _, r := range parityResults {
			text.WriteString(formatSingleResult(r))
		}
	}

	// 和值长龙
	if len(sumResults) > 0 {
		sortResults(sumResults)
		text.WriteString("<blockquote>🔢 <b>【和值长龙】</b></blockquote>\n")
		for _, r := range sumResults {
			text.WriteString(formatSingleResult(r))
		}
	}

	// 组合长龙
	if len(comboResults) > 0 {
		sortResults(comboResults)
		text.WriteString("<blockquote>🔄 <b>【组合长龙】</b></blockquote>\n")
		for _, r := range comboResults {
			text.WriteString(formatSingleResult(r))
		}
	}

	return strings.TrimRight(text.String(), "\n")
}

func formatSingleResult(r *dragon.PatternResult) string {
	patternNames := map[string]string{
		"a":     "连续",
		"ab":    "交替",
		"abb":   "abb",
		"ab_ac": "固定交替",
		"ab_cd": "双交替",
		"abab":  "组合重复",
	}

	// 计算显示的次数
	displayCount := r.Count
	countUnit := "期"

	// abb格式按组计算（3个元素=1组）
	if r.PatternType == "abb" {
		displayCount = r.Count / 3
		countUnit = "组"
	}

	// 格式化模式详情，让它更直观
	pattern := r.PatternDetail
	if r.PatternType == "abb" {
		// abb格式用括号分组显示
		parts := strings.Split(r.PatternDetail, " ")
		var groups []string
		for i := 0; i < len(parts); i += 3 {
			if i+2 < len(parts) {
				groups = append(groups, fmt.Sprintf("(%s %s %s)", parts[i], parts[i+1], parts[i+2]))
			} else if i < len(parts) {
				// 不完整的部分
				remaining := parts[i:]
				groups = append(groups, strings.Join(remaining, " "))
			}
		}
		pattern = strings.Join(groups, " ")
	}

	return fmt.Sprintf("  • %s格式 连续<b>%d%s</b>\n    <code>%s</code>\n    起始: %s期\n\n",
		patternNames[r.PatternType],
		displayCount,
		countUnit,
		pattern,
		r.StartQihao,
	)
}
