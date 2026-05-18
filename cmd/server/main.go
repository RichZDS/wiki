package main

import (
	"fmt"
	"os"

	"aisearch/internal/config"
	"aisearch/internal/router"
	"aisearch/pkg/logger"
)

func main() {
	if len(os.Args) > 1 {
		env := os.Args[1]
		os.Setenv("APP_ENV", env)
	}

	config.LoadEnvFile()
	cfg := config.Load()

	logger.Init(cfg.Env)
	log := logger.GetLogger()

	// 启动前检查 MySQL 和 Redis 连接
	cfg.PreStartCheck()

	r := router.New(cfg)
	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf("[%s] server started on http://localhost%s (env=%s, log=%s)",
		cfg.AppName, addr, cfg.Env, cfg.LogLevel)
	log.Printf("database: %s:%s/%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
