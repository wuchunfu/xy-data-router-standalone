package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

func setupRouter(app *fiber.App) {
	// 动态接口
	v1 := app.Group("/v1", middleware.WebAPILogger())
	{
		v1.Post("/:apiname/bulk/gzip", V1APIHandler)
		v1.Post("/:apiname/bulk", V1APIHandler)
		v1.Post("/:apiname/gzip", V1APIHandler)
		v1.Post("/:apiname", V1APIHandler)
		v1.Get("/:apiname", V1APIHandler)
	}

	// 兼容旧 ES 上报接口
	oldAPI := []string{"/start/", "/stop/", "/tp2cn/", "/pubg_proxy/bulk/", "/tcp_proxy/bulk/"}
	for _, u := range oldAPI {
		app.Post(u, oldAPIHandler(nil))
	}

	// 测速数据上报 JSON 修正 (临时方案)
	app.Post("/speed_report/", oldAPIHandler([]string{"data.node_line_type"}))

	// ES 相关接口
	es := app.Group("/es", middleware.CheckESWhiteList(true))
	{
		// ES 通用查询
		es.Post("/search", ESSearchHandler)
		// ES Scroll
		es.Post("/scroll", ESScrollHandler)
	}

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
		return c.SendString(service.ExternalIPv4)
	})

	// 服务器状态
	app.Get("/sys/status", runningStatusHandler)
	app.Get("/sys/status/queue", runningQueueStatusHandler)
	app.Get("/sys/check", middleware.CheckESWhiteList(false), func(c *fiber.Ctx) error {
		return c.SendString(c.IP() + " - " + c.Get("x-forwarded-for"))
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
}
