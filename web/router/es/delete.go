package es

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES 删除文档接口
//
// {
//    "index": "test",
//    "refresh": "true",
//    "document_id": "1"
// }
func deleteHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" || params.DocumentID == "" {
		return response.APIFailure(c, "缺失必填参数", "index, document_id")
	}

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Delete(params.Index, params.DocumentID,
		es.Client.Delete.WithRefresh(fixedRefresh(params.Refresh)),
	)

	return sendResult(c, resp, params)
}
