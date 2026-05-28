package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"aisearch/internal/config"
	"aisearch/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// gormLogWriter 适配器，将 GORM 日志输出到项目的 log.Logger（同时写入 stdout + 每日滚动文件）。
type gormLogWriter struct{}

func (w *gormLogWriter) Printf(format string, args ...any) {
	logger.GetLogger().Printf(format, args...)
}

func newGormLogger() gormlogger.Interface {
	return gormlogger.New(
		&gormLogWriter{},
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

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

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
