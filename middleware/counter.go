package middleware

import (
	"sync/atomic"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service"
)

// 请求简单计数
func HTTPCounter() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		atomic.AddUint64(&service.HTTPRequestCounters, 1)
		return c.Next()
	}
}
