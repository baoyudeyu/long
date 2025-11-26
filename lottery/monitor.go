package lottery

import (
	"dragon-alert-bot/db"
	"time"
)

type Monitor struct {
	OnNewData func(data *LotteryData)
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) Start() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.checkNewData()
	}
}

func (m *Monitor) checkNewData() {
	// 获取当前最新期号
	var latestQihao string
	err := db.ReadDB.QueryRow("SELECT qihao FROM latest_lottery_data ORDER BY opentime DESC LIMIT 1").Scan(&latestQihao)
	if err != nil {
		return
	}

	// 获取上次检查的期号
	var lastQihao string
	err = db.WriteDB.QueryRow("SELECT last_qihao FROM lottery_check_state WHERE id = 1").Scan(&lastQihao)
	if err != nil {
		return
	}

	// 如果有新数据
	if latestQihao != lastQihao && latestQihao != "" {
		// 获取最新开奖数据
		data, err := m.getLatestData()
		if err != nil {
			return
		}

		// 更新检查状态
		_, err = db.WriteDB.Exec("UPDATE lottery_check_state SET last_qihao = ?, last_check_time = ? WHERE id = 1",
			latestQihao, time.Now())
		if err != nil {
			return
		}

		// 触发回调
		if m.OnNewData != nil {
			m.OnNewData(data)
		}
	}
}

func (m *Monitor) getLatestData() (*LotteryData, error) {
	var data LotteryData
	var openTimeStr, createdAtStr, updatedAtStr string

	err := db.ReadDB.QueryRow(`
		SELECT qihao, opentime, opennum, sum_value, source, created_at, updated_at 
		FROM latest_lottery_data 
		ORDER BY opentime DESC 
		LIMIT 1
	`).Scan(&data.Qihao, &openTimeStr, &data.OpenNum, &data.SumValue, &data.Source, &createdAtStr, &updatedAtStr)

	if err != nil {
		return nil, err
	}

	// 解析时间字符串
	data.OpenTime, _ = time.Parse("2006-01-02 15:04:05", openTimeStr)
	data.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
	data.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

	return &data, nil
}

// GetHistoryData 获取历史数据（用于长龙分析）
func (m *Monitor) GetHistoryData(limit int) ([]LotteryData, error) {
	rows, err := db.ReadDB.Query(`
		SELECT qihao, opentime, opennum, sum_value, source, created_at, updated_at 
		FROM latest_lottery_data 
		ORDER BY opentime DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataList []LotteryData
	for rows.Next() {
		var data LotteryData
		var openTimeStr, createdAtStr, updatedAtStr string

		err := rows.Scan(&data.Qihao, &openTimeStr, &data.OpenNum, &data.SumValue, &data.Source, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, err
		}

		// 解析时间字符串
		data.OpenTime, _ = time.Parse("2006-01-02 15:04:05", openTimeStr)
		data.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		data.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAtStr)

		dataList = append(dataList, data)
	}

	return dataList, nil
}
