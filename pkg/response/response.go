package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Err  string `json:"err"`
}

func Success(c *gin.Context, code int, data any) {
	c.JSON(code, Body{
		Code: code,
		Data: data,
	})
}

func Error(c *gin.Context, code int, err string) {
	c.JSON(code, Body{
		Code: code,
		Err:  err,
	})
}
