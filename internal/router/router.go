package router

import (
	"aisearch/internal/config"
	"aisearch/internal/middleware"
	"net/http"

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
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	_ = r.Group("/api/v1")
	{
	}

	return r
}
