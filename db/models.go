package db

import (
	"time"
)

// ChatConfig 群组配置
type ChatConfig struct {
	ChatID    int64     `db:"chat_id"`
	Enabled   bool      `db:"enabled"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// DragonRule 长龙规则配置
type DragonRule struct {
	ID            int64     `db:"id"`
	ChatID        int64     `db:"chat_id"`
	PatternType   string    `db:"pattern_type"`   // a, ab, abb, ab_ac, ab_cd
	AttributeType string    `db:"attribute_type"` // size, parity, sum, size_parity
	Threshold     int       `db:"threshold"`
	Enabled       bool      `db:"enabled"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// DragonAlert 长龙提醒记录
type DragonAlert struct {
	ID             int64     `db:"id"`
	ChatID         int64     `db:"chat_id"`
	PatternType    string    `db:"pattern_type"`
	AttributeType  string    `db:"attribute_type"`
	StartQihao     string    `db:"start_qihao"`
	CurrentQihao   string    `db:"current_qihao"`
	Count          int       `db:"count"`
	PatternDetail  string    `db:"pattern_detail"`
	LastAlertCount int       `db:"last_alert_count"`
	Status         string    `db:"status"` // active, ended
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// LotteryCheckState 数据检查状态
type LotteryCheckState struct {
	ID            int       `db:"id"`
	LastQihao     string    `db:"last_qihao"`
	LastCheckTime time.Time `db:"last_check_time"`
}


