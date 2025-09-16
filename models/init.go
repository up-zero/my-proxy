package models

import (
	"github.com/glebarez/sqlite"
	"github.com/up-zero/my-proxy/util"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var DB *gorm.DB

func NewGormDB() {
	// 获取数据库文件路径
	dbPath, err := util.GetDbPath()
	if err != nil {
		panic(err)
	}
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}
	// 数据库迁移
	err = db.AutoMigrate(&ProxyBasic{}, &UserBasic{}, &ConfigBasic{})
	if err != nil {
		panic(err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(130)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	// 初始化数据
	initData()
}
