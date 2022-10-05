package web

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/web/router/apiv1"
	"github.com/fufuok/xy-data-router/web/router/es"
	"github.com/fufuok/xy-data-router/web/router/sys"
)

func setupRouter(app *fiber.App) {
	// 动态接口
	apiv1.SetupRouter(app)

	// ES 相关接口
	es.SetupRouter(app)

	// 服务器状态
	sys.SetupRouter(app)

	// 健康检查
	app.Get("/heartbeat", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("PONG")
	})

	// 客户端和服务端 IP
	app.Get("/client_ip", func(c *fiber.Ctx) error {
		return c.SendString(c.IP())
	})
	app.Get("/server_ip", func(c *fiber.Ctx) error {
		return c.SendString(common.ExternalIPv4)
	})

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})
}
