package service

import (
	"time"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/fufuok/utils/pools/readerpool"
	"github.com/panjf2000/ants/v2"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/schema"
)

// ES 批量写入响应
type tESBulkResponse struct {
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

// ES 批量写入协程池初始化
func initESBulkPool() {
	esBulkPool, _ = ants.NewPoolWithFunc(
		conf.Config.DataConf.ESBulkWorkerSize,
		func(v interface{}) {
			esBulk(v.(*tDataItems))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.DataConf.ESBulkMaxWorkerSize),
		ants.WithPanicHandler(func(r interface{}) {
			common.LogSampled.Error().Interface("recover", r).Msg("panic")
		}),
	)
}

// 获取 ES 索引名称
func getUDPESIndex(body []byte, key string) string {
	index := gjson.GetBytes(body, key).String()
	if index != "" {
		return utils.ToLower(utils.Trim(index, ' '))
	}

	return ""
}

// 生成 esBluk 索引头
func getESBulkHeader(apiConf *conf.TAPIConf, ymd string) []byte {
	esIndex := apiConf.ESIndex
	if esIndex == "" {
		esIndex = apiConf.APIName
	}

	// 索引切割: ff ff_21 ff_2105 ff_210520
	switch apiConf.ESIndexSplit {
	case "year":
		esIndex = esIndex + "_" + ymd[:2]
	case "month":
		esIndex = esIndex + "_" + ymd[:4]
	case "none":
		break
	default:
		esIndex = esIndex + "_" + ymd
	}

	indexType := ""
	if common.ESLessThan7 {
		indexType = `","_type":"_doc`
	}

	return utils.AddStringBytes(`{"index":{"_index":"`, esIndex, indexType, `"}}`, "\n")
}

// 每日更新所有接口 esBluk 索引头
func updateESBulkHeader() {
	time.Sleep(10 * time.Second)
	for {
		now := common.GetGlobalTime()
		ymd := now.Format("060102")
		dataRouters.Range(func(_ string, value interface{}) bool {
			dr := value.(*tDataRouter)
			dr.apiConf.ESBulkHeader = getESBulkHeader(dr.apiConf, ymd)
			dr.apiConf.ESBulkHeaderLength = len(dr.apiConf.ESBulkHeader)
			return true
		})
		// 等待明天 0 点再执行
		now = common.GetGlobalTime()
		time.Sleep(utils.Get0Tomorrow(now).Sub(now))
	}
}

// ES 数据入口
func esWorker() {
	ticker := common.TWms.NewTicker(conf.Config.DataConf.ESPostMaxIntervalDuration)
	defer ticker.Stop()

	dis := getDataItems()
	for {
		select {
		case <-ticker.C:
			// 达到最大时间间隔写入 ES
			if dis.count == 0 {
				continue
			}
			submitESBulk(dis)
			dis = getDataItems()
		case v, ok := <-esChan.Out:
			if !ok {
				// 消费所有数据
				if dis.count > 0 {
					submitESBulk(dis)
				}
				common.Log.Error().Msg("PostES worker exited")
				return
			}

			dis.add(v.(*schema.DataItem))

			// 达到限定数量或大小写入 ES
			if dis.count%conf.Config.DataConf.ESPostBatchNum == 0 || dis.size > conf.Config.DataConf.ESPostBatchBytes {
				submitESBulk(dis)
				dis = getDataItems()
			}

			esDataTotal.Inc()
		}
	}
}

// 提交批量任务, 提交不阻塞, 有执行并发限制, 最大排队数限制
func submitESBulk(dis *tDataItems) {
	_ = common.GoPool.Submit(func() {
		esBulkTodoCount.Inc()
		if err := esBulkPool.Invoke(dis); err != nil {
			esBulkDiscards.Inc()
			esBulkTodoCount.Dec()
			common.LogSampled.Error().Err(err).Msg("go esBulk")
		}
	})
}

// 批量写入 ES
func esBulk(dis *tDataItems) bool {
	esBody := bytespool.Make(dis.size)
	for i := 0; i < dis.count; i++ {
		esBody = append(esBody, dis.items[i].Body...)
	}
	rBody := readerpool.New(esBody)
	dis.release()

	defer func() {
		esBulkTodoCount.Dec()
		bytespool.Release(esBody)
		readerpool.Release(rBody)
	}()

	resp, err := common.ES.Bulk(rBody)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es bulk")
		esBulkErrors.Inc()
		return false
	}

	// 批量写入完成计数
	esBulkCount.Inc()

	if resp.Body == nil {
		return false
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	return esBulkResult(resp, esBody)
}

func esBulkResult(resp *esapi.Response, esBody []byte) bool {
	// 低级别日志配置时(Warn), 开启批量写入错误抽样日志, Error 时关闭批量写入错误日志
	if !conf.Config.StateConf.CheckESBulkResult {
		return true
	}

	if resp.IsError() {
		esBulkErrors.Inc()
		buf := bufferpool.Get()
		defer bufferpool.Put(buf)
		if _, err := buf.ReadFrom(resp.Body); err != nil {
			return false
		}
		common.LogSampled.Warn().
			Int("http_code", resp.StatusCode).
			Str("error_type", gjson.GetBytes(buf.Bytes(), "error.type").String()).
			Str("error_reason", gjson.GetBytes(buf.Bytes(), "error.reason").String()).
			Msg("es bulk")
		return false
	}

	// 低级别批量日志时(Warn), 解析批量写入结果
	if !conf.Config.StateConf.CheckESBulkErrors {
		return true
	}

	var blk tESBulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&blk); err != nil {
		common.LogSampled.Error().Err(err).
			Str("resp", resp.String()).
			Str("error_reason", "failure to to parse response body").
			Msg("es bulk")
		return false
	}

	if !blk.Errors {
		return true
	}

	esBulkErrors.Inc()

	for _, d := range blk.Items {
		if d.Index.Status <= 201 {
			continue
		}
		l := common.LogSampled.Warn().Int("status", d.Index.Status).
			Str("error_type", d.Index.Error.Type).
			Str("error_reason", d.Index.Error.Reason).
			Str("error_cause_type", d.Index.Error.Cause.Type).
			Str("error_cause_reason", d.Index.Error.Cause.Reason)

		// Warn 级别时, 抽样数据详情
		if conf.Config.StateConf.RecordESBulkBody {
			l.Bytes("body", esBody)
		}
		l.Msg("es bulk")
	}

	return false
}
