package datarouter

import (
	"time"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
)

// initDataRouter 根据接口配置初始化数据分发处理器
func initDataRouter() {
	// 关闭配置中已取消的接口
	dataRouters.Range(func(apiname string, dr *router) bool {
		if _, ok := conf.APIConfig[apiname]; !ok {
			dataRouters.Delete(apiname)
			close(dr.drChan.In)
		}
		return true
	})

	// 按接口创建数据分发处理器
	ymd := common.GTimeNowString("060102")
	for apiname, cfg := range conf.APIConfig {
		apiConf := cfg
		apiConf.ESBulkHeader = getESBulkHeader(apiConf, ymd)
		apiConf.ESBulkHeaderLength = len(apiConf.ESBulkHeader)
		if conf.Debug {
			logger.Info().RawJSON(apiname, apiConf.ESBulkHeader).Msg("ESBulkHeader")
		}
		dr, ok := dataRouters.Load(apiname)
		if ok {
			// 更新接口配置
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
func dataRouter(dr *router) {
	logger.Info().Str("apiname", dr.apiConf.APIName).Msg("Start DataRouter worker")

	// 开启接口对应 API 推送处理器
	go apiWorker(dr)

	// 接收数据
	for item := range dr.drChan.Out {
		// 提交不阻塞, 有执行并发限制, 最大排队数限制
		dp := newDataPorcessor(dr, item)
		_ = common.GoPool.Submit(func() {
			if err := DataProcessorPool.Invoke(dp); err != nil {
				releaseDataProcessor(dp)
				DataProcessorDiscards.Inc()
				sampler.Warn().Err(err).Msg("go dataProcessor discards")
			}
		})
	}

	// 准备退出
	time.Sleep(2 * time.Second)
	close(dr.apiChan.In)
	logger.Warn().Str("apiname", dr.apiConf.APIName).Msg("DataRouter worker exited")
}
