package controller

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/middleware"
)

// ES Scroll 接口
func ESScrollHandler(c *gin.Context) {
	esScroll := new(tESSearch)
	if err := c.ShouldBindJSON(&esScroll); err != nil || esScroll.Scroll == 0 || esScroll.ScrollID == "" {
		middleware.APIFailure(c, "查询参数有误")
		return
	}

	resp, err := common.ES.Scroll(
		common.ES.Scroll.WithContext(context.Background()),
		common.ES.Scroll.WithScroll(time.Duration(esScroll.Scroll)*time.Second),
		common.ES.Scroll.WithScrollID(esScroll.ScrollID),
	)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es scroll, getting response")
		middleware.APIFailure(c, "查询失败, 服务繁忙")
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	esScroll.ClientIP = c.ClientIP()
	res, count, msg := parseESSearch(resp, esScroll)
	if msg != "" {
		middleware.APIFailure(c, msg)
		return
	}
	middleware.APISuccess(c, res, count)
}
