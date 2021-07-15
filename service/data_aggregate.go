package service

import (
	"sync/atomic"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// 接收数据推入队列
func PushDataToChanx(apiname, ip string, body *[]byte) {
	if conf.ForwardTunnel != "" {
		// 发送数据到 Tun
		TunChan.In <- &common.GenDataItem{
			APIName: apiname,
			IP:      ip,
			Body:    *body,
		}
		atomic.AddUint64(&TunDataTotal, 1)

		return
	}

	dr, ok := dataRouters.Get(apiname)
	if !ok {
		common.LogSampled.Error().Str("apiname", apiname).Int("len", len(apiname)).Msg("nonexistence")
		return
	}
	dr.(*tDataRouter).drChan.In <- newDataItem(apiname, ip, body)
}
