package main

import (
	"dragon-alert-bot/alert"
	"dragon-alert-bot/bot"
	"dragon-alert-bot/config"
	"dragon-alert-bot/db"
	"dragon-alert-bot/dragon"
	"dragon-alert-bot/lottery"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("长龙提醒机器人启动中...")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	cfg := config.Load()
	log.Println("✓ 配置加载完成")

	// 初始化数据库
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()
	log.Println("✓ 数据库初始化完成")

	// 初始化 Bot
	if err := bot.InitBot(cfg); err != nil {
		log.Fatalf("Bot 初始化失败: %v", err)
	}
	log.Println("✓ Bot 初始化完成")

	// 创建模块
	monitor := lottery.NewMonitor()
	analyzer := dragon.NewAnalyzer(monitor)
	tracker := dragon.NewTracker()
	dispatcher := alert.NewDispatcher(analyzer, tracker)

	// 设置新数据回调
	monitor.OnNewData = func(data *lottery.LotteryData) {
		attrs := data.CalculateAttributes()
		log.Printf("[新开奖] 期号:%s 开奖:%s 和值:%d %s%s", data.Qihao, data.OpenNum, data.SumValue, attrs.Size, attrs.Parity)

		// 构建当前开奖信息
		currentInfo := &dragon.CurrentLotteryInfo{
			Qihao:    data.Qihao,
			OpenNum:  data.OpenNum,
			SumValue: data.SumValue,
			Size:     attrs.Size,
			Parity:   attrs.Parity,
		}

		// 分析长龙
		results := analyzer.Analyze(data)

		// 获取所有启用的群组
		chatIDs, err := analyzer.GetActiveChats()
		if err != nil {
			return
		}

		if len(chatIDs) == 0 {
			return
		}

		// 为每个群组并发处理（提高多群组处理效率）
		alertCount := 0
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, chatID := range chatIDs {
			// 跳过私聊（chatID > 0 为私聊）
			if chatID > 0 {
				continue
			}

			wg.Add(1)
			go func(cid int64) {
				defer wg.Done()

				// 获取群组规则
				rules, err := analyzer.GetChatRules(cid)
				if err != nil {
					return
				}

				if len(rules) == 0 {
					return
				}

				// 根据规则过滤结果
				filteredResults := analyzer.FilterResultsByRules(results, rules)

				if len(filteredResults) > 0 {
					log.Printf("[长龙提醒] 群组:%d 匹配:%d个长龙", cid, len(filteredResults))
					dispatcher.ProcessNewData(cid, filteredResults, currentInfo)

					mu.Lock()
					alertCount++
					mu.Unlock()
				}
			}(chatID)
		}

		wg.Wait()

		if alertCount == 0 && len(results) > 0 {
			log.Printf("[长龙检测] 发现%d个长龙但未达到任何群组阈值", len(results))
		}
	}

	// 启动监测（在 goroutine 中）
	go monitor.Start()
	log.Println("✓ 开奖监测启动")

	// 启动 Bot（在 goroutine 中）
	go bot.Start()
	log.Println("✓ Bot 消息处理启动")

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ 系统运行中 (每秒检测开奖数据)")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\n收到退出信号，正在关闭...")
	log.Println("再见！")
}
