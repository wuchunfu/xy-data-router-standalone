package middleware

import (
	"github.com/gin-gonic/gin"
)

// 基于请求 IP 限流
func IPLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// pass
		c.Next()
	}
}
