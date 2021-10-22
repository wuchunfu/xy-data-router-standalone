package service

import (
	"net"

	"github.com/fufuok/utils"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/goroutine"
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
		body := utils.CopyBytes(frame)
		clientIP, _, err := net.SplitHostPort(c.RemoteAddr().String())
		if err == nil {
			_ = s.pool.Submit(func() {
				saveUDPData(body, clientIP)
			})
		}
	}

	return
}
