package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 兼容旧接口
func oldAPIHandler(delKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := c.GetRawData()
		if len(body) == 0 {
			c.String(http.StatusOK, "0")
			return
		}

		// 接口名
		apiname := strings.Trim(strings.Replace(c.Request.URL.String(), "/bulk", "", 1), "/")

		// 接口配置检查
		apiConf, ok := conf.APIConfig[apiname]
		if !ok || apiConf.APIName == "" {
			common.LogSampled.Error().Str("uri", c.Request.RequestURI).Int("len", len(apiname)).Msg("api not found")
			c.String(http.StatusOK, "0")
			return
		}

		// 必有字段校验
		if !common.CheckRequiredField(&body, apiConf.RequiredField) {
			c.String(http.StatusOK, "0")
			return
		}

		// 删除可能非法中文编码的字段
		for _, k := range delKeys {
			body, _ = sjson.DeleteBytes(body, k)
		}

		// 请求 IP
		ip := utils.GetString(c.ClientIP(), common.IPv4Zero)

		// 写入队列
		_ = common.Pool.Submit(func() {
			service.PushDataToChanx(apiname, ip, &body)
		})

		// 旧接口返回值处理
		c.String(http.StatusOK, "1")
	}
}
