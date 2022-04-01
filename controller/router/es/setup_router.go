package es

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/middleware"
)

// SetupRouter ES 相关接口
func SetupRouter(app *fiber.App) {
	es := app.Group("/es", middleware.CheckESWhiteList(true))
	{
		// ES 查询总数
		es.Post("/count", countHandler)
		// ES DSL 通用查询
		es.Post("/search", searchHandler)
		// ES Scroll
		es.Post("/scroll", scrollHandler)
	}
}
