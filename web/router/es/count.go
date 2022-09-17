package es

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES 统计数量接口
//
// POST /es/count HTTP/1.1
// Content-Type: application/json
// Content-Length: 25
//
// {
//    "index": "test"
// }
func countHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" {
		return response.APIFailure(c, "必填参数: index")
	}

	resp := getResponse()
	resp.response, resp.err = common.ES.Count(
		common.ES.Count.WithContext(context.Background()),
		common.ES.Count.WithIndex(params.Index),
	)
	resp.totalPath = "count"

	return sendResult(c, resp, params)
}
