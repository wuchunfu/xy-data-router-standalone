package service

import (
	"fmt"
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

// 心跳日志
func initHeartbeat() {
	ticker := common.TWm.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		data := utils.S2B(fmt.Sprintf(`{"type":"%s","version":"%s","internal_ipv4":"%s","external_ipv4":"%s"}`,
			conf.APPName, conf.Version, InternalIPv4, ExternalIPv4))
		item := schema.New(conf.Config.SYSConf.HeartbeatIndex, ExternalIPv4, data)
		PushDataToChanx(item)
	}
}
