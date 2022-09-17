package es

import (
	"context"
	"time"

	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES 通用查询接口
//
// POST /es/search HTTP/1.1
// Content-Type: application/json
// Content-Length: 108
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
		return response.APIFailure(c, "必填参数: index, body[查询语句JSON]")
	}

	bodyBuf := bufferpool.Get()
	defer bufferpool.Put(bodyBuf)

	_ = json.NewEncoder(bodyBuf).Encode(params.Body)

	resp := getResponse()
	resp.response, resp.err = common.ES.Search(
		common.ES.Search.WithContext(context.Background()),
		common.ES.Search.WithTrackTotalHits(true),
		common.ES.Search.WithScroll(time.Duration(params.Scroll)*time.Second),
		common.ES.Search.WithIndex(params.Index),
		common.ES.Search.WithBody(bodyBuf),
	)

	return sendResult(c, resp, params)
}
