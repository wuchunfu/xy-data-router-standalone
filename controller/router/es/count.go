package es

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
)

// ES 统计数量接口
func countHandler(c *fiber.Ctx) error {
	params := getParams()
	defer putParams(params)

	if err := c.BodyParser(params); err != nil || params.Index == "" {
		return middleware.APIFailure(c, "查询参数有误")
	}

	resp := getResponse()
	resp.response, resp.err = common.ES.Count(
		common.ES.Count.WithContext(context.Background()),
		common.ES.Count.WithIndex(params.Index),
	)
	resp.totalPath = "count"

	ret := parseESResponse(resp, params)
	defer putResult(ret)

	if ret.errMsg != "" {
		return middleware.APIFailure(c, ret.errMsg)
	}
	return middleware.APISuccessBytes(c, ret.result, ret.count)
}
