package service

import (
	"github.com/fufuok/xy-data-router/common"
)

// 接收数据推入队列
func PushDataToChanx(apiname, ip string, body *[]byte) {
	dr, ok := dataRouters.Get(apiname)
	if !ok {
		common.LogSampled.Error().Str("apiname", apiname).Strs("dr_keys", dataRouters.Keys()).Msg("nonexistence")
		return
	}
	dr.(*tDataRouter).drChan.In <- newDataItem(apiname, ip, body)
}
