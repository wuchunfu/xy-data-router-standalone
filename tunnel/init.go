package tunnel

import (
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/log"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

const tunMethod = "tunnel"

func InitTunnel() {
	log.SetLogger(newLogger())
	arpc.SetSendQueueSize(conf.Config.TunConf.SendQueueSize)
	arpc.EnablePool(true)
	arpc.SetAsyncResponse(true)
	arpc.SetAsyncExecutor(func(f func()) {
		_ = common.GoPool.Submit(f)
	})

	go initTunServer()
	go initTunClient()
}
