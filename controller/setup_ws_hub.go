package controller

import (
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/service"
)

func setupWsHub(app *fiber.App) {
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/hub", websocket.New(func(c *websocket.Conn) {
		var (
			msg []byte
			err error
		)
		for {
			if _, msg, err = c.ReadMessage(); err != nil {
				common.LogSampled.Warn().Err(err).
					Bool("unexpected_close", websocket.IsUnexpectedCloseError(err)).
					Msg("wshub read")
				break
			}

			// 数据解码
			var d common.GenDataItem
			if _, err := d.Unmarshal(msg); err == nil {
				// 写入队列
				_ = common.Pool.Submit(func() {
					atomic.AddUint64(&service.WsHubRequestCounters, 1)
					service.PushDataToChanx(d.APIName, d.IP, &d.Body)
				})
			}
		}
	}))
}
