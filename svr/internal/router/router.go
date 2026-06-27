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

// New 创建并初始化 Gin 引擎，注册所有路由。
func New(cfg config.Config, ragSvc *model.RAGService, jobManager *model.JobManager) *gin.Engine {
	if cfg.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 注入全局依赖到 controller 层。
	controller.SetRAGService(ragSvc)
	controller.SetJobManager(jobManager)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(cfg))

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, 200, gin.H{"message": "OK"})
	})

	api := r.Group("/api/v1")
	{
		api.POST("/users", controller.CreateUser)
		api.GET("/users", controller.ListUsers)
		api.GET("/users/:id", controller.GetUser)
		api.PUT("/users/:id", controller.UpdateUser)
		api.DELETE("/users/:id", controller.DeleteUser)

		api.POST("/chunk", controller.Chunk)
		api.POST("/chunk/compare", controller.CompareChunk)

		api.POST("/rag/ingest", controller.IngestRAG)
		api.POST("/rag/search", controller.SearchRAG)

		api.GET("/jobs", controller.ListJobs)
		api.GET("/jobs/:name", controller.GetJob)
		api.POST("/jobs/:name/start", controller.StartJob)
		api.POST("/jobs/:name/stop", controller.StopJob)
		api.POST("/jobs/:name/run", controller.RunJobNow)
		api.GET("/jobs/:name/logs", controller.ListJobLogs)
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
