package controller

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"

	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/sjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

// V1APIHandler 处理接口请求
func V1APIHandler(c *fiber.Ctx) error {
	// Some values returned from *fiber.Ctx are not immutable by default
	apiname := utils.CopyString(c.Params("apiname"))

	// 检查接口配置
	apiConf, ok := conf.APIConfig[apiname]
	if !ok || apiConf.APIName == "" {
		common.LogSampled.Error().
			Str("client_ip", c.IP()).Str("uri", c.OriginalURL()).Int("len", len(apiname)).
			Msg("api not found")
		return middleware.APIFailure(c, "接口配置有误")
	}

	// 按场景获取数据
	var body []byte
	chkField := true
	if c.Method() == "GET" {
		// GET 单条数据
		body = query2JSON(c)
		if len(body) == 0 {
			return middleware.APIFailure(c, "数据为空")
		}
	} else {
		body = utils.CopyBytes(c.Body())
		if len(body) == 0 {
			return middleware.APIFailure(c, "数据为空")
		}

		uri := strings.TrimRight(c.Path(), "/")
		if strings.HasSuffix(uri, "/gzip") {
			// 请求体解压缩
			uri = uri[:len(uri)-5]
			unRaw, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				return middleware.APIFailure(c, "数据解压失败")
			}
			body, err = ioutil.ReadAll(unRaw)
			if err != nil {
				return middleware.APIFailure(c, "数据读取失败")
			}
		}

		if strings.HasSuffix(uri, "/bulk") {
			// 批量数据不检查必有字段
			chkField = false
		}
	}

	if chkField {
		// 检查必有字段
		if !common.CheckRequiredField(body, apiConf.RequiredField) {
			return middleware.APIFailure(c, utils.AddString("必填字段: ", strings.Join(apiConf.RequiredField, ",")))
		}
	}

	// 请求 IP
	ip := utils.GetSafeString(c.IP(), common.IPv4Zero)

	// 写入队列
	_ = common.Pool.Submit(func() {
		service.PushDataToChanx(apiname, ip, body)
	})

	return middleware.APISuccessNil(c)
}

// 获取所有 GET 请求参数
func query2JSON(c *fiber.Ctx) (body []byte) {
	c.Request().URI().QueryArgs().VisitAll(func(key []byte, val []byte) {
		body, _ = sjson.SetBytes(body, utils.B2S(key), val)
	})

	return
}
