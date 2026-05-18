package database

import (
	"fmt"
	"time"

	"aisearch/internal/config"
	"aisearch/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMySQL 使用 GORM 初始化 MySQL，连接失败会 panic。
// cfg: 数据库连接池配置
func InitMySQL(cfg config.DBConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		panic(fmt.Errorf("【启动检查失败】MySQL 连接不可用: %w\n程序无法启动，请检查数据库配置和网络连接", err))
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Errorf("【启动检查失败】MySQL 获取底层连接失败: %w", err))
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	logger.GetLogger().Printf("MySQL 连接成功 %s:%s/%s (max_open=%d max_idle=%d)",
		cfg.Host, cfg.Port, cfg.DBName, cfg.MaxOpenConns, cfg.MaxIdleConns)
}

// Close 关闭数据库连接。
func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
