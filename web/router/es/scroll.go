package es

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
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

	resp := getResponse()
	resp.response, resp.err = common.ES.Scroll(
		common.ES.Scroll.WithContext(context.Background()),
		common.ES.Scroll.WithScroll(time.Duration(params.Scroll)*time.Second),
		common.ES.Scroll.WithScrollID(params.ScrollID),
	)

	return sendResult(c, resp, params)
}
