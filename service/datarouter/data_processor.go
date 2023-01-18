package datarouter

import (
	"bytes"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xjson/gjson"
	"github.com/fufuok/utils/xjson/pretty"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/service/schema"
)

// 数据处理和分发
// 格式化每个 JSON 数据, 附加系统字段, 发送给 ES 和 API 队列, 释放 DataItem
func dataProcessor(dp *processor) {
	defer releaseDataProcessor(dp)

	// 丢弃可选写入 ES 数据项
	isDiscards := ESOptionalWrite.Load() && dp.dr.apiConf.ESOptionalWrite
	if isDiscards {
		ESDataItemDiscards.Inc()
	}
	isPostToES := !(isDiscards || ESDisableWrite.Load() || dp.dr.apiConf.ESDisableWrite)
	isPostToAPI := dp.dr.apiConf.PostAPI.Interval > 0
	if !isPostToES && !isPostToAPI {
		return
	}

	// 兼容 {body} 或 {body}=-:-=[{body},{body}]
	dp.data.Body = pretty.UglyInPlace(dp.data.Body)
	for _, js := range bytes.Split(dp.data.Body, esBodySep) {
		if len(js) < jsonMinLen {
			continue
		}

		if !gjson.ValidBytes(js) {
			sampler.Info().
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

func sendOneData(dp *processor, js []byte, isPostToES, isPostToAPI bool) {
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
		ESChan.In <- item
	}
	if isPostToAPI {
		// 需要在 API 使用后回收 DataItem
		item := schema.Make()
		item.Body = jsData
		dp.dr.apiChan.In <- item
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
		buf = append(buf, common.Now3339Z...)
		buf = append(buf, `","_gtime":"`...)
		buf = append(buf, common.Now3339...)
		buf = append(buf, `",`...)
		i = 1
	}

	buf = append(buf, js[i:]...)

	return buf
}
