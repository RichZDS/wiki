package router

import (
	"github.com/gin-gonic/gin"

	"aisearch/internal/config"
	"aisearch/internal/handler"
	"aisearch/internal/middleware"
	"aisearch/internal/repository"
	"aisearch/internal/service"
)

func New(cfg config.Config) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	healthHandler := handler.NewHealthHandler(cfg)

	wikiRepo := repository.NewWikiRepository()
	wikiService := service.NewWikiService(wikiRepo)
	wikiHandler := handler.NewWikiHandler(wikiService)

	r.GET("/health", healthHandler.Check)

	api := r.Group("/api/v1")
	{
		api.GET("/wikis", wikiHandler.List)
		api.GET("/wikis/:id", wikiHandler.GetByID)
	}

	return r
}
