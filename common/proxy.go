package common

import (
	"fmt"

	"github.com/fufuok/utils"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal"
)

const (
	HeaderXProxyClientIP = "X-Proxy-ClientIP"
	HeaderXProxyToken    = "X-Proxy-Token"
	HeaderXProxyTime     = "X-Proxy-Time"
	HeaderXProxyPass     = "X-Proxy-Pass"
)

var (
	// ForwardTunnel 数据转发地址
	ForwardTunnel string

	// ForwardHTTP API 查询接口转发后端服务地址
	ForwardHTTP []string
)

func initProxy() {
	if conf.ForwardHost != "" {
		_, port := utils.SplitHostPort(conf.Config.SYSConf.TunServerAddr)
		ForwardTunnel = fmt.Sprintf("%s:%s", conf.ForwardHost, port)
		_, port = utils.SplitHostPort(conf.Config.SYSConf.WebServerAddr)
		ForwardHTTP = []string{
			fmt.Sprintf("http://%s:%s", conf.ForwardHost, port),
		}
	}
}

// SetClientIP 首个代理加密客户端 IP, 中间代理透传
// 当前来访为内网 IP 会跳过设置
// immutable
func SetClientIP(c *fiber.Ctx) {
	xip := c.Get(HeaderXProxyClientIP)
	if xip == "" {
		xip = c.IP()
		if !utils.IsInternalIPv4String(xip) {
			xtime := Now3399
			xtoken := internal.HashString(xip, xtime, conf.Config.SYSConf.BaseSecretValue)
			c.Request().Header.Set(HeaderXProxyClientIP, xip)
			c.Request().Header.Set(HeaderXProxyToken, xtoken)
			c.Request().Header.Set(HeaderXProxyTime, xtime)
		}
	}
}

// GetClientIP 获取客户端 IP
// 1. 上下文存储中获取
// 2. 下游代理头信息中获取
// 3. TCP 协议 RemoteIP()
// immutable
func GetClientIP(c *fiber.Ctx) string {
	clientIP, _ := c.Locals(HeaderXProxyClientIP).(string)
	if clientIP != "" {
		return clientIP
	}

	xip := c.Get(HeaderXProxyClientIP)
	if xip != "" {
		xtoken := c.Get(HeaderXProxyToken)
		xtime := c.Get(HeaderXProxyTime)
		if xtoken == internal.HashString(xip, xtime, conf.Config.SYSConf.BaseSecretValue) {
			c.Locals(HeaderXProxyClientIP, xip)
			return xip
		}
	}

	clientIP = c.IP()
	if clientIP == "" {
		clientIP = IPv4Zero
	}
	c.Locals(HeaderXProxyClientIP, clientIP)
	return clientIP
}
