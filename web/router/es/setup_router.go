package es

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/web/middleware"
)

// SetupRouter ES 相关接口
func SetupRouter(app *fiber.App) {
	// 反向代理服务
	if len(common.ForwardHTTP) > 0 {
		app.Use("/es/", proxy.Balancer(proxy.Config{
			Servers: common.ForwardHTTP,
			// * 超时参数重启生效
			Timeout: esProxyTimeout,
			ModifyRequest: func(c *fiber.Ctx) error {
				common.SetClientIP(c)
				return nil
			},
			ModifyResponse: func(c *fiber.Ctx) error {
				c.Response().Header.Set(common.HeaderXProxyPass, conf.ForwardHost)
				return nil
			},
		}))
		return
	}
	es := app.Group("/es", middleware.CheckESWhiteList(true))
	{
		// ES 查询总数
		es.Post("/count", countHandler)
		// ES DSL 通用查询
		es.Post("/search", searchHandler)
		// ES Scroll 滚动查询
		es.Post("/scroll", scrollHandler)
		// ES Update 更新单个文档
		es.Post("/update", updateHandler)
		// ES Delete 删除单个文档
		es.Post("/delete", deleteHandler)
		// ES 集群健康
		es.Get("/health", healthHandler)
	}
}
