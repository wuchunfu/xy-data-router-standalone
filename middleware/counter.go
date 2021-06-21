package middleware

import (
	"sync/atomic"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/service"
)

// 请求简单计数
func HTTPCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		atomic.AddUint64(&service.HTTPRequestCounters, 1)
		c.Next()
	}
}
