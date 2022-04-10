package service

import (
	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

// PushDataToChanx 接收数据推入队列
func PushDataToChanx(item *schema.DataItem) {
	if common.ForwardTunnel != "" {
		TunDataTotal.Inc()
		// 设置压缩标识
		if item.Size() >= conf.Config.SYSConf.TunCompressMinSize {
			item.Flag = 1
			TunCompressTotal.Inc()
		}
		// 发送数据到 Tun
		TunChan.In <- item
		return
	}

	dr, ok := dataRouters.Load(item.APIName)
	if !ok {
		common.LogSampled.Error().Str("apiname", item.APIName).Int("len", len(item.APIName)).Msg("nonexistence")
		item.Release()
		return
	}
	dr.(*tDataRouter).drChan.In <- item
}
