package main

import (
	"fmt"
	"os"

	"aisearch/internal/config"
	"aisearch/internal/router"
	"aisearch/pkg/database"
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

	// 初始化 GORM MySQL（连接失败会 panic）
	database.InitMySQL(cfg.DB)
	defer database.Close()

	r := router.New(cfg)
	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf("server started on http://localhost%s (env=%s)", addr, cfg.Env)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
