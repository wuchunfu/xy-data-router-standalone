package apiv1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/web/middleware"
)

func SetupRouter(app *fiber.App) {
	// 动态接口
	v1 := app.Group("/v1", middleware.WebAPILogger())
	{
		v1.Post("/:apiname/bulk/gzip", apiHandler)
		v1.Post("/:apiname/bulk", apiHandler)
		v1.Post("/:apiname/gzip", apiHandler)
		v1.Post("/:apiname", apiHandler)
		v1.Get("/:apiname", apiHandler)
	}

	// 兼容旧 ES 上报接口
	oldAPI := []string{"/start/", "/stop/", "/tp2cn/", "/pubg_proxy/bulk/", "/tcp_proxy/bulk/"}
	for _, u := range oldAPI {
		app.Post(u, oldAPIHandler(nil))
	}

	// 测速数据上报 JSON 修正 (临时方案)
	app.Post("/speed_report/", oldAPIHandler([]string{"data.node_line_type"}))
}
