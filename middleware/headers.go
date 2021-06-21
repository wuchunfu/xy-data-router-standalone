package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/conf"
)

// 响应头
func ResponseHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Server", conf.WebAPPName)
		c.Next()
	}
}
