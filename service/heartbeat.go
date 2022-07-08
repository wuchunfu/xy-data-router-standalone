package service

import (
	"time"

	"github.com/fufuok/utils/jsongen"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 心跳日志
func initHeartbeat() {
	ticker := common.TWm.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		js := jsongen.NewMap()
		js.PutString("type", conf.APPName)
		js.PutString("version", conf.Version)
		js.PutString("internal_ipv4", common.InternalIPv4)
		js.PutString("external_ipv4", common.ExternalIPv4)
		js.PutString("time", time.Now().Format(time.RFC3339))
		data := js.Serialize(nil)
		item := schema.New(conf.Config.SYSConf.HeartbeatIndex, common.ExternalIPv4, data)
		schema.PushDataToChanx(item)
	}
}
