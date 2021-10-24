package service

import (
	"bytes"
	"sync"
	"time"

	"github.com/fufuok/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	bbPool "github.com/valyala/bytebufferpool"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

var dpPool = sync.Pool{
	New: func() interface{} {
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
	dataProcessorTodoCount.Dec()
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

// 数据处理和分发
// 格式化每个 JSON 数据, 附加系统字段, 发送给 ES 和 API 队列, 释放 DataItem
func dataProcessor(dp *tDataProcessor) {
	defer releaseDataProcessor(dp)

	isPostToES := !conf.Config.SYSConf.ESDisableWrite
	isPostToAPI := dp.dr.apiConf.PostAPI.Interval > 0
	if !isPostToES && !isPostToAPI {
		return
	}

	// 兼容 {body} 或 {body}=-:-=[{body},{body}]
	for _, js := range bytes.Split(pretty.Ugly(dp.data.Body), esBodySep) {
		if len(js) == 0 {
			continue
		}

		if !gjson.ValidBytes(js) {
			common.LogSampled.Warn().
				Bytes("body", js).Str("apiname", dp.data.APIName).Str("client_ip", dp.data.IP).
				Msg("Invalid JSON")
			continue
		}

		switch js[0] {
		case '[':
			// 字典列表 [{body},{body}]
			gjson.Result{Type: gjson.JSON, Raw: utils.B2S(js)}.ForEach(func(_, v gjson.Result) bool {
				if v.IsObject() {
					body := appendSYSField(utils.S2B(v.String()), dp.data.IP)
					if isPostToES {
						esData := bbPool.Get()
						_, _ = esData.Write(dp.dr.apiConf.ESBulkHeader)
						_, _ = esData.Write(body)
						_, _ = esData.Write(ln)
						dp.dr.drOut.esChan.In <- esData
					}
					if isPostToAPI {
						dp.dr.drOut.apiChan.In <- body
					}
				}
				return true
			})
		case '{':
			// 批量文本数据 {body}=-:-={body}
			js = appendSYSField(js, dp.data.IP)
			if isPostToES {
				esData := bbPool.Get()
				_, _ = esData.Write(dp.dr.apiConf.ESBulkHeader)
				_, _ = esData.Write(js)
				_, _ = esData.Write(ln)
				dp.dr.drOut.esChan.In <- esData
			}
			if isPostToAPI {
				dp.dr.drOut.apiChan.In <- js
			}
		}
	}
}
