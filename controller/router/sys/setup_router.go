package sys

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
)

// SetupRouter 服务器状态
func SetupRouter(app *fiber.App) {
	sys := app.Group("/sys")
	{
		sys.Get("/status", runningStatusHandler)
		sys.Get("/status/queue", runningQueueStatusHandler)
		sys.Get("/check", middleware.CheckESWhiteList(false), func(c *fiber.Ctx) error {
			return c.SendString(common.GetClientIP(c) + " - " + c.IP())
		})
	}
}
