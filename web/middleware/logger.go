package middleware

import (
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
)

// WebAPILogger Web 日志
func WebAPILogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Handle request, store err for logging
		chainErr := c.Next()

		// Manually call error handler
		if chainErr != nil {
			sampler.Error().Err(chainErr).
				Bytes("body", c.Body()).
				Str("client_ip", common.GetClientIP(c)).Str("method", c.Method()).
				Msg(c.OriginalURL())
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		costTime := time.Since(start)
		if costTime > conf.Config.WebConf.SlowResponseDuration ||
			c.Response().StatusCode() >= conf.Config.WebConf.ErrCodeLog {
			// 记录慢响应日志或错误响应日志
			sampler.Warn().
				Bytes("body", c.Body()).
				Str("client_ip", common.GetClientIP(c)).Dur("duration", costTime).
				Str("method", c.Method()).Int("http_code", c.Response().StatusCode()).
				Msg(c.OriginalURL())
		}

		return nil
	}
}

// RecoverLogger Recover 并记录日志
func RecoverLogger() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				if err, ok = r.(*fiber.Error); !ok {
					// 屏蔽错误细节, 让全局错误处理响应 500
					err = &fiber.Error{
						Code:    500,
						Message: "Internal Server Error",
					}
				}
				sampler.Error().
					Bytes("stack", debug.Stack()).
					Bytes("body", c.Body()).
					Str("client_ip", c.IP()).
					Str("method", c.Method()).Str("uri", c.OriginalURL()).
					Msgf("Recovery: %s", r)
			}
		}()
		return c.Next()
	}
}
