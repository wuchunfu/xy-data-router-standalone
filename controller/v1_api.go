package controller

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"

	"github.com/fufuok/utils"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

// 处理接口请求
func V1APIHandler(c *gin.Context) {
	// 检查接口配置
	apiname := c.Param("apiname")
	apiConf, ok := conf.APIConfig[apiname]
	if !ok || apiConf.APIName == "" {
		common.LogSampled.Error().Str("uri", c.Request.RequestURI).Int("len", len(apiname)).Msg("api not found")
		middleware.APIFailure(c, "接口配置有误")
		return
	}

	// 按场景获取数据
	var body []byte
	chkField := true
	if c.Request.Method == "GET" {
		// GET 单条数据
		body = query2JSON(c)
	} else {
		body, _ = c.GetRawData()
		if len(body) == 0 {
			middleware.APIFailure(c, "数据为空")
			return
		}

		uri := c.Request.URL.String()
		if strings.HasSuffix(uri, "/gzip") {
			// 请求体解压缩
			uri = uri[:len(uri)-5]
			unRaw, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				middleware.APIFailure(c, "数据解压失败")
				return
			}
			body, err = ioutil.ReadAll(unRaw)
			if err != nil {
				middleware.APIFailure(c, "数据读取失败")
				return
			}
		}

		if strings.HasSuffix(uri, "/bulk") {
			// 批量数据不检查必有字段
			chkField = false
		}
	}

	if chkField {
		// 检查必有字段
		if !common.CheckRequiredField(&body, apiConf.RequiredField) {
			middleware.APIFailure(c, utils.AddString("必填字段: ", strings.Join(apiConf.RequiredField, ",")))
			return
		}
	}

	// 请求 IP
	ip := utils.GetString(c.ClientIP(), common.IPv4Zero)

	// 写入队列
	_ = common.Pool.Submit(func() {
		service.PushDataToChanx(apiname, ip, &body)
	})

	middleware.APISuccessNil(c)
}

// 获取所有 GET 请求参数
func query2JSON(c *gin.Context) (body []byte) {
	for k, v := range c.Request.URL.Query() {
		body, _ = sjson.SetBytes(body, k, utils.AddStringBytes(v...))
	}

	return
}
