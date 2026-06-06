package response

import (
	"aisearch/internal/model"

	"github.com/gin-gonic/gin"
)

// Success 负责处理当前模块中的对应业务逻辑。
func Success(c *gin.Context, code int, data any) {
	c.JSON(code, model.ResponseBody{
		Code: code,
		Data: data,
	})
}

// Error 负责处理当前模块中的对应业务逻辑。
func Error(c *gin.Context, code int, err string) {
	c.JSON(code, model.ResponseBody{
		Code: code,
		Err:  err,
	})
}
