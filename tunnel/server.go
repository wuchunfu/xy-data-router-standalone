package tunnel

import (
	"log"
	"net"

	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
	"github.com/fufuok/xy-data-router/service"
)

// 数据通道(RPC)服务初始化
func initTunServer() {
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

	srv := arpc.NewServer()
	srv.Codec = &genCodec{}
	srv.Handler.SetLogTag("[Tunnel SRV]")
	srv.Handler.Handle(tunMethod, onData)
	common.Log.Info().
		Str("addr", conf.Config.SYSConf.TunServerAddr).
		Int("send_queue_size", srv.Handler.SendQueueSize()).
		Int("send_buffer_size", srv.Handler.SendBufferSize()).
		Int("recv_buffer_size", srv.Handler.RecvBufferSize()).
		Msg("Listening and serving Tunnel")
	if err = srv.Serve(ln); err != nil {
		return err
	}

	return nil
}

// 处理通道数据
func onData(c *arpc.Context) {
	item := schema.Make()
	if err := c.Bind(item); err != nil || item.APIName == "" {
		common.LogSampled.Warn().
			Err(err).Str("apiname", item.APIName).Str("client_ip", item.IP).
			Str("remote_addr", c.Client.Conn.RemoteAddr().String()).
			Msg("TunRecvBad")
		service.TunRecvBadCount.Inc()
		item.Release()
		return
	}

	// 写入队列
	_ = common.Pool.Submit(func() {
		service.TunRecvCount.Inc()
		service.PushDataToChanx(item)
	})
}
