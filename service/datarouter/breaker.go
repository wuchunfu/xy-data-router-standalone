package datarouter

import (
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xsync"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger/alarm"
)

var (
	// ESOptionalWrite 为真时, 接口配置了该选项的数据将不会写入 ES
	ESOptionalWrite utils.Bool

	// ESDisableWrite 为真时, 关闭所有 ES 写入
	ESDisableWrite utils.Bool

	// ESBreakerCount 熔断计数
	ESBreakerCount xsync.Counter
)

// 同步配置中的 ES 写入状态
func initESWriteStatus() {
	ESDisableWrite.Store(conf.Config.DataConf.ESDisableWrite)
}

// ES 熔断器, 繁忙时禁止可选写入的状态初始化
func initESWriteBreaker() {
	ticker := common.TWms.NewTicker(conf.UpdateESOptionalInterval)
	defer ticker.Stop()

	// 记录熔断参数值, 被丢弃的批量数据计数
	discards := ESBulkDiscards.Value()
	fusing := false

	for range ticker.C {
		// ES 批量写入有数据被丢弃时, 熔断一个检查周期
		bd := ESBulkDiscards.Value()
		if bd > discards {
			ESDisableWrite.StoreTrue()
			discards = bd
			fusing = true
		}

		if fusing {
			// 直到有空闲 Worker 时, 重置 ES 写入状态
			if ESBulkPool.Free() > 0 {
				initESWriteStatus()
				fusing = false
			} else {
				ESBreakerCount.Inc()
				ESDisableWrite.StoreTrue()
			}
		}

		// 部分接口数据写入熔断: ES 批量写入排队数 >= 10 且 > 最大排队数 * 0.5
		n := ESBulkPool.Waiting()
		m := int(float64(conf.Config.DataConf.ESBulkerWaitingLimit) * conf.Config.DataConf.ESBusyPercent)
		ESOptionalWrite.Store(n >= conf.ESBulkerWaitingMin && n > m)
	}
	alarm.Error().Msg("Exception: DataRouter ESWriteBreaker worker exited")
}
