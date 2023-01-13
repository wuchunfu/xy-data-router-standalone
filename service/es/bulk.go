package es

import (
	"io"

	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
)

// ES 批量写入响应
type esBulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

// BulkRequest 批量索引文档
func BulkRequest(body io.Reader, esBody []byte) bool {
	resp := GetResponse()
	defer PutResponse(resp)
	resp.Response, resp.Err = Client.Bulk(body)
	if resp.Err != nil {
		sampler.Error().Err(resp.Err).Msg("es bulk")
		BulkErrors.Inc()
		return false
	}

	// 批量写入完成计数
	BulkCount.Inc()

	if resp.Response.Body == nil {
		return false
	}

	defer func() {
		_ = resp.Response.Body.Close()
	}()

	return esBulkResult(resp, esBody)
}

func esBulkResult(resp *Response, esBody []byte) bool {
	// 低级别日志配置时(Warn), 开启批量写入错误抽样日志, Error 时关闭批量写入错误日志
	if !conf.Config.StateConf.CheckESBulkResult {
		return true
	}

	if resp.Response.IsError() {
		BulkErrors.Inc()
		buf := bufferpool.Get()
		defer bufferpool.Put(buf)
		if _, err := buf.ReadFrom(resp.Response.Body); err != nil {
			return false
		}
		sampler.Warn().
			Int("http_code", resp.Response.StatusCode).
			Str("error_type", gjson.GetBytes(buf.Bytes(), "error.type").String()).
			Str("error_reason", gjson.GetBytes(buf.Bytes(), "error.reason").String()).
			Msg("es bulk")
		return false
	}

	// 低级别批量日志时(Warn), 解析批量写入结果
	if !conf.Config.StateConf.CheckESBulkErrors {
		return true
	}

	var blk esBulkResponse
	if err := json.NewDecoder(resp.Response.Body).Decode(&blk); err != nil {
		sampler.Error().Err(err).
			Str("resp", resp.Response.String()).
			Str("error_reason", "failure to to parse response body").
			Msg("es bulk")
		return false
	}

	if !blk.Errors {
		return true
	}

	BulkErrors.Inc()

	i := 0
	for _, d := range blk.Items {
		if d.Index.Status <= 201 {
			continue
		}
		l := sampler.Warn().Int("status", d.Index.Status).
			Str("error_type", d.Index.Error.Type).
			Str("error_reason", d.Index.Error.Reason).
			Str("error_cause_type", d.Index.Error.Cause.Type).
			Str("error_cause_reason", d.Index.Error.Cause.Reason)

		// Warn 级别时, 抽样数据详情
		if i == 0 && conf.Config.StateConf.RecordESBulkBody {
			i++
			l.Bytes("body", esBody)
		}
		l.Msg("es bulk")
	}

	return false
}
