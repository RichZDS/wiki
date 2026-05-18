package database

import (
	"fmt"
	"log"
	"time"

	"aisearch/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMySQL 使用 GORM 初始化 MySQL 连接并自动迁移
func InitMySQL(cfg config.DBConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic(fmt.Errorf("[GORM] 连接 MySQL 失败: %w", err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Errorf("[GORM] 获取底层 sql.DB 失败: %w", err))
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	log.Printf("[GORM] MySQL 连接成功: %s:%s/%s (max_open=%d, max_idle=%d)",
		cfg.Host, cfg.Port, cfg.DBName, cfg.MaxOpenConns, cfg.MaxIdleConns)
}

// Close 关闭数据库连接
func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
