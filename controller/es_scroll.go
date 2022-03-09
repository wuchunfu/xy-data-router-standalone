package controller

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
)

// ESScrollHandler ES Scroll 接口
func ESScrollHandler(c *fiber.Ctx) error {
	esScroll := new(tESSearch)
	if err := c.BodyParser(esScroll); err != nil || esScroll.Scroll == 0 || esScroll.ScrollID == "" {
		return middleware.APIFailure(c, "查询参数有误")
	}

	resp, err := common.ES.Scroll(
		common.ES.Scroll.WithContext(context.Background()),
		common.ES.Scroll.WithScroll(time.Duration(esScroll.Scroll)*time.Second),
		common.ES.Scroll.WithScrollID(esScroll.ScrollID),
	)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es scroll, getting response")
		return middleware.APIFailure(c, "查询失败, 服务繁忙")
	}

	if resp.Body == nil {
		common.LogSampled.Error().Err(err).Msg("es scroll, resp.Body is nil")
		return middleware.APIFailure(c, "查询失败, 服务异常")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	esScroll.ClientIP = c.IP()
	res, count, msg := parseESSearch(resp, esScroll)
	if msg != "" {
		return middleware.APIFailure(c, msg)
	}

	return middleware.APISuccessBytes(c, res, count)
}
