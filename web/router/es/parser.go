package es

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fufuok/utils"
	"github.com/fufuok/utils/xjson/gjson"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/service/es"
	"github.com/fufuok/xy-data-router/web/response"
)

// 发送响应, 记录日志
func sendResult(c *fiber.Ctx, resp *es.Response, params *params) error {
	params.ClientIP = common.GetClientIP(c)
	ret := parseESResponse(resp, params)
	ret.ReqUri = utils.CopyString(c.OriginalURL())
	ret.ReqTime = common.Now.Str3339
	ret.ReqType = conf.APPName
	defer func() {
		log(params, ret)
		putResult(ret)
	}()

	if ret.ErrMsg != "" {
		return response.APIFailure(c, ret.ErrMsg, ret.Error)
	}
	return response.APISuccessBytes(c, ret.result, ret.Count)
}

// 解析 ES 请求结果
func parseESResponse(resp *es.Response, params *params) *result {
	ret := getResult()
	if resp.Err != nil || resp.Response == nil {
		sampler.Error().Err(resp.Err).Msg("getting response")
		ret.ErrMsg = "请求失败, 服务繁忙"
		ret.Error = resp.Err.Error()
		if errors.Is(resp.Err, fasthttp.ErrTimeout) {
			ret.ErrMsg = fmt.Sprintf("ESAPI 请求超时: %s", defaultESAPITimeout.String())
		}
		return ret
	}

	ret.StatusCode = resp.Response.StatusCode
	if resp.Response.Body == nil {
		sampler.Error().Int("status_code", ret.StatusCode).Msg("response.Body is nil")
		ret.ErrMsg = "请求失败, 服务异常"
		return ret
	}

	defer func() {
		_ = resp.Response.Body.Close()
	}()

	return parseESResult(resp, params, ret)
}

// 处理搜索结果
func parseESResult(resp *es.Response, params *params, ret *result) *result {
	res, err := io.ReadAll(resp.Response.Body)
	if err != nil {
		sampler.Error().Err(err).Msg("response.Body")
		ret.ErrMsg = "请求失败, 请稍后重试"
		ret.Error = err.Error()
		return ret
	}

	if resp.Response.IsError() {
		ret.Error = gjson.GetBytes(res, "error.reason").String()
		if ret.Error == "" {
			ret.Error = gjson.GetBytes(res, "error").String()
			if ret.Error == "" {
				ret.Error = "Status: " + resp.Response.Status()
			}
		}
		ret.ErrMsg = "请求错误, 请检查参数"
		sampler.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Int("http_code", resp.Response.StatusCode).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Str("error_type", gjson.GetBytes(res, "error.type").String()).
			Str("error_reason", ret.Error).
			Msg(ret.ErrMsg)
		return ret
	}

	ret.result = res

	// 慢查询日志
	ret.Took = gjson.GetBytes(res, "took").Int()
	costTime := time.Duration(ret.Took) * time.Millisecond
	if costTime > conf.Config.WebConf.ESSlowQueryDuration {
		sampler.Warn().
			RawJSON("body", json.MustJSON(params.Body)).
			Str("client_ip", params.ClientIP).
			Str("index", params.Index).
			Dur("duration", costTime).
			Msgf("es search slow, timeout: %v", gjson.GetBytes(res, "timed_out"))
	}

	ret.Count = int(gjson.GetBytes(res, resp.TotalPath).Int())
	return ret
}
