package datarouter

import (
	"fmt"

	"github.com/fufuok/utils"
	"github.com/panjf2000/gnet/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/service/schema"
)

type tUDPServerG struct {
	// 是否应答
	withSendTo bool

	gnet.BuiltinEventEngine
}

func udpServerG(addr string, withSendTo bool) error {
	return gnet.Run(
		&tUDPServerG{withSendTo: withSendTo},
		fmt.Sprintf("udp://%s", addr),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
	)
}

func (s *tUDPServerG) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ip, _, err := utils.GetIPPort(c.RemoteAddr())
	if err != nil {
		return
	}
	clientIP := ip.String()

	buf, _ := c.Next(-1)
	n := len(buf)
	if s.withSendTo || n < jsonMinLen {
		// echo 服务
		out := outBytes
		if n == 2 {
			// 返回客户端 IP
			out = utils.S2B(clientIP)
		}
		_ = common.GoPool.Submit(func() {
			_ = c.AsyncWrite(out, nil)
		})
	}

	if n >= jsonMinLen {
		item := schema.NewSafeBody("", clientIP, buf)
		_ = common.GoPool.Submit(func() {
			if !saveUDPData(item) {
				item.Release()
			}
		})
	}
	return
}
