package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// ES 接口白名单检查
func CheckESWhiteList(asAPI bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(conf.ESWhiteListConfig) > 0 {
			clientIP, _ := c.RemoteIP()
			forbidden := true
			for ipNet := range conf.ESWhiteListConfig {
				if ipNet.Contains(clientIP) {
					forbidden = false
					break
				}
			}

			if forbidden {
				msg := "非法来访: " + clientIP.String()
				common.LogSampled.Warn().Msg(msg)
				if asAPI {
					APIFailure(c, msg)
				} else {
					TxtMsg(c, msg)
				}

				return
			}
		}

		c.Next()
	}
}
