package bot

import (
	"dragon-alert-bot/db"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 检查用户是否是群组管理员
func isAdmin(chatID int64, userID int64) bool {
	chatConfig := tgbotapi.ChatConfigWithUser{
		ChatID: chatID,
		UserID: userID,
	}

	member, err := BotAPI.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: chatConfig})
	if err != nil {
		log.Printf("[权限检查] 获取用户信息失败: %v", err)
		return false
	}

	// 检查是否为创建者或管理员
	return member.Status == "creator" || member.Status == "administrator"
}

func handleCommand(message *tgbotapi.Message) {
	if !message.IsCommand() {
		return
	}

	chatID := message.Chat.ID
	command := message.Command()

	log.Printf("[Bot命令] %s from %d (@%s)", command, chatID, message.Chat.Title)

	switch command {
	case "start":
		handleStart(message)
	case "long":
		handleDragon(message)
	case "data":
		handleData(message)
	}
}

func handleStart(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// 判断是群组还是私聊
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		text := `欢迎使用长龙提醒机器人！🎲

功能：
• 自动监测开奖数据
• 识别各种长龙模式
• 自定义提醒规则

命令：
/long - 配置长龙提醒（仅管理员）`

		msg := tgbotapi.NewMessage(chatID, text)
		BotAPI.Send(msg)

		// 异步初始化群组配置
		go ensureChatConfig(chatID)
	} else {
		text := `欢迎使用长龙提醒机器人！🎲

⚠️ 本机器人仅支持群组使用

功能特点：
• 自动监测开奖数据
• 识别多种长龙模式
• 灵活的自定义规则

使用步骤：
1. 点击下方按钮添加到群组
2. 在群组中发送 /long 命令
3. 管理员可配置提醒规则`

		// 创建内联按钮
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"➕ 添加机器人到群组",
					fmt.Sprintf("https://t.me/%s?startgroup=1", BotAPI.Self.UserName),
				),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		BotAPI.Send(msg)
	}
}

func handleDragon(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// 只允许在群组中使用
	if message.Chat.Type != "group" && message.Chat.Type != "supergroup" {
		text := "⚠️ 长龙提醒仅支持群组使用\n\n请点击下方按钮将机器人添加到群组"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"➕ 添加到群组",
					fmt.Sprintf("https://t.me/%s?startgroup=1", BotAPI.Self.UserName),
				),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		BotAPI.Send(msg)
		return
	}

	// 检查权限（只有群组管理员可以配置）
	member, err := BotAPI.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: message.From.ID,
		},
	})

	if err != nil || (member.Status != "creator" && member.Status != "administrator") {
		msg := tgbotapi.NewMessage(chatID, "⚠️ 仅限群组管理员操作")
		BotAPI.Send(msg)
		return
	}

	// 异步确保配置存在，不阻塞响应
	go ensureChatConfig(chatID)

	// 显示主菜单
	showMainMenu(chatID, 0)
}

func handleData(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// 获取群组数量
	var totalGroups int
	db.WriteDB.QueryRow("SELECT COUNT(*) FROM chat_configs WHERE chat_id < 0").Scan(&totalGroups)

	// 获取启用的群组数量
	var enabledGroups int
	db.WriteDB.QueryRow("SELECT COUNT(*) FROM chat_configs WHERE chat_id < 0 AND enabled = TRUE").Scan(&enabledGroups)

	// 获取总规则数
	var totalRules int
	db.WriteDB.QueryRow("SELECT COUNT(*) FROM dragon_rules WHERE enabled = TRUE").Scan(&totalRules)

	// 获取活跃长龙数量
	var activeDragons int
	db.WriteDB.QueryRow("SELECT COUNT(*) FROM dragon_alerts WHERE status = 'active'").Scan(&activeDragons)

	// 获取今日提醒次数（需要添加统计表，暂时显示活跃长龙）
	text := fmt.Sprintf(`📊 <b>机器人数据统计</b>

👥 <b>群组数据</b>
• 总群组数: <code>%d</code>
• 启用提醒: <code>%d</code>
• 禁用提醒: <code>%d</code>

⚙️ <b>配置数据</b>
• 启用规则: <code>%d</code> 条

🔥 <b>长龙数据</b>
• 活跃长龙: <code>%d</code> 个

💡 使用 /long 配置长龙提醒`,
		totalGroups,
		enabledGroups,
		totalGroups-enabledGroups,
		totalRules,
		activeDragons,
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	BotAPI.Send(msg)
}

func ensureChatConfig(chatID int64) {
	// 只为群组创建配置（chatID < 0 为群组）
	if chatID > 0 {
		return
	}

	// 检查配置是否存在
	var exists bool
	err := db.WriteDB.QueryRow("SELECT EXISTS(SELECT 1 FROM chat_configs WHERE chat_id = ?)", chatID).Scan(&exists)
	if err != nil || !exists {
		// 创建默认配置
		db.WriteDB.Exec("INSERT INTO chat_configs (chat_id, enabled) VALUES (?, TRUE)", chatID)

		// 创建默认规则
		createDefaultRules(chatID)

		log.Printf("[配置初始化] 群组:%d", chatID)
	}
}

// createDefaultRules 创建默认规则
func createDefaultRules(chatID int64) {
	defaultRules := []struct {
		pattern   string
		attribute string
		threshold int
	}{
		{"a", "size", 5},
		{"a", "parity", 5},
		{"a", "sum", 5},
		{"ab", "size", 2},
		{"ab", "parity", 2},
		{"ab", "sum", 2},
		{"abb", "size", 2},
		{"abb", "parity", 2},
		{"abb", "sum", 2},
		{"ab_ac", "size_parity", 2},
		{"ab_cd", "size_parity", 2},
		{"abab", "size_parity", 2},
	}

	for _, rule := range defaultRules {
		db.WriteDB.Exec(`
			INSERT INTO dragon_rules (chat_id, pattern_type, attribute_type, threshold, enabled)
			VALUES (?, ?, ?, ?, TRUE)
			ON DUPLICATE KEY UPDATE threshold = ?, enabled = TRUE
		`, chatID, rule.pattern, rule.attribute, rule.threshold, rule.threshold)
	}
}

// ensureDefaultRules 确保规则存在
func ensureDefaultRules(chatID int64) {
	defaultRules := []struct {
		pattern   string
		attribute string
		threshold int
	}{
		{"a", "size", 5},
		{"a", "parity", 5},
		{"a", "sum", 5},
		{"ab", "size", 2},
		{"ab", "parity", 2},
		{"ab", "sum", 2},
		{"abb", "size", 2},
		{"abb", "parity", 2},
		{"abb", "sum", 2},
		{"ab_ac", "size_parity", 2},
		{"ab_cd", "size_parity", 2},
		{"abab", "size_parity", 2},
	}

	for _, rule := range defaultRules {
		var exists bool
		err := db.WriteDB.QueryRow(`
			SELECT EXISTS(SELECT 1 FROM dragon_rules 
			WHERE chat_id = ? AND pattern_type = ? AND attribute_type = ?)
		`, chatID, rule.pattern, rule.attribute).Scan(&exists)

		if err != nil || !exists {
			db.WriteDB.Exec(`
				INSERT INTO dragon_rules (chat_id, pattern_type, attribute_type, threshold, enabled)
				VALUES (?, ?, ?, ?, TRUE)
			`, chatID, rule.pattern, rule.attribute, rule.threshold)
		}
	}
}
