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
		os.Setenv("APP_ENV", os.Args[1])
	}

	cfg := config.Load()

	logger.Init(cfg.Env)
	log := logger.GetLogger()

	database.InitMySQL(cfg.MySQL)
	defer database.Close()

	database.InitRedis(cfg.Redis)
	defer database.CloseRedis()

	r := router.New(cfg)
	addr := fmt.Sprintf(":%s", cfg.Server.Port)

	log.Printf("server started on http://localhost%s (env=%s)", addr, cfg.Env)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
