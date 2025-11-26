package alert

import (
	"dragon-alert-bot/bot"
	"dragon-alert-bot/dragon"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Dispatcher struct {
	analyzer *dragon.Analyzer
	tracker  *dragon.Tracker
}

func NewDispatcher(analyzer *dragon.Analyzer, tracker *dragon.Tracker) *Dispatcher {
	return &Dispatcher{
		analyzer: analyzer,
		tracker:  tracker,
	}
}

// ProcessNewData 处理新开奖数据
func (d *Dispatcher) ProcessNewData(chatID int64, results []*dragon.PatternResult, currentData *dragon.CurrentLotteryInfo) {
	if len(results) == 0 {
		return
	}

	// 跟踪所有长龙并收集需要提醒的
	var alertResults []*dragon.PatternResult

	for _, result := range results {
		shouldAlert, _ := d.tracker.TrackDragon(chatID, result)
		if shouldAlert {
			alertResults = append(alertResults, result)
		}
	}

	// 结束不活跃的长龙
	d.tracker.EndInactiveDragons(chatID, results)

	// 如果有需要提醒的，发送消息
	if len(alertResults) > 0 {
		d.sendAlert(chatID, alertResults, currentData)
	}
}

func (d *Dispatcher) sendAlert(chatID int64, results []*dragon.PatternResult, currentData *dragon.CurrentLotteryInfo) {
	message := bot.FormatAlertMessage(results, currentData)
	if message == "" {
		return
	}

	// 异步发送消息，避免阻塞
	go func(cid int64, msg string) {
		msgConfig := tgbotapi.NewMessage(cid, msg)
		msgConfig.ParseMode = "HTML"
		msgConfig.DisableWebPagePreview = true

		_, err := bot.BotAPI.Send(msgConfig)
		if err != nil {
			log.Printf("[发送失败] 群组:%d 错误:%v", cid, err)
		}
	}(chatID, message)
}
