package es

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
)

// ES Scroll 接口
func scrollHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Scroll <= 0 || params.ScrollID == "" {
		return middleware.APIFailure(c, "查询参数有误")
	}
	params.ClientIP = c.IP()

	resp := getResponse()
	resp.response, resp.err = common.ES.Scroll(
		common.ES.Scroll.WithContext(context.Background()),
		common.ES.Scroll.WithScroll(time.Duration(params.Scroll)*time.Second),
		common.ES.Scroll.WithScrollID(params.ScrollID),
	)

	ret := parseESResponse(resp, params)
	defer putResult(ret)

	if ret.errMsg != "" {
		return middleware.APIFailure(c, ret.errMsg)
	}

	return middleware.APISuccessBytes(c, ret.result, ret.count)
}
