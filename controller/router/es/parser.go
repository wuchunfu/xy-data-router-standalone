package es

import (
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/middleware"
)

// 解析执行结果, 记录日志, 发送响应
func sendResult(c *fiber.Ctx, resp *tResponse, params *tParams) error {
	params.ClientIP = common.GetClientIP(c)
	ret := parseESResponse(resp, params)
	defer func() {
		log(params, ret)
		putResult(ret)
	}()

	if ret.ErrMsg != "" {
		return middleware.APIFailure(c, ret.ErrMsg)
	}
	return middleware.APISuccessBytes(c, ret.result, ret.Count)
}

// 执行 ES 请求
func parseESResponse(resp *tResponse, params *tParams) *tResult {
	defer putResponse(resp)
	ret := getResult()
	ret.StatusCode = resp.response.StatusCode

	if resp.err != nil {
		common.LogSampled.Error().Err(resp.err).Msg("getting response")
		ret.ErrMsg = "查询失败, 服务繁忙"
		ret.Error = resp.err.Error()
		return ret
	}

	if resp.response.Body == nil {
		common.LogSampled.Error().Int("status_code", ret.StatusCode).Msg("response.Body is nil")
		ret.ErrMsg = "查询失败, 服务异常"
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
		ret.ErrMsg = "查询失败, 请稍后重试"
		ret.Error = err.Error()
		return ret
	}

	if resp.response.IsError() {
		ret.ErrMsg = "查询失败, 查询语句有误"
		common.LogSampled.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Int("http_code", resp.response.StatusCode).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Str("error_type", gjson.GetBytes(res, "error.type").String()).
			Str("error_reason", gjson.GetBytes(res, "error.reason").String()).
			Msg(ret.ErrMsg)
		return ret
	}

	ret.result = res

	// 慢查询日志
	ret.Took = gjson.GetBytes(res, "took").Int()
	costTime := time.Duration(ret.Took) * time.Millisecond
	if costTime > conf.Config.WebConf.ESSlowQueryDuration {
		common.LogSampled.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Dur("duration", costTime).
			Msgf("es search slow, timeout: %v", gjson.GetBytes(res, "timed_out"))
	}

	ret.Count = int(gjson.GetBytes(res, resp.totalPath).Int())
	return ret
}
