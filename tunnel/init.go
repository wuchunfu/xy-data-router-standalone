package tunnel

import (
	"github.com/lesismal/arpc"
	"github.com/lesismal/arpc/log"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	tunMethod = "tunnel"
)

func init() {
	arpc.DefaultHandler.SetSendQueueSize(conf.Config.SYSConf.TunSendQueueSize)
	log.SetLogger(&logger{})
}
