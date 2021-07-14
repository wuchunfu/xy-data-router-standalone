package tunnel

import (
	"log"
	"net"
	"sync/atomic"

	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// 数据通道(RPC)服务初始化
func InitTunServer() {
	common.Log.Info().Str("addr", conf.Config.SYSConf.TunServerAddr).Msg("Listening and serving Tunnel")
	ln, err := net.Listen("tcp", conf.Config.SYSConf.TunServerAddr)
	if err != nil {
		log.Fatalln("Failed to start Tunnel Server:", err, "\nbye.")
	}

	svr := arpc.NewServer()
	svr.Codec = &genCodec{}
	svr.Handler.Handle(tunMethod, onData)
	if err = svr.Serve(ln); err != nil {
		log.Fatalln("Failed to start Tunnel Server:", err, "\nbye.")
	}
}

// 处理通道数据
func onData(c *arpc.Context) {
	d := &common.GenDataItem{}
	if err := c.Bind(d); err != nil {
		atomic.AddUint64(&service.TunRecvBadCounters, 1)
		return
	}

	// 写入队列
	_ = common.Pool.Submit(func() {
		atomic.AddUint64(&service.TunRecvCounters, 1)
		if d.APIName != "" {
			service.PushDataToChanx(d.APIName, d.IP, &d.Body)
		}
	})
}
