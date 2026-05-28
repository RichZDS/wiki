package router

import (
	"aisearch/internal/config"
	"aisearch/internal/controller"
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

	userCtl := controller.NewUserController()

	api := r.Group("/api/v1")
	{
		api.POST("/users", userCtl.Create)
		api.GET("/users", userCtl.List)
		api.GET("/users/:id", userCtl.Get)
		api.PUT("/users/:id", userCtl.Update)
		api.DELETE("/users/:id", userCtl.Delete)
	}

	return r
}
