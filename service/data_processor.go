package service

import (
	"bytes"
	"sync"
	"time"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

var (
	// 为真时, 接口配置了该选项的数据将不会写入 ES
	esOptionalWrite bool

	// 数据处理池
	dpPool = sync.Pool{
		New: func() interface{} {
			return new(tDataProcessor)
		},
	}
)

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
		conf.Config.DataConf.ProcessorSize,
		func(i interface{}) {
			dataProcessor(i.(*tDataProcessor))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.DataConf.ProcessorMaxWorkerSize),
		ants.WithPanicHandler(func(r interface{}) {
			common.LogSampled.Error().Interface("recover", r).Msg("panic")
		}),
	)
}

// ES 繁忙时禁止可选写入的状态初始化
func initESOptionalWrite() {
	ticker := common.TWms.NewTicker(conf.UpdateESOptionalInterval)
	defer ticker.Stop()

	for range ticker.C {
		// ES 批量写入排队数 > 10 且 > 最大排队数 * 0.5
		n := esBulkTodoCount.Value()
		m := int64(float64(conf.Config.DataConf.ESBulkMaxWorkerSize) * conf.Config.DataConf.ESBusyPercent)
		esOptionalWrite = n > int64(conf.ESBulkMinWorkerSize) && n > m
	}
}

// 数据处理和分发
// 格式化每个 JSON 数据, 附加系统字段, 发送给 ES 和 API 队列, 释放 DataItem
func dataProcessor(dp *tDataProcessor) {
	defer releaseDataProcessor(dp)

	isPostToES := !(conf.Config.DataConf.ESDisableWrite || esOptionalWrite && dp.dr.apiConf.ESOptionalWrite)
	isPostToAPI := dp.dr.apiConf.PostAPI.Interval > 0
	if !isPostToES {
		// 丢弃可选写入 ES 数据项计数
		esDataItemDiscards.Inc()
		if !isPostToAPI {
			return
		}
	}

	// 兼容 {body} 或 {body}=-:-=[{body},{body}]
	dp.data.Body = pretty.UglyInPlace(dp.data.Body)
	for _, js := range bytes.Split(dp.data.Body, esBodySep) {
		if len(js) < jsonMinLen {
			continue
		}

		if !gjson.ValidBytes(js) {
			common.LogSampled.Info().
				Bytes("body", js).Str("apiname", dp.data.APIName).Str("client_ip", dp.data.IP).
				Msg("Invalid JSON")
			continue
		}

		switch js[0] {
		case '[': // 字典列表 [{body},{body}]
			gjson.Result{Type: gjson.JSON, Raw: utils.B2S(js)}.ForEach(func(_, v gjson.Result) bool {
				if v.IsObject() {
					sendOneData(dp, utils.S2B(v.String()), isPostToES, isPostToAPI)
				}
				return true
			})
		case '{': // 批量文本数据 {body}=-:-={body}
			sendOneData(dp, js, isPostToES, isPostToAPI)
		}
	}
}

func sendOneData(dp *tDataProcessor, js []byte, isPostToES, isPostToAPI bool) {
	jsData := appendSYSField(js, dp.data.IP)
	if jsData == nil {
		return
	}
	if isPostToES {
		esData := bytespool.Make(dp.dr.apiConf.ESBulkHeaderLength + len(jsData) + 1)
		esData = append(esData, dp.dr.apiConf.ESBulkHeader...)
		esData = append(esData, jsData...)
		esData = append(esData, '\n')
		// 需要在 ES 使用后回收 DataItem
		item := schema.Make()
		item.Body = esData
		dp.dr.drOut.esChan.In <- item
	}
	if isPostToAPI {
		// 需要在 API 使用后回收 DataItem
		item := schema.Make()
		item.Body = jsData
		dp.dr.drOut.apiChan.In <- item
	} else {
		bytespool.Release(jsData)
	}
}

// 附加系统字段, 已存在 _cip 字段时不重复附加, Immutable
// 返回值使用完成后可以回收, 如:
// jsData := appendSYSField([]byte(`{"f":"f"}`, "1.1.1.1")
// bytespool.Release(jsData)
func appendSYSField(js []byte, ip string) []byte {
	n := len(js)
	if n < jsonMinLen {
		return nil
	}

	exist := gjson.GetBytes(js, "_cip").Exists()
	if !exist {
		// 加系统字段 JSON 长度
		n += len(ip) + 79
	}

	buf := bytespool.Make(n)

	i := 0
	if !exist {
		buf = append(buf, `{"_cip":"`...)
		buf = append(buf, ip...)
		buf = append(buf, `","_ctime":"`...)
		buf = append(buf, common.Now3399UTC...)
		buf = append(buf, `","_gtime":"`...)
		buf = append(buf, common.Now3399...)
		buf = append(buf, `",`...)
		i = 1
	}

	buf = append(buf, js[i:]...)

	return buf
}
