package controller

import (
	"context"
	"io"
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/middleware"
)

type tESSearch struct {
	Index    string                 `json:"index"`
	Scroll   int                    `json:"scroll"`
	ScrollID string                 `json:"scroll_id"`
	Body     map[string]interface{} `json:"body"`
	ClientIP string
}

// ESSearchHandler ES 通用查询接口
func ESSearchHandler(c *fiber.Ctx) error {
	esSearch := new(tESSearch)
	if err := c.BodyParser(esSearch); err != nil || esSearch.Index == "" || esSearch.Body == nil {
		return middleware.APIFailure(c, "查询参数有误")
	}

	bodyBuf := bufferpool.Get()
	defer bufferpool.Put(bodyBuf)

	_ = json.NewEncoder(bodyBuf).Encode(esSearch.Body)

	resp, err := common.ES.Search(
		common.ES.Search.WithContext(context.Background()),
		common.ES.Search.WithIndex(esSearch.Index),
		common.ES.Search.WithScroll(time.Duration(esSearch.Scroll)*time.Second),
		common.ES.Search.WithBody(bodyBuf),
		common.ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es search, getting response")
		return middleware.APIFailure(c, "查询失败, 服务繁忙")
	}

	if resp.Body == nil {
		common.LogSampled.Error().Err(err).Msg("es search, resp.Body is nil")
		return middleware.APIFailure(c, "查询失败, 服务异常")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	esSearch.ClientIP = c.IP()
	res, count, msg := parseESSearch(resp, esSearch)
	if msg != "" {
		return middleware.APIFailure(c, msg)
	}

	return middleware.APISuccessBytes(c, res, count)
}

// 处理搜索结果
func parseESSearch(resp *esapi.Response, esSearch *tESSearch) (res []byte, count int, msg string) {
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		msg = "查询失败, 请稍后重试"
		return
	}

	if resp.IsError() {
		msg = "查询失败, 查询语句有误"
		common.LogSampled.Warn().
			Bytes("body", json.MustJSON(esSearch)).Int("http_code", resp.StatusCode).
			Str("error_type", gjson.GetBytes(res, "error.type").String()).
			Str("error_reason", gjson.GetBytes(res, "error.reason").String()).
			Str("msg", msg).Msg("es search")
		return
	}

	// 慢查询日志
	took := gjson.GetBytes(res, "took").Int()
	costTime := time.Duration(took) * time.Millisecond
	if costTime > conf.Config.SYSConf.ESSlowQueryDuration {
		common.LogSampled.Warn().
			Bytes("body", json.MustJSON(esSearch)).Dur("duration", costTime).
			Msgf("es search slow, timeout: %v", gjson.GetBytes(res, "timed_out"))
	}

	count = int(gjson.GetBytes(res, "hits.total").Int())
	return
}
