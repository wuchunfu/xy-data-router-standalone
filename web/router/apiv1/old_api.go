package apiv1

import (
	"strings"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xjson/sjson"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 兼容旧接口
func oldAPIHandler(delKeys []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(c.Body()) == 0 {
			return c.SendString("0")
		}

		// 接口名
		apiname := utils.Trim(strings.Replace(c.Path(), "/bulk", "", 1), '/')

		// 接口配置检查
		apiConf, ok := conf.APIConfig[apiname]
		if !ok || apiConf.APIName == "" {
			sampler.Info().Str("uri", c.OriginalURL()).Int("len", len(apiname)).Msg("api not found")
			return c.SendString("0")
		}

		// 必有字段校验
		body := c.Body()
		if !common.CheckRequiredField(body, apiConf.RequiredField) {
			return c.SendString("0")
		}

		// 删除可能非法中文编码的字段
		for _, k := range delKeys {
			body, _ = sjson.DeleteBytes(body, k)
		}

		// 请求 IP
		ip := common.GetClientIP(c)

		// 写入队列
		item := schema.NewSafe(apiname, ip, body)
		schema.PushDataToChanx(item)

		// 旧接口返回值处理
		return c.SendString("1")
	}
}
