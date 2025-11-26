package db

import (
	"log"
)

func InitTables() error {
	tables := []string{
		// 群组配置表
		`CREATE TABLE IF NOT EXISTS chat_configs (
			chat_id BIGINT PRIMARY KEY,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 长龙规则配置表
		`CREATE TABLE IF NOT EXISTS dragon_rules (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			chat_id BIGINT NOT NULL,
			pattern_type VARCHAR(20) NOT NULL,
			attribute_type VARCHAR(20) NOT NULL,
			threshold INT DEFAULT 4,
			enabled BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY unique_rule (chat_id, pattern_type, attribute_type),
			INDEX idx_chat_id (chat_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 长龙提醒记录表
		`CREATE TABLE IF NOT EXISTS dragon_alerts (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			chat_id BIGINT NOT NULL,
			pattern_type VARCHAR(20) NOT NULL,
			attribute_type VARCHAR(20) NOT NULL,
			start_qihao VARCHAR(20) NOT NULL,
			current_qihao VARCHAR(20) NOT NULL,
			count INT NOT NULL,
			pattern_detail TEXT,
			last_alert_count INT DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_chat_status (chat_id, status),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 数据检查状态表
		`CREATE TABLE IF NOT EXISTS lottery_check_state (
			id INT PRIMARY KEY DEFAULT 1,
			last_qihao VARCHAR(20) DEFAULT '',
			last_check_time DATETIME DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, table := range tables {
		if _, err := WriteDB.Exec(table); err != nil {
			return err
		}
	}

	log.Println("数据库表初始化完成")

	// 初始化检查状态
	WriteDB.Exec("INSERT IGNORE INTO lottery_check_state (id, last_qihao) VALUES (1, '')")

	// 清理私聊配置（chatID > 0 为私聊）
	WriteDB.Exec("DELETE FROM chat_configs WHERE chat_id > 0")
	WriteDB.Exec("DELETE FROM dragon_rules WHERE chat_id > 0")
	WriteDB.Exec("DELETE FROM dragon_alerts WHERE chat_id > 0")

	return nil
}
