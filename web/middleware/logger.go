package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// WebAPILogger Web 日志
func WebAPILogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Handle request, store err for logging
		chainErr := c.Next()

		// Manually call error handler
		if chainErr != nil {
			common.LogSampled.Error().Err(chainErr).
				Bytes("body", c.Body()).
				Str("client_ip", common.GetClientIP(c)).Str("uri", c.OriginalURL()).
				Msg(c.Method())
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		costTime := time.Since(start)
		if costTime > conf.Config.WebConf.SlowResponseDuration ||
			c.Response().StatusCode() >= conf.Config.WebConf.ErrCodeLog {
			// 记录慢响应日志或错误响应日志
			common.LogSampled.Warn().
				Bytes("body", c.Body()).
				Str("client_ip", common.GetClientIP(c)).Dur("duration", costTime).
				Str("uri", c.OriginalURL()).Int("http_code", c.Response().StatusCode()).
				Msg(c.Method())
		}

		return nil
	}
}

// RecoverLogger Recover 并记录日志
func RecoverLogger() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		// Catch panics
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				if err, ok = r.(error); !ok {
					// Set error that will call the global error handler
					err = fmt.Errorf("%v", r)
				}
				common.LogSampled.Error().Err(err).
					Bytes("body", c.Body()).
					Str("client_ip", common.GetClientIP(c)).Str("uri", c.OriginalURL()).
					Msg(c.Method())
			}
		}()

		return c.Next()
	}
}
