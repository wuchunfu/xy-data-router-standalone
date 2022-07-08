package tunnel

import (
	"log"
	"net"

	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 数据通道(RPC)服务初始化
func initTunServer() {
	if err := newTunServer(); err != nil {
		log.Fatalln("Failed to start Tunnel Server:", err, "\nbye.")
	}
}

// 新建通道(RPC)服务
func newTunServer() error {
	ln, err := net.Listen("tcp", conf.Config.TunConf.ServerAddr)
	if err != nil {
		return err
	}

	srv := arpc.NewServer()
	srv.Codec = &genCodec{}
	srv.Handler.SetLogTag("[Tunnel SRV" + logType + "]")
	srv.Handler.Handle(tunMethod, onData)
	common.Log.Info().
		Str("addr", conf.Config.TunConf.ServerAddr).
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
// handler 函数的结束不代表 Context 的结束, 是可以在 handler 之外异步执行后再调用 ctx.Write 的
// 可以设置使用异步接收, 并自定义异步执行器, SetAsyncExecutor, 默认 go util.Safe(f)
// 也可以使用默认的同步接收, 把 onData 中所有操作都起协程处理: func onData(c *arpc.Context) { go func() { c.Bind... } }
func onData(c *arpc.Context) {
	item := schema.Make()
	if err := c.Bind(item); err != nil || item.APIName == "" {
		common.LogSampled.Warn().
			Err(err).Str("apiname", item.APIName).Str("client_ip", item.IP).
			Str("remote_addr", c.Client.Conn.RemoteAddr().String()).
			Msg("TunRecvBad")
		TunRecvBadCount.Inc()
		item.Release()
		return
	}

	// 写入队列
	TunRecvCount.Inc()
	schema.PushDataToChanx(item)
}
