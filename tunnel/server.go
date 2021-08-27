package tunnel

import (
	"log"
	"net"

	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service"
)

// InitTunServer 数据通道(RPC)服务初始化
func InitTunServer() {
	common.Log.Info().Str("addr", conf.Config.SYSConf.TunServerAddr).Msg("Listening and serving Tunnel")
	if err := newTunServer(); err != nil {
		log.Fatalln("Failed to start Tunnel Server:", err, "\nbye.")
	}
}

// 新建通道(RPC)服务
func newTunServer() error {
	ln, err := net.Listen("tcp", conf.Config.SYSConf.TunServerAddr)
	if err != nil {
		return err
	}

	svr := arpc.NewServer()
	svr.Codec = &genCodec{}
	svr.Handler.SetLogTag("[Tunnel SVR]")
	svr.Handler.Handle(tunMethod, onData)
	if err = svr.Serve(ln); err != nil {
		return err
	}

	return nil
}

// 处理通道数据
func onData(c *arpc.Context) {
	d := &common.GenDataItem{}
	if err := c.Bind(d); err != nil {
		service.TunRecvBadCount.Inc()
		return
	}

	// 写入队列
	_ = common.Pool.Submit(func() {
		service.TunRecvCount.Inc()
		if d.APIName != "" {
			service.PushDataToChanx(d.APIName, d.IP, &d.Body)
		}
	})
}
