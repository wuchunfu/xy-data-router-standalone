package datarouter

import (
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 数据处理池
var dpPool = sync.Pool{
	New: func() any {
		return new(tDataProcessor)
	},
}

func newDataPorcessor(dr *tDataRouter, data *schema.DataItem) *tDataProcessor {
	dp := dpPool.Get().(*tDataProcessor)
	dp.dr = dr
	dp.data = data
	return dp
}

func releaseDataProcessor(dp *tDataProcessor) {
	dp.dr = nil
	dp.data.Release()
	dpPool.Put(dp)
	DataProcessorTodoCount.Dec()
}

// 数据处理协程池初始化
func initDataProcessorPool() {
	DataProcessorPool, _ = ants.NewPoolWithFunc(
		conf.Config.DataConf.ProcessorSize,
		func(i any) {
			dataProcessor(i.(*tDataProcessor))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.DataConf.ProcessorMaxWorkerSize),
		ants.WithLogger(common.NewAppLogger()),
		ants.WithPanicHandler(func(r any) {
			common.LogSampled.Error().Msgf("Recovery dataProcessor: %s", r)
		}),
	)
}
