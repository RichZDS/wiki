package router

import (
	"aisearch/internal/config"
	"aisearch/internal/middleware"
	"aisearch/pkg/response"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, 200, gin.H{"message": "OK"})
	})

	return r
}
