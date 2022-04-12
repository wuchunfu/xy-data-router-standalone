package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// IPLimiter 基于请求 IP 限流
func IPLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        conf.Config.SYSConf.LimitRequest,
		Expiration: time.Duration(conf.Config.SYSConf.LimitExpiration) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return common.GetClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			common.LogSampled.Error().
				Str("uri", c.OriginalURL()).Str("ip", common.GetClientIP(c)).Strs("ips", c.IPs()).
				Msg("limit reached")
			return c.SendStatus(fiber.StatusTooManyRequests)
		},
	})
}
