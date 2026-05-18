package check

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"aisearch/internal/config"
)

// CheckMySQL 检查 MySQL 数据库连接是否正常
// 如果连接失败，会返回 error 信息
// ctx: 上下文，用于控制超时
// cfg: 数据库配置
// 返回值: error 如果连接失败
func CheckMySQL(ctx context.Context, cfg config.DBConfig) error {
	// 构建 DSN (Data Source Name) 连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// 创建数据库连接，设置 5 秒超时
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("[MySQL] 创建数据库连接失败: %w", err)
	}
	defer db.Close()

	// 设置连接池参数
	db.SetConnMaxLifetime(5 * time.Second)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// 尝试 Ping 数据库以验证连接
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("[MySQL] 无法连接到数据库 %s:%s, 错误: %w", cfg.Host, cfg.Port, err)
	}

	return nil
}

// CheckRedis 检查 Redis 服务器连接是否正常
// 如果连接失败，会返回 error 信息
// ctx: 上下文，用于控制超时
// cfg: Redis 配置
// 返回值: error 如果连接失败
func CheckRedis(ctx context.Context, cfg config.RedisConfig) error {
	// 构建 Redis 地址
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	// 创建 Redis 客户端配置
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       0,
	})
	defer rdb.Close()

	// 设置 5 秒超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 尝试 Ping Redis 服务器
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("[Redis] 无法连接到 Redis %s, 错误: %w", addr, err)
	}

	return nil
}

// PreStartCheck 程序启动前的依赖检查
// 会依次检查 MySQL 和 Redis 连接
// 如果任何一项检查失败，程序会 panic 并退出
// cfg: 应用程序配置
func PreStartCheck(cfg *config.Config) {
	ctx := context.Background()

	// 检查 MySQL 连接
	if err := CheckMySQL(ctx, cfg.DB); err != nil {
		panic(fmt.Sprintf("【启动检查失败】MySQL 数据库连接不可用: %v\n程序无法启动，请检查数据库配置和网络连接。", err))
	}

	// 检查 Redis 连接
	if err := CheckRedis(ctx, cfg.Redis); err != nil {
		panic(fmt.Sprintf("【启动检查失败】Redis 缓存连接不可用: %v\n程序无法启动，请检查 Redis 配置和网络连接。", err))
	}
}
