package middleware

import (
	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/web/response"
)

// CheckESWhiteList ES 查询接口白名单检查
func CheckESWhiteList(asAPI bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(conf.ESWhiteListConfig) > 0 {
			clientIP := common.GetClientIP(c)
			if !utils.InIPNetString(clientIP, conf.ESWhiteListConfig) {
				msg := "非法来访"
				sampler.Info().
					Str("cip", c.IP()).Str("x_forwarded_for", c.Get(fiber.HeaderXForwardedFor)).
					Str(common.HeaderXProxyClientIP, c.Get(common.HeaderXProxyClientIP)).
					Str("method", c.Method()).Str("uri", c.OriginalURL()).Str("client_ip", clientIP).
					Msg(msg)
				if asAPI {
					return response.APIFailure(c, msg, clientIP)
				} else {
					return response.TxtMsg(c, msg)
				}
			}
		}
		return c.Next()
	}
}
