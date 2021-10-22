package service

import (
	"bytes"
	"time"

	"github.com/fufuok/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// InitDataRouter 根据接口配置初始化数据分发处理器
func InitDataRouter() {
	// 关闭配置中已取消的接口
	dataRouters.Range(func(key string, value interface{}) bool {
		if _, ok := conf.APIConfig[key]; !ok {
			dataRouters.Delete(key)
			close(value.(*tDataRouter).drChan.In)
		}
		return true
	})

	// 按接口创建数据分发处理器
	ymd := common.GetGlobalDataTime("060102")
	for apiname, cfg := range conf.APIConfig {
		apiConf := cfg
		apiConf.ESBulkHeader = getESBulkHeader(apiConf, ymd)
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

// 数据处理协程池初始化
func initDataProcessorPool() {
	dataProcessorPool, _ = ants.NewPoolWithFunc(
		conf.Config.SYSConf.DataProcessorSize,
		func(i interface{}) {
			dataProcessor(i.(*tDataProcessor))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.SYSConf.DataProcessorMaxWorkerSize),
		ants.WithPanicHandler(func(r interface{}) {
			common.LogSampled.Error().Interface("recover", r).Msg("panic")
		}),
	)
}

// 数据分发处理器
func dataRouter(dr *tDataRouter) {
	common.Log.Info().Str("apiname", dr.apiConf.APIName).Msg("Start DataRouter worker")

	// 开启接口对应 API 推送处理器
	go apiWorker(dr)

	// 接收数据
	for item := range dr.drChan.Out {
		// 提交不阻塞, 有执行并发限制, 最大排队数限制
		_ = common.Pool.Submit(func() {
			dataProcessorTodoCount.Inc()
			if err := dataProcessorPool.Invoke(&tDataProcessor{
				dr:   dr,
				data: item.(*tDataItem),
			}); err != nil {
				dataProcessorDiscards.Inc()
				common.LogSampled.Error().Err(err).Msg("go dataProcessor")
			}
		})
	}

	// 准备退出
	time.Sleep(time.Second)
	close(dr.drOut.apiChan.In)
	common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("DataRouter worker exited")
}

// 数据处理和分发
// 格式化每个 JSON 数据, 附加系统字段, 发送给 ES 和 API 队列
func dataProcessor(dp *tDataProcessor) {
	defer dataProcessorTodoCount.Dec()
	isPostToAPI := dp.dr.apiConf.PostAPI.Interval > 0

	// 兼容 {body} 或 {body}=-:-=[{body},{body}]
	for _, js := range bytes.Split(pretty.Ugly(dp.data.body), esBodySep) {
		if len(js) == 0 {
			continue
		}

		if !gjson.ValidBytes(js) {
			common.LogSampled.Warn().Bytes("body", js).Str("apiname", dp.data.apiname).Msg("Invalid JSON")
			continue
		}

		switch js[0] {
		case '[':
			// 字典列表 [{body},{body}]
			gjson.Result{Type: gjson.JSON, Raw: utils.B2S(js)}.ForEach(func(_, v gjson.Result) bool {
				if v.IsObject() {
					body := appendSYSField(utils.S2B(v.String()), dp.data.ip)
					dp.dr.drOut.esChan.In <- utils.JoinBytes(dp.dr.apiConf.ESBulkHeader, body, ln)
					if isPostToAPI {
						dp.dr.drOut.apiChan.In <- body
					}
				}
				return true
			})
		case '{':
			// 批量文本数据 {body}=-:-={body}
			js = appendSYSField(js, dp.data.ip)
			dp.dr.drOut.esChan.In <- utils.JoinBytes(dp.dr.apiConf.ESBulkHeader, js, ln)
			if isPostToAPI {
				dp.dr.drOut.apiChan.In <- js
			}
		}
	}
}
