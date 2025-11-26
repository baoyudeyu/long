package config

import (
	"fmt"
)

type Config struct {
	// Telegram Bot
	BotToken string

	// 只读数据库 - 开奖数据
	ReadDB DatabaseConfig

	// 读写数据库 - 用户数据
	WriteDB DatabaseConfig

	// 轮询间隔（秒）
	PollInterval int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (dc DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dc.User, dc.Password, dc.Host, dc.Port, dc.Database)
}

func Load() *Config {
	return &Config{
		BotToken: "7861097095:AAEmyA5Rd1xkrhAcblkQZh1I_CxP0ADp-Nk",
		ReadDB: DatabaseConfig{
			Host:     "rm-bp14172o6ehyk82g0vo.mysql.rds.aliyuncs.com",
			Port:     3306,
			User:     "pc28_help",
			Password: "04By0302",
			Database: "pc28_help",
		},
		WriteDB: DatabaseConfig{
			Host:     "rm-bp14172o6ehyk82g0vo.mysql.rds.aliyuncs.com",
			Port:     3306,
			User:     "t3bot",
			Password: "04By0302",
			Database: "t3bot",
		},
		PollInterval: 1,
	}
}


