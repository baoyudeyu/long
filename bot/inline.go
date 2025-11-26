package bot

import (
	"dragon-alert-bot/db"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	data := callback.Data

	// 立即回应回调（最快响应，防止加载动画）
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	BotAPI.Request(callbackConfig)

	// 全异步处理（包括权限检查）
	go func() {
		// 异步检查管理员权限，非管理员直接忽略
		if !isAdmin(chatID, callback.From.ID) {
			return
		}

		// 处理回调
		parts := strings.Split(data, "_")
		if len(parts) < 2 {
			return
		}

		action := parts[1]

		switch action {
		case "main":
			showMainMenu(chatID, messageID)
		case "toggle":
			toggleDragonAlert(chatID, messageID)
		case "size":
			showAttributeMenu(chatID, messageID, "size", "大小")
		case "parity":
			showAttributeMenu(chatID, messageID, "parity", "单双")
		case "sum":
			showAttributeMenu(chatID, messageID, "sum", "和值")
		case "combo":
			showComboMenu(chatID, messageID)
		case "status":
			showStatusMenu(chatID, messageID)
		case "refresh":
			showStatusMenu(chatID, messageID)
		case "set":
			if len(parts) >= 5 {
				handleSetRule(chatID, messageID, parts[2], parts[3], parts[4])
			}
		case "combo2":
			if len(parts) >= 4 {
				handleComboRule(chatID, messageID, parts[2], parts[3])
			}
		}
	}()
}

func showMainMenu(chatID int64, messageID int) {
	// 获取当前启用状态
	var enabled bool
	db.WriteDB.QueryRow("SELECT enabled FROM chat_configs WHERE chat_id = ?", chatID).Scan(&enabled)

	status := "❌ 已禁用"
	toggleText := "✅ 启用提醒"
	if enabled {
		status = "✅ 已启用"
		toggleText = "❌ 禁用提醒"
	}

	text := fmt.Sprintf(`🎲 长龙提醒配置
当前状态: %s`, status)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(toggleText, "dragon_toggle"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 配置大小长龙", "dragon_size"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 配置单双长龙", "dragon_parity"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔢 配置和值长龙", "dragon_sum"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 配置组合长龙", "dragon_combo"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 查看配置状态", "dragon_status"),
		),
	)

	if messageID > 0 {
		msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
		msg.ReplyMarkup = &keyboard
		BotAPI.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		BotAPI.Send(msg)
	}
}

func toggleDragonAlert(chatID int64, messageID int) {
	// 切换启用状态
	_, err := db.WriteDB.Exec("UPDATE chat_configs SET enabled = NOT enabled WHERE chat_id = ?", chatID)
	if err != nil {
		log.Printf("切换状态失败: %v", err)
	}

	showMainMenu(chatID, messageID)
}

func showAttributeMenu(chatID int64, messageID int, attrType, attrName string) {
	ensureDefaultRules(chatID)

	// 获取规则配置
	rows, err := db.WriteDB.Query(`
		SELECT pattern_type, threshold, enabled 
		FROM dragon_rules 
		WHERE chat_id = ? AND attribute_type = ?
		ORDER BY 
			CASE pattern_type
				WHEN 'a' THEN 1
				WHEN 'ab' THEN 2
				WHEN 'abb' THEN 3
			END
	`, chatID, attrType)
	if err != nil {
		log.Printf("查询规则失败: %v", err)
		return
	}
	defer rows.Close()

	rules := make(map[string]struct {
		threshold int
		enabled   bool
	})

	for rows.Next() {
		var pattern string
		var threshold int
		var enabled bool
		rows.Scan(&pattern, &threshold, &enabled)
		rules[pattern] = struct {
			threshold int
			enabled   bool
		}{threshold, enabled}
	}

	text := fmt.Sprintf("🎲 %s长龙配置\n[+][-]调整触发值 | 点击名称切换启用", attrName)

	var buttons [][]tgbotapi.InlineKeyboardButton

	patterns := []struct {
		key  string
		name string
	}{
		{"a", "a格式(连续)"},
		{"ab", "ab格式(交替)"},
		{"abb", "abb格式(A-B-B组)"},
	}

	for _, p := range patterns {
		rule, exists := rules[p.key]
		if !exists {
			rule.threshold = 5
			if p.key == "ab" || p.key == "abb" {
				rule.threshold = 2
			}
			rule.enabled = true
		}

		statusIcon := "✅"
		if !rule.enabled {
			statusIcon = "❌"
		}

		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", statusIcon, p.name),
				fmt.Sprintf("dragon_set_%s_%s_toggle", attrType, p.key),
			),
		))

		// a格式用"次"，其他格式用"组"
		unit := "次"
		if p.key != "a" {
			unit = "组"
		}

		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➖", fmt.Sprintf("dragon_set_%s_%s_dec", attrType, p.key)),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("触发: %d%s", rule.threshold, unit), "dragon_noop"),
			tgbotapi.NewInlineKeyboardButtonData("➕", fmt.Sprintf("dragon_set_%s_%s_inc", attrType, p.key)),
		))
	}

	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀️ 返回主菜单", "dragon_main"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ReplyMarkup = &keyboard
	BotAPI.Send(msg)
}

func showComboMenu(chatID int64, messageID int) {
	ensureDefaultRules(chatID)

	// 获取组合规则配置
	rows, err := db.WriteDB.Query(`
		SELECT pattern_type, threshold, enabled 
		FROM dragon_rules 
		WHERE chat_id = ? AND attribute_type = 'size_parity'
	`, chatID)
	if err != nil {
		log.Printf("查询组合规则失败: %v", err)
		return
	}
	defer rows.Close()

	rules := make(map[string]struct {
		threshold int
		enabled   bool
	})

	for rows.Next() {
		var pattern string
		var threshold int
		var enabled bool
		rows.Scan(&pattern, &threshold, &enabled)
		rules[pattern] = struct {
			threshold int
			enabled   bool
		}{threshold, enabled}
	}

	text := "🔄 组合长龙配置\n大小+单双组合 | [+][-]调整触发值"

	var buttons [][]tgbotapi.InlineKeyboardButton

	patterns := []struct {
		key  string
		name string
	}{
		{"ab_ac", "ab,ac格式(固定+交替)"},
		{"ab_cd", "ab,cd格式(同时交替)"},
		{"abab", "abab格式(组合重复)"},
	}

	for _, p := range patterns {
		rule, exists := rules[p.key]
		if !exists {
			rule.threshold = 2
			rule.enabled = true
		}

		statusIcon := "✅"
		if !rule.enabled {
			statusIcon = "❌"
		}

		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", statusIcon, p.name),
				fmt.Sprintf("dragon_combo2_%s_toggle", p.key),
			),
		))

		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➖", fmt.Sprintf("dragon_combo2_%s_dec", p.key)),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("触发: %d组", rule.threshold), "dragon_noop"),
			tgbotapi.NewInlineKeyboardButtonData("➕", fmt.Sprintf("dragon_combo2_%s_inc", p.key)),
		))
	}

	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀️ 返回主菜单", "dragon_main"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ReplyMarkup = &keyboard
	BotAPI.Send(msg)
}

func showStatusMenu(chatID int64, messageID int) {
	// 获取所有规则
	rows, err := db.WriteDB.Query(`
		SELECT pattern_type, attribute_type, threshold, enabled 
		FROM dragon_rules 
		WHERE chat_id = ?
		ORDER BY attribute_type, pattern_type
	`, chatID)
	if err != nil {
		return
	}
	defer rows.Close()

	var enabledCount int
	var text strings.Builder
	text.WriteString("📋 配置状态\n")

	attrNames := map[string]string{
		"size":        "📊大小",
		"parity":      "🎯单双",
		"sum":         "🔢和值",
		"size_parity": "🔄组合",
	}

	patternNames := map[string]string{
		"a":     "a",
		"ab":    "ab",
		"abb":   "abb",
		"ab_ac": "ab,ac",
		"ab_cd": "ab,cd",
		"abab":  "abab",
	}

	currentAttr := ""
	for rows.Next() {
		var pattern, attr string
		var threshold int
		var enabled bool
		rows.Scan(&pattern, &attr, &threshold, &enabled)

		if attr != currentAttr {
			if currentAttr != "" {
				text.WriteString("\n")
			}
			text.WriteString(fmt.Sprintf("%s: ", attrNames[attr]))
			currentAttr = attr
		}

		status := "✅"
		if !enabled {
			status = "❌"
		} else {
			enabledCount++
		}

		// a格式显示次数，其他格式显示组数
		unit := "次"
		if pattern != "a" {
			unit = "组"
		}

		text.WriteString(fmt.Sprintf("%s%s:%d%s ", status, patternNames[pattern], threshold, unit))
	}

	text.WriteString(fmt.Sprintf("\n\n已启用 %d 条规则", enabledCount))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 刷新", "dragon_refresh"),
			tgbotapi.NewInlineKeyboardButtonData("◀️ 返回", "dragon_main"),
		),
	)

	msg := tgbotapi.NewEditMessageText(chatID, messageID, text.String())
	msg.ReplyMarkup = &keyboard
	BotAPI.Send(msg)
}

func handleSetRule(chatID int64, messageID int, attrType, pattern, action string) {
	// 统一步长为1（所有类型都按组调整）
	switch action {
	case "inc":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET threshold = LEAST(threshold + 1, 20) 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = ?
		`, chatID, pattern, attrType)

	case "dec":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET threshold = GREATEST(threshold - 1, 1) 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = ?
		`, chatID, pattern, attrType)

	case "toggle":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET enabled = NOT enabled 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = ?
		`, chatID, pattern, attrType)
	}

	// 快速响应：异步刷新
	attrNames := map[string]string{
		"size":   "大小",
		"parity": "单双",
		"sum":    "和值",
	}

	go showAttributeMenu(chatID, messageID, attrType, attrNames[attrType])
}

func handleComboRule(chatID int64, messageID int, pattern, action string) {
	switch action {
	case "inc":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET threshold = LEAST(threshold + 1, 20) 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = 'size_parity'
		`, chatID, pattern)

	case "dec":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET threshold = GREATEST(threshold - 1, 1) 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = 'size_parity'
		`, chatID, pattern)

	case "toggle":
		db.WriteDB.Exec(`
			UPDATE dragon_rules 
			SET enabled = NOT enabled 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = 'size_parity'
		`, chatID, pattern)
	}

	// 快速响应：异步刷新
	go showComboMenu(chatID, messageID)
}
