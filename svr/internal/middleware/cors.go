package middleware

import (
	"net/http"
	"strings"

	"wiki/internal/config"

	"github.com/gin-gonic/gin"
)

// CORS 创建跨域资源共享中间件，根据配置中的 allowed_origins 白名单允许前端跨域访问。
// 处理 OPTIONS 预检请求时直接返回 204。
func CORS(cfg config.Config) gin.HandlerFunc {
	allowedOrigins := cfg.Server.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:94001"}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if strings.EqualFold(o, origin) || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
