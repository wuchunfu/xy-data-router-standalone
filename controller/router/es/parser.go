package es

import (
	"io"
	"time"

	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
)

// 执行 ES 请求
func parseESResponse(resp *tResponse, params *tParams) *tResult {
	defer putResponse(resp)
	ret := getResult()

	if resp.err != nil {
		common.LogSampled.Error().Err(resp.err).Msg("getting response")
		ret.errMsg = "查询失败, 服务繁忙"
		return ret
	}

	if resp.response.Body == nil {
		common.LogSampled.Error().Err(resp.err).Msg("response.Body is nil")
		ret.errMsg = "查询失败, 服务异常"
		return ret
	}

	defer func() {
		_ = resp.response.Body.Close()
	}()

	return parseESResult(resp, params, ret)
}

// 处理搜索结果
func parseESResult(resp *tResponse, params *tParams, ret *tResult) *tResult {
	res, err := io.ReadAll(resp.response.Body)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("response.Body")
		ret.errMsg = "查询失败, 请稍后重试"
		return ret
	}

	if resp.response.IsError() {
		ret.errMsg = "查询失败, 查询语句有误"
		common.LogSampled.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Int("http_code", resp.response.StatusCode).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Str("error_type", gjson.GetBytes(res, "error.type").String()).
			Str("error_reason", gjson.GetBytes(res, "error.reason").String()).
			Msg(ret.errMsg)
		return ret
	}

	ret.result = res

	// 慢查询日志
	took := gjson.GetBytes(res, "took").Int()
	costTime := time.Duration(took) * time.Millisecond
	if costTime > conf.Config.SYSConf.ESSlowQueryDuration {
		common.LogSampled.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Dur("duration", costTime).
			Msgf("es search slow, timeout: %v", gjson.GetBytes(res, "timed_out"))
	}

	ret.count = int(gjson.GetBytes(res, resp.totalPath).Int())
	return ret
}
