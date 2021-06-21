package middleware

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// Web 日志
func WebAPILogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 错误日志
		if len(c.Errors) > 0 {
			common.LogSampled.Error().
				Strs("errors", c.Errors.Errors()).
				Str("client_ip", c.ClientIP()).Str("uri", c.Request.RequestURI).
				Msg(c.Request.Method)
		}

		costTime := time.Since(start)
		if costTime > conf.Config.SYSConf.WebSlowRespDuration ||
			c.Writer.Status() >= conf.Config.SYSConf.WebErrCodeLog {
			// 记录慢响应日志或错误响应日志
			common.LogSampled.Warn().
				Str("client_ip", c.ClientIP()).Dur("duration", costTime).
				Str("uri", c.Request.RequestURI).Int("http_code", c.Writer.Status()).
				Msg(c.Request.Method)
		}
	}
}

// GinRecovery 及日志
// Ref: https://github.com/gin-contrib/zap
func RecoverLogger(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					common.LogSampled.Error().
						Str("client_ip", c.ClientIP()).
						Str("path", c.Request.URL.Path).Bytes("request", httpRequest).
						Msgf("Recovery: %s", err)
					// If the connection is dead, we can't write a status to it.
					// nolint: errcheck
					_ = c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					common.LogSampled.Error().
						Str("client_ip", c.ClientIP()).
						Str("path", c.Request.URL.Path).Bytes("request", httpRequest).Bytes("stack", debug.Stack()).
						Msgf("Recovery: %s", err)
				} else {
					common.LogSampled.Error().
						Str("client_ip", c.ClientIP()).
						Str("path", c.Request.URL.Path).Bytes("request", httpRequest).
						Msgf("Recovery: %s", err)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
