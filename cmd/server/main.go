package main

import (
	"fmt"
	"log"

	"aisearch/internal/config"
	"aisearch/internal/router"
)

func main() {
	cfg := config.Load()

	r := router.New(cfg)
	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf("server started on http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
