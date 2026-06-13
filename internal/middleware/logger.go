package middleware

import (
	"time"

	"wiki/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RequestLogger 负责处理当前模块中的对应业务逻辑。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.GetLogger().Printf(
			"%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
