package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, status int, data any) {
	c.JSON(status, Body{
		Code:    status,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Body{
		Code:    status,
		Message: message,
	})
}
