package tunnel

import (
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/log"

	"github.com/fufuok/xy-data-router/conf"
)

const tunMethod = "tunnel"

func InitTunnel() {
	arpc.DefaultHandler.SetSendQueueSize(conf.Config.SYSConf.TunSendQueueSize)
	arpc.DefaultHandler.SetLogTag("[Tunnel CLI]")
	log.SetLogger(&logger{})

	go initTunServer()
	go initTunClient()
}
