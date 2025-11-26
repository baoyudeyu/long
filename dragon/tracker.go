package dragon

import (
	"dragon-alert-bot/db"
	"time"
)

type Tracker struct{}

func NewTracker() *Tracker {
	return &Tracker{}
}

// TrackDragon 跟踪长龙状态
func (t *Tracker) TrackDragon(chatID int64, result *PatternResult) (shouldAlert bool, isNew bool) {
	// 查找活跃的长龙记录
	var alert db.DragonAlert
	err := db.WriteDB.QueryRow(`
		SELECT id, chat_id, pattern_type, attribute_type, start_qihao, current_qihao, 
		       count, pattern_detail, last_alert_count, status, created_at, updated_at
		FROM dragon_alerts 
		WHERE chat_id = ? AND pattern_type = ? AND attribute_type = ? AND status = 'active'
		ORDER BY id DESC LIMIT 1
	`, chatID, result.PatternType, result.AttributeType).Scan(
		&alert.ID, &alert.ChatID, &alert.PatternType, &alert.AttributeType,
		&alert.StartQihao, &alert.CurrentQihao, &alert.Count, &alert.PatternDetail,
		&alert.LastAlertCount, &alert.Status, &alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		// 没有活跃记录，创建新记录
		_, err = db.WriteDB.Exec(`
			INSERT INTO dragon_alerts 
			(chat_id, pattern_type, attribute_type, start_qihao, current_qihao, count, pattern_detail, last_alert_count, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')
		`, chatID, result.PatternType, result.AttributeType, result.StartQihao, result.CurrentQihao,
			result.Count, result.PatternDetail, result.Count)

		if err != nil {
			return false, false
		}

		return true, true // 新长龙，需要提醒
	}

	// 检查是否是同一个长龙的延续
	if result.StartQihao == alert.StartQihao {
		// 长龙延续，更新记录
		_, err = db.WriteDB.Exec(`
			UPDATE dragon_alerts 
			SET current_qihao = ?, count = ?, pattern_detail = ?, last_alert_count = ?, updated_at = ?
			WHERE id = ?
		`, result.CurrentQihao, result.Count, result.PatternDetail, result.Count, time.Now(), alert.ID)

		if err != nil {
			return false, false
		}

		return true, false // 延续的长龙，每次都提醒
	}

	// 旧长龙已结束，标记为结束
	db.WriteDB.Exec("UPDATE dragon_alerts SET status = 'ended' WHERE id = ?", alert.ID)

	// 创建新长龙记录
	_, err = db.WriteDB.Exec(`
		INSERT INTO dragon_alerts 
		(chat_id, pattern_type, attribute_type, start_qihao, current_qihao, count, pattern_detail, last_alert_count, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')
	`, chatID, result.PatternType, result.AttributeType, result.StartQihao, result.CurrentQihao,
		result.Count, result.PatternDetail, result.Count)

	if err != nil {
		return false, false
	}

	return true, true
}

// EndInactiveDragons 结束不活跃的长龙
func (t *Tracker) EndInactiveDragons(chatID int64, activeResults []*PatternResult) {
	// 获取所有活跃的长龙记录
	rows, err := db.WriteDB.Query(`
		SELECT id, pattern_type, attribute_type, start_qihao 
		FROM dragon_alerts 
		WHERE chat_id = ? AND status = 'active'
	`, chatID)
	if err != nil {
		return
	}
	defer rows.Close()

	type ActiveAlert struct {
		ID            int64
		PatternType   string
		AttributeType string
		StartQihao    string
	}

	var activeAlerts []ActiveAlert
	for rows.Next() {
		var alert ActiveAlert
		if err := rows.Scan(&alert.ID, &alert.PatternType, &alert.AttributeType, &alert.StartQihao); err != nil {
			continue
		}
		activeAlerts = append(activeAlerts, alert)
	}

	// 检查每个活跃记录是否还在当前结果中
	for _, alert := range activeAlerts {
		found := false
		for _, result := range activeResults {
			if result.PatternType == alert.PatternType &&
				result.AttributeType == alert.AttributeType &&
				result.StartQihao == alert.StartQihao {
				found = true
				break
			}
		}

		// 如果不在当前结果中，标记为结束
		if !found {
			db.WriteDB.Exec("UPDATE dragon_alerts SET status = 'ended' WHERE id = ?", alert.ID)
		}
	}
}
