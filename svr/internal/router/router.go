package router

import (
	"wiki/internal/config"
	"wiki/internal/controller"
	"wiki/internal/middleware"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// New 创建并初始化对应的实例。
func New(cfg config.Config) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(cfg))

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, 200, gin.H{"message": "OK"})
	})

	userCtl := controller.NewUserController()
	chunkCtl := controller.NewChunkController()

	api := r.Group("/api/v1")
	{
		api.POST("/users", userCtl.Create)
		api.GET("/users", userCtl.List)
		api.GET("/users/:id", userCtl.Get)
		api.PUT("/users/:id", userCtl.Update)
		api.DELETE("/users/:id", userCtl.Delete)

		api.POST("/chunk", chunkCtl.Chunk)
		api.POST("/chunk/compare", chunkCtl.Compare)
	}

	return r
}
