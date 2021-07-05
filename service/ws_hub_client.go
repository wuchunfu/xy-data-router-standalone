package service

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/fasthttp/websocket"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	wsHubClient *websocket.Conn
)

// WsHub 客户端
func initWsHubClient() {
	if conf.ForwardWsHub == "" {
		return
	}

	common.Log.Info().Str("addr", conf.ForwardWsHub).Msg("Start WsHub Client")
	if err := wsHubDial(); err != nil {
		// 重试 3 次
		for i := 1; i < 4; i++ {
			if err = wsHubDial(); err == nil {
				break
			}
			common.Log.Error().Err(err).Msg("dial wshub")
			time.Sleep(time.Duration(i) * 10 * time.Second)
		}
		if err != nil {
			log.Fatalln("Failed to start WsHub Client:", err, "\nbye.")
		}
	}

	ctx, cancel := context.WithCancel(common.CtxBG)
	go WsHubKeepalive(ctx)

	defer func() {
		cancel()
		close(wsHubChan.In)
		_ = wsHubClient.Close()
	}()

	// 接收数据
	for item := range wsHubChan.Out {
		data := item.(common.GenDataItem)
		// 编码并提交数据到 WsHub
		if b, err := data.Marshal(nil); err == nil {
			err := wsHubClient.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				common.LogSampled.Info().Err(err).Msg("write to wshub")
			}
		}
	}
}

// 保持 WsHub 连接
func WsHubKeepalive(ctx context.Context) {
	ticker := common.TWs.NewTicker(conf.WsHubHeartbeat)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case <-ctx.Done():
			common.Log.Warn().Msg("wshub client exited")
			return
		default:
		}

		// 检测连接状态
		if err := wsHubClient.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
			// 断线重连
			if err := wsHubDial(); err != nil {
				common.Log.Error().Err(err).Msg("wshub client redail")
			}
		}
	}
}

// 连接 WsHub
func wsHubDial() error {
	u := url.URL{Scheme: "ws", Host: conf.ForwardWsHub, Path: "/ws/hub"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	wsHubClient = conn

	return nil
}
