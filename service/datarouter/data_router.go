package datarouter

import (
	"time"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/schema"
)

// initDataRouter 根据接口配置初始化数据分发处理器
func initDataRouter() {
	// 关闭配置中已取消的接口
	dataRouters.Range(func(key string, value interface{}) bool {
		if _, ok := conf.APIConfig[key]; !ok {
			dataRouters.Delete(key)
			close(value.(*tDataRouter).drChan.In)
		}
		return true
	})

	// 按接口创建数据分发处理器
	ymd := common.GTimeNowString("060102")
	for apiname, cfg := range conf.APIConfig {
		apiConf := cfg
		apiConf.ESBulkHeader = getESBulkHeader(apiConf, ymd)
		apiConf.ESBulkHeaderLength = len(apiConf.ESBulkHeader)
		v, ok := dataRouters.Load(apiname)
		if ok {
			// 更新接口配置
			dr := v.(*tDataRouter)
			dr.apiConf = apiConf
		} else {
			// 新建数据通道
			dr := newDataRouter(apiConf)
			dataRouters.Store(apiname, dr)

			// 开启数据分发处理器
			go dataRouter(dr)
		}
	}
}

// 数据分发处理器
func dataRouter(dr *tDataRouter) {
	common.Log.Info().Str("apiname", dr.apiConf.APIName).Msg("Start DataRouter worker")

	// 开启接口对应 API 推送处理器
	go apiWorker(dr)

	// 接收数据
	for item := range dr.drChan.Out {
		// 提交不阻塞, 有执行并发限制, 最大排队数限制
		dp := newDataPorcessor(dr, item.(*schema.DataItem))
		_ = common.GoPool.Submit(func() {
			DataProcessorTodoCount.Inc()
			if err := DataProcessorPool.Invoke(dp); err != nil {
				releaseDataProcessor(dp)
				DataProcessorDiscards.Inc()
				common.LogSampled.Error().Err(err).Msg("go dataProcessor")
			}
		})
	}

	// 准备退出
	time.Sleep(2 * time.Second)
	close(dr.apiChan.In)
	common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("DataRouter worker exited")
}
