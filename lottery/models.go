package lottery

import (
	"time"
)

// LotteryData 开奖数据
type LotteryData struct {
	Qihao     string    `db:"qihao"`
	OpenTime  time.Time `db:"opentime"`
	OpenNum   string    `db:"opennum"`
	SumValue  int       `db:"sum_value"`
	Source    string    `db:"source"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Attributes 开奖属性
type Attributes struct {
	Qihao    string
	Size     string // 大/小
	Parity   string // 单/双
	SumValue int    // 和值
}

// CalculateAttributes 计算属性
func (ld *LotteryData) CalculateAttributes() Attributes {
	// <14为小，≥14为大
	size := "大"
	if ld.SumValue < 14 {
		size = "小"
	}

	parity := "双"
	if ld.SumValue%2 != 0 {
		parity = "单"
	}

	return Attributes{
		Qihao:    ld.Qihao,
		Size:     size,
		Parity:   parity,
		SumValue: ld.SumValue,
	}
}
