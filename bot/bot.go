package bot

import (
	"dragon-alert-bot/config"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	BotAPI *tgbotapi.BotAPI
)

func InitBot(cfg *config.Config) error {
	var err error
	BotAPI, err = tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return err
	}

	BotAPI.Debug = false
	log.Printf("Bot 已授权: @%s (ID:%d)", BotAPI.Self.UserName, BotAPI.Self.ID)

	// 注册Bot命令菜单
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "查看欢迎信息和使用说明",
		},
		{
			Command:     "long",
			Description: "配置长龙提醒（仅群组管理员）",
		},
		{
			Command:     "data",
			Description: "查看机器人数据统计",
		},
	}

	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err = BotAPI.Request(cmdConfig)
	if err != nil {
		log.Printf("注册命令菜单失败: %v", err)
	} else {
		log.Println("Bot命令菜单注册成功")
	}

	return nil
}

func Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := BotAPI.GetUpdatesChan(u)

	// 使用工作池处理更新，提高并发能力
	workerCount := 10
	for i := 0; i < workerCount; i++ {
		go func() {
			for update := range updates {
				handleUpdate(update)
			}
		}()
	}

	// 保持主goroutine运行
	select {}
}

func handleUpdate(update tgbotapi.Update) {
	// 处理命令
	if update.Message != nil {
		handleCommand(update.Message)
		return
	}

	// 处理回调查询
	if update.CallbackQuery != nil {
		handleCallback(update.CallbackQuery)
		return
	}
}
