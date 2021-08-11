package controller

import (
	"bytes"
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v6/esapi"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/json"
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
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

	var bodyBuf bytes.Buffer
	_ = json.NewEncoder(&bodyBuf).Encode(esSearch.Body)

	resp, err := common.ES.Search(
		common.ES.Search.WithContext(context.Background()),
		common.ES.Search.WithIndex(esSearch.Index),
		common.ES.Search.WithScroll(time.Duration(esSearch.Scroll)*time.Second),
		common.ES.Search.WithBody(&bodyBuf),
		common.ES.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es search, getting response")
		return middleware.APIFailure(c, "查询失败, 服务繁忙")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	esSearch.ClientIP = c.IP()
	res, count, msg := parseESSearch(resp, esSearch)
	if msg != "" {
		return middleware.APIFailure(c, msg)
	}

	return middleware.APISuccess(c, res, count)
}

// 处理搜索结果
func parseESSearch(resp *esapi.Response, esSearch *tESSearch) (map[string]interface{}, int, string) {
	var res map[string]interface{}
	if resp.IsError() {
		msg := "查询失败, 查询语句有误"
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			common.LogSampled.Warn().Err(err).
				Str("resp", resp.String()).
				Msg("es search, parsing the response body")
		} else {
			common.LogSampled.Warn().
				Bytes("body", utils.MustJSON(esSearch)).Int("http_code", resp.StatusCode).
				Msgf("es search, %s, %+v", msg, res["error"])
		}

		return nil, 0, msg
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		common.LogSampled.Warn().Err(err).
			Str("resp", resp.String()).
			Msg("es search, parsing the response body")
		return nil, 0, "查询失败, 请稍后重试"
	}

	// 慢查询日志
	costTime := time.Duration(int(res["took"].(float64))) * time.Millisecond
	if costTime > conf.Config.SYSConf.ESSlowQueryDuration {
		common.LogSampled.Warn().
			Bytes("body", utils.MustJSON(esSearch)).Dur("duration", costTime).
			Msgf("es search slow, timeout: %s", res["timed_out"])
	}

	return res, int(res["hits"].(map[string]interface{})["total"].(float64)), ""
}
