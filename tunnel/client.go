package tunnel

import (
	"fmt"
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

	clients := newTunClients()
	common.Log.Info().
		Str("addr", conf.ForwardTunnel).
		Int("client_num", conf.Config.SYSConf.TunClientNum).
		Int("send_queue_size", arpc.DefaultHandler.SendQueueSize()).
		Int("send_buffer_size", arpc.DefaultHandler.SendBufferSize()).
		Int("recv_buffer_size", arpc.DefaultHandler.RecvBufferSize()).
		Msg("Start Tunnel Client")

	// 支持创建多个 client, 每 client 支持多协程并发处理数据
	for i := range clients {
		client := clients[i]
		common.Log.Debug().Msgf("tunnel client[%d] is working: %p", i, &client.Conn)
		go func() {
			defer client.Stop()
			// 接收数据转发到通道
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
		}()
	}
}

// 支持创建多个 client, 每 client 支持多协程并发处理数据
func newTunClients() (clients []*arpc.Client) {
	for i := 0; i < conf.Config.SYSConf.TunClientNum; i++ {
		handler := arpc.DefaultHandler.Clone()
		handler.SetLogTag(fmt.Sprintf("[Tunnel CLI-%d%s]", i, logType))

		client, err := arpc.NewClient(dialer, handler)
		if err != nil {
			log.Fatalln("Failed to start Tunnel Client:", err, "\nbye.")
		}

		client.Codec = &genCodec{}
		client.Handler.HandleOverstock(onOverstock)
		clients = append(clients, client)

		common.Log.Debug().Msgf("new tunnel client: %p", &client.Conn)
	}
	return
}

func dialer() (net.Conn, error) {
	return net.DialTimeout("tcp", conf.ForwardTunnel, conf.TunDialTimeout)
}

// 不应出现该情况, 线路不畅? 关闭连接, 强制重连
func onOverstock(c *arpc.Client, _ *arpc.Message) {
	_ = c.Conn.Close()
}
