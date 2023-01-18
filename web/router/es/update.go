package es

import (
	"context"

	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

type updateBody struct {
	Doc map[string]any `json:"doc"`
}

// ES 更新文档接口
//
// {
//    "index": "test",
//    "document_id": "1",
//    "refresh": "true",
//    "body": {
//        "counter": 1,
//        "tags": [
//            "red"
//        ]
//    }
// }
func updateHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" || params.DocumentID == "" || params.Body == nil {
		return response.APIFailure(c, "缺失必填参数", "index, document_id, body(更新内容JSON)")
	}

	bodyBuf := bufferpool.Get()
	defer bufferpool.Put(bodyBuf)

	// 为更新内容包裹 doc 字段
	_ = json.NewEncoder(bodyBuf).Encode(updateBody{Doc: params.Body})

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Update(params.Index, params.DocumentID, bodyBuf,
		es.Client.Update.WithContext(context.Background()),
		es.Client.Update.WithRefresh(fixedRefresh(params.Refresh)),
	)

	return sendResult(c, resp, params)
}
