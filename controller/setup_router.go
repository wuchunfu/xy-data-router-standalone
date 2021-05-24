package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/middleware"
)

func setupRouter(app *fiber.App) *fiber.App {
	v1 := app.Group("/v1/:apiname", middleware.WebAPILogger())
	{
		v1.Get("", V1APIHandler)
		v1.Post("", V1APIHandler)
		v1.Post("/gzip", V1APIHandler)
		v1.Post("/bulk", V1APIHandler)
		v1.Post("/bulk/gzip", V1APIHandler)
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

	// 服务器状态
	app.Get("/sys/status", runningStatusHandler)
	app.Get("/sys/status/queue", runningQueueStatusHandler)
	app.Get("/sys/check", middleware.CheckESWhiteList(false), func(c *fiber.Ctx) error {
		return c.SendString(c.IP() + " - " + c.Get("x-forwarded-for"))
	})

	return app
}
