package es

import (
	"context"
	"time"

	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES 通用查询接口
//
// {
//    "index": "test",
//    "body": {
//        "query": {
//            "match_all": {}
//        }
//    }
// }
func searchHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" || params.Body == nil {
		return response.APIFailure(c, "缺失必填参数", "index, body(查询语句JSON)")
	}

	bodyBuf := bufferpool.Get()
	defer bufferpool.Put(bodyBuf)

	_ = json.NewEncoder(bodyBuf).Encode(params.Body)

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Search(
		es.Client.Search.WithContext(context.Background()),
		es.Client.Search.WithTrackTotalHits(true),
		es.Client.Search.WithScroll(time.Duration(params.Scroll)*time.Second),
		es.Client.Search.WithIndex(params.Index),
		es.Client.Search.WithBody(bodyBuf),
	)

	return sendResult(c, resp, params)
}
