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

	// 记录意外: https://github.com/gofiber/fiber/issues/1388
	app.All("/", func(c *fiber.Ctx) error {
		originalUrl := c.OriginalURL()
		if originalUrl != "/" && c.Method() == "POST" {
			common.LogSampled.Error().
				Str("client_ip", c.IP()).
				Str("method", c.Method()).
				Str("apiname", c.Params("apiname")).
				Str("path", c.Path()).
				Str("base_url", c.BaseURL()).
				Str("original_url", originalUrl).
				Str("ctx_path", string(c.Context().Path())).
				Str("ctx_request_uri", string(c.Context().RequestURI())).
				Str("ctx_uri", c.Context().URI().String()).
				Str("ctx_uri_fulluri", string(c.Context().URI().FullURI())).
				Str("ctx_uri_fulluri", string(c.Context().URI().PathOriginal())).
				Str("request_uri", c.Request().URI().String()).
				Str("request_uri_path", string(c.Request().URI().Path())).
				Str("request_uri_request_uri", string(c.Request().URI().RequestURI())).
				Str("request_uri_fulluri", string(c.Request().URI().FullURI())).
				Str("request_uri_path_original", string(c.Request().URI().PathOriginal())).
				Bytes("body", c.Body()).
				Msg("at / ???")
		}

		return c.SendStatus(fiber.StatusNotFound)
	})

	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})
}
