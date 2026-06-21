package router

import (
	"net/http"
	"os"
	"path/filepath"

	"wiki/internal/config"
	"wiki/internal/controller"
	"wiki/internal/middleware"
	"wiki/internal/model"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
)

// New 创建并初始化对应的实例。
func New(cfg config.Config, ragSvc *model.RAGService, jobManager *model.JobManager) *gin.Engine {
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
	ragCtl := controller.NewRAGController(ragSvc)
	jobCtl := controller.NewJobController(jobManager)

	api := r.Group("/api/v1")
	{
		api.POST("/users", userCtl.Create)
		api.GET("/users", userCtl.List)
		api.GET("/users/:id", userCtl.Get)
		api.PUT("/users/:id", userCtl.Update)
		api.DELETE("/users/:id", userCtl.Delete)

		api.POST("/chunk", chunkCtl.Chunk)
		api.POST("/chunk/compare", chunkCtl.Compare)

		api.POST("/rag/ingest", ragCtl.Ingest)
		api.POST("/rag/search", ragCtl.Search)

		api.GET("/jobs", jobCtl.List)
		api.GET("/jobs/:name", jobCtl.Get)
		api.POST("/jobs/:name/start", jobCtl.Start)
		api.POST("/jobs/:name/stop", jobCtl.Stop)
		api.POST("/jobs/:name/run", jobCtl.RunNow)
		api.GET("/jobs/:name/logs", jobCtl.Logs)
	}

	// 生产环境下托管前端静态文件（SPA fallback）
	if publicDir := resolvePublicDir(); publicDir != "" {
		r.Static("/assets", filepath.Join(publicDir, "assets"))

		indexPath := filepath.Join(publicDir, "index.html")
		r.NoRoute(func(c *gin.Context) {
			// 如果是 API 路径，让 Gin 默认 404 处理
			if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "err": "not found"})
				return
			}
			c.File(indexPath)
		})
	}

	return r
}

// resolvePublicDir 查找前端静态文件目录，不存在则返回空字符串。
func resolvePublicDir() string {
	candidates := []string{"public", "../public", "../../public"}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}
