package db

import (
	"database/sql"
	"dragon-alert-bot/config"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ReadDB  *sql.DB
	WriteDB *sql.DB
)

func InitDB(cfg *config.Config) error {
	var err error

	// 初始化只读数据库
	ReadDB, err = sql.Open("mysql", cfg.ReadDB.DSN())
	if err != nil {
		return err
	}
	ReadDB.SetMaxOpenConns(50)
	ReadDB.SetMaxIdleConns(25)
	ReadDB.SetConnMaxLifetime(time.Hour)
	ReadDB.SetConnMaxIdleTime(10 * time.Minute)

	if err = ReadDB.Ping(); err != nil {
		return err
	}
	log.Printf("只读数据库连接成功 (%s)", cfg.ReadDB.Database)

	// 初始化读写数据库
	WriteDB, err = sql.Open("mysql", cfg.WriteDB.DSN())
	if err != nil {
		return err
	}
	WriteDB.SetMaxOpenConns(100)
	WriteDB.SetMaxIdleConns(50)
	WriteDB.SetConnMaxLifetime(time.Hour)
	WriteDB.SetConnMaxIdleTime(10 * time.Minute)

	if err = WriteDB.Ping(); err != nil {
		return err
	}
	log.Printf("读写数据库连接成功 (%s)", cfg.WriteDB.Database)

	// 初始化表结构
	if err = InitTables(); err != nil {
		return err
	}

	return nil
}

func Close() {
	if ReadDB != nil {
		ReadDB.Close()
	}
	if WriteDB != nil {
		WriteDB.Close()
	}
}
