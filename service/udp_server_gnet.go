package service

import (
	"github.com/fufuok/utils"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"

	"github.com/fufuok/xy-data-router/schema"
)

type tUDPServerG struct {
	*gnet.EventServer
	pool       *goroutine.Pool
	withSendTo bool
}

func udpServerG(addr string, withSendTo bool) error {
	p := goroutine.Default()
	defer p.Release()
	return gnet.Serve(
		&tUDPServerG{pool: p, withSendTo: withSendTo},
		"udp://"+addr,
		gnet.WithMulticore(true),
		gnet.WithReusePort(true),
	)
}

// React PS: 一次接收的数据上限为: 64K
func (s *tUDPServerG) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	n := len(frame)
	if s.withSendTo || n < 7 {
		out = outBytes
	}
	if n >= 7 {
		clientIP, _, err := utils.GetIPPort(c.RemoteAddr())
		if err != nil {
			return
		}
		item := schema.New("", clientIP.String(), frame)
		_ = s.pool.Submit(func() {
			if !saveUDPData(item) {
				item.Release()
			}
		})
	}

	return
}
