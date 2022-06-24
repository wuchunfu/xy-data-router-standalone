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
func searchHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" || params.Body == nil {
		return response.APIFailure(c, "查询参数有误")
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
