package service

import (
	"github.com/fufuok/utils"
	"github.com/panjf2000/gnet"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/schema"
)

type tUDPServerG struct {
	// 是否应答
	withSendTo bool

	*gnet.EventServer
}

func udpServerG(addr string, withSendTo bool) error {
	return gnet.Serve(
		&tUDPServerG{withSendTo: withSendTo},
		"udp://"+addr,
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
	)
}

// React PS: 一次接收的数据上限为: 64K
func (s *tUDPServerG) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	ip, _, err := utils.GetIPPort(c.RemoteAddr())
	if err != nil {
		return
	}
	clientIP := ip.String()

	n := len(frame)
	if s.withSendTo || n < 7 {
		out = outBytes
		if n == 2 {
			// 返回客户端 IP
			out = utils.S2B(clientIP)
		}
	}

	if n >= 7 {
		item := schema.NewSafeBody("", clientIP, frame)
		_ = common.Pool.Submit(func() {
			if !saveUDPData(item) {
				item.Release()
			}
		})
	}

	return
}
