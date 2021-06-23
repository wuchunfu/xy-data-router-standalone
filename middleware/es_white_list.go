package middleware

import (
	"net"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// ES 接口白名单检查
func CheckESWhiteList(asAPI bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(conf.ESWhiteListConfig) > 0 {
			clientIP := c.IP()
			ip := net.ParseIP(clientIP)
			forbidden := true
			for ipNet := range conf.ESWhiteListConfig {
				if ipNet.Contains(ip) {
					forbidden = false
					break
				}
			}

			if forbidden {
				msg := "非法来访: " + clientIP
				common.LogSampled.Warn().Str("method", c.Method()).Str("uri", c.OriginalURL()).Msg(msg)
				if asAPI {
					return APIFailure(c, msg)
				} else {
					return TxtMsg(c, msg)
				}
			}
		}

		return c.Next()
	}
}
