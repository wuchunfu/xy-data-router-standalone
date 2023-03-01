package es

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES 统计数量接口
//
// {
//    "index": "test"
// }
func countHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil {
		return response.APIFailure(c, "参数解析错误", err.Error())
	}
	if params.Index == "" {
		return response.APIFailure(c, "缺失必填参数", "index")
	}

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Count(
		es.Client.Count.WithIndex(params.Index),
	)
	resp.TotalPath = "count"

	return sendResult(c, resp, params)
}
