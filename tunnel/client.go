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

func initTunClient() {
	if conf.ForwardTunnel == "" {
		return
	}

	client, err := arpc.NewClient(func() (net.Conn, error) {
		return net.DialTimeout("tcp", conf.ForwardTunnel, conf.TunDialTimeout)
	})
	if err != nil {
		log.Fatalln("Failed to start Tunnel Client:", err, "\nbye.")
	}

	defer client.Stop()
	client.Codec = &genCodec{}
	common.Log.Info().
		Str("addr", conf.ForwardTunnel).
		Int("send_queue_size", client.Handler.SendQueueSize()).
		Int("send_buffer_size", client.Handler.SendBufferSize()).
		Int("recv_buffer_size", client.Handler.RecvBufferSize()).
		Msg("Start Tunnel Client")

	// 接收数据转发到通道 (支持创建多个 client, 每 client 支持多协程并发处理数据)
	for item := range service.TunChan.Out {
		data := item.(*schema.DataItem)
		_ = common.Pool.Submit(func() {
			// 不超时, 直到 ErrClientOverstock
			if err := client.Notify(tunMethod, data, arpc.TimeZero); err != nil {
				common.LogSampled.Warn().Err(err).Msg("Failed to write Tunnel")
				service.TunSendErrors.Inc()
				return
			}
			service.TunSendCount.Inc()
		})
	}
}
