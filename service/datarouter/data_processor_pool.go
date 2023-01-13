package datarouter

import (
	"sync"
	"time"

	"github.com/fufuok/ants"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 数据处理池
var dpPool = sync.Pool{
	New: func() any {
		return new(processor)
	},
}

func newDataPorcessor(dr *router, data *schema.DataItem) *processor {
	dp := dpPool.Get().(*processor)
	dp.dr = dr
	dp.data = data
	return dp
}

func releaseDataProcessor(dp *processor) {
	dp.dr = nil
	dp.data.Release()
	dpPool.Put(dp)
}

// 数据处理协程池初始化
func initDataProcessorPool() {
	DataProcessorPool, _ = ants.NewPoolWithFunc(
		conf.Config.DataConf.ProcessorSize,
		func(i any) {
			dataProcessor(i.(*processor))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.DataConf.ProcessorWaitingLimit),
		ants.WithLogger(common.NewAppLogger()),
		ants.WithPanicHandler(func(r any) {
			common.LogSampled.Error().Msgf("Recovery dataProcessor: %s", r)
		}),
	)
}
