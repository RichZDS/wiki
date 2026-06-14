package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wiki/internal/config"
	"wiki/internal/job"
	"wiki/internal/router"
	"wiki/pkg/database"
	"wiki/pkg/logger"
)

// main 初始化应用依赖并启动 HTTP 服务。
func main() {
	if len(os.Args) > 1 {
		os.Setenv("APP_ENV", os.Args[1])
	}

	cfg := config.Load()

	logger.Init(cfg.Env)
	logger := logger.GetLogger()

	database.InitMySQL(cfg.MySQL)
	defer database.Close()

	database.InitRedis(cfg.Redis)
	defer database.CloseRedis()

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()

	jobManager := job.NewManager()
	jobGroup := job.NewDefaultJobGroup(database.DB)
	if err := jobGroup.RegisterAll(jobManager); err != nil {
		logger.Fatalf("register jobs: %v", err)
	}
	jobManager.Start(ctx)

	r := router.New(cfg)
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	serverErrors := make(chan error, 1)

	logger.Printf("server started on http://localhost%s (env=%s)", addr, cfg.Env)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-signalCtx.Done():
		logger.Printf("shutdown signal received")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("server stopped: %v", err)
		}
	}

	cancel()
	jobManager.Wait()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("server shutdown failed: %v", err)
	}
}
