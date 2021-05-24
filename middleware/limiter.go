package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// 基于请求 IP 限流
func IPLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        conf.Config.SYSConf.LimitRequest,
		Expiration: time.Duration(conf.Config.SYSConf.LimitExpiration) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			cip := c.IP()
			if utils.IsInternalIPv4String(cip) {
				if fip := c.Get("x-forwarded-for"); fip != "" {
					return fip
				}
			}

			return cip
		},
		LimitReached: func(c *fiber.Ctx) error {
			common.LogSampled.Error().
				Str("uri", c.OriginalURL()).Str("ip", c.IP()).Strs("ips", c.IPs()).
				Msg("limit reached")
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	})
}
