package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
)

// HTTPCounter 请求简单计数
func HTTPCounter() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		common.HTTPRequestCount.Inc()
		return c.Next()
	}
}
