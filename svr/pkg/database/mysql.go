package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"wiki/internal/config"
	"wiki/internal/model"
	"wiki/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// newGormLogger 创建接入项目日志系统的 GORM 日志器。
func newGormLogger() gormlogger.Interface {
	return gormlogger.New(
		&model.GormLoggerWriter{Logger: logger.GetLogger()},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// InitMySQL 初始化 MySQL 连接池并验证连接可用性。
func InitMySQL(cfg config.MySQLConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newGormLogger(),
	})
	if err != nil {
		// 确保有日志输出（logger 可能尚未初始化）
		log.Printf("【启动检查失败】MySQL 连接不可用: %v", err)
		fmt.Fprintf(os.Stderr, "程序无法启动，请检查数据库配置和网络连接\n")
		os.Exit(1)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Errorf("【启动检查失败】MySQL 获取底层连接失败: %w", err))
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	logger.GetLogger().Printf("MySQL 连接成功 %s:%s/%s (max_open=%d max_idle=%d)",
		cfg.Host, cfg.Port, cfg.Database, cfg.MaxOpenConns, cfg.MaxIdleConns)
}

// Close 关闭 MySQL 底层连接池。
func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
