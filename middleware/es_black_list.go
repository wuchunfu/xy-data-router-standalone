package middleware

import (
	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// CheckESBlackList ES 数据上报接口黑名单检查
func CheckESBlackList(asAPI bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(conf.ESBlackListConfig) > 0 && utils.InIPNetString(c.IP(), conf.ESBlackListConfig) {
			msg := "非法访问: " + c.IP()
			common.LogSampled.Info().Str("method", c.Method()).Str("uri", c.OriginalURL()).Msg(msg)
			if asAPI {
				return APIFailure(c, msg)
			} else {
				return TxtMsg(c, msg)
			}
		}

		return c.Next()
	}
}
