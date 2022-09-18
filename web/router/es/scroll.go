package es

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

// ES Scroll 接口
//
// POST /es/search HTTP/1.1
// Content-Type: application/json
// Content-Length: 127
//
// {
//    "index": "test",
//    "scroll": 10,
//    "body": {
//        "query": {
//            "match_all": {}
//        }
//    }
// }
//
// POST /es/scroll HTTP/1.1
// Content-Type: application/json
// Content-Length: 214
//
// {
//    "index": "test",
//    "scroll": 10,
//    "scroll_id": "DnF1ZXJ5VGhlbkZldGNoAwAAAABLNpHAFlVZdG10NF9JVFFXVVJQQmNXbm40YkEAAAAASzaRvxZVWXRtdDRfSVRRV1VSUEJjV25uNGJBAAAAADJlmXcWbTlraG1MV1hTYUthaVVlS3pubjVMQQ=="
// }
func scrollHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Scroll <= 0 || params.ScrollID == "" {
		return response.APIFailure(c, "必填参数: scroll, scroll_id")
	}

	resp := es.GetResponse()
	defer es.PutResponse(resp)
	resp.Response, resp.Err = es.Client.Scroll(
		es.Client.Scroll.WithContext(context.Background()),
		es.Client.Scroll.WithScroll(time.Duration(params.Scroll)*time.Second),
		es.Client.Scroll.WithScrollID(params.ScrollID),
	)

	return sendResult(c, resp, params)
}
