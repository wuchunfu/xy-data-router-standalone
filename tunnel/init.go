package tunnel

import (
	"github.com/lesismal/arpc"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

const tunMethod = "tunnel"

func InitTunnel() {
	initLogger()
	arpc.SetSendQueueSize(conf.Config.TunConf.SendQueueSize)
	arpc.EnablePool(true)
	arpc.SetAsyncResponse(true)
	arpc.SetAsyncExecutor(func(f func()) {
		_ = common.GoPool.Submit(f)
	})

	go initTunServer()
	go initTunClient()
}
