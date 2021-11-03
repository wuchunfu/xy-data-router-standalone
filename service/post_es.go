package service

import (
	"bytes"
	"time"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/pools/readerpool"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
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
		conf.Config.SYSConf.ESBulkWorkerSize,
		func(i interface{}) {
			esBulk(i.([]byte))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.SYSConf.ESBulkMaxWorkerSize),
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

	return utils.AddStringBytes(`{"index":{"_index":"`, esIndex, `","_type":"_doc"}}`, "\n")
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
	ticker := common.TWms.NewTicker(conf.Config.SYSConf.ESPostMaxIntervalDuration)
	defer ticker.Stop()

	var buf bytes.Buffer
	n := 0
	for {
		select {
		case <-ticker.C:
			// 达到最大时间间隔写入 ES
			postES(&buf)
		case v, ok := <-esChan.Out:
			if !ok {
				// 消费所有数据
				postES(&buf)
				common.Log.Error().Msg("PostES worker exited")
				return
			}

			esData := v.([]byte)
			buf.Write(esData)
			bytespool.Release(esData)

			n += 1
			if n%conf.Config.SYSConf.ESPostBatchNum == 0 || buf.Len() > conf.Config.SYSConf.ESPostBatchBytes {
				// 达到限定数量或大小写入 ES
				postES(&buf)
				n = 0
			}

			esDataTotal.Inc()
		}
	}
}

// 创建写入协程
func postES(buf *bytes.Buffer) {
	if buf.Len() == 0 {
		return
	}

	esBody := bytespool.New(buf.Len())
	copy(esBody, buf.Bytes())
	buf.Reset()

	submitESBulk(esBody)
}

// 提交批量任务, 提交不阻塞, 有执行并发限制, 最大排队数限制
func submitESBulk(esBody []byte) {
	_ = common.Pool.Submit(func() {
		esBulkTodoCount.Inc()
		if err := esBulkPool.Invoke(esBody); err != nil {
			esBulkDiscards.Inc()
			common.LogSampled.Error().Err(err).Msg("go esBulk")
		}
	})
}

// 批量写入 ES
func esBulk(esBody []byte) {
	r := readerpool.New(esBody)

	defer func() {
		esBulkTodoCount.Dec()
		bytespool.Release(esBody)
		readerpool.Release(r)
	}()

	resp, err := common.ES.Bulk(r)
	if err != nil {
		common.LogSampled.Error().Err(err).Msg("es bulk")
		esBulkErrors.Inc()
		return
	}

	// 批量写入完成计数
	esBulkCount.Inc()

	defer func() {
		_ = resp.Body.Close()
	}()

	// 低级别日志配置时(Warn), 开启批量写入错误抽样日志, Error 时关闭批量写入错误日志
	if conf.Config.SYSConf.Log.Level <= int(zerolog.WarnLevel) {
		var res map[string]interface{}
		var blk tESBulkResponse

		if resp.IsError() {
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				common.LogSampled.Error().Err(err).
					Str("resp", resp.String()).
					Msg("es bulk, parsing the response body")
			} else {
				esBulkErrors.Inc()
				common.LogSampled.Error().
					Int("http_code", resp.StatusCode).
					Msgf("es bulk, err: %+v", res["error"])
			}
			return
		}

		// 低级别批量日志时(Error), 解析批量写入结果
		if conf.Config.SYSConf.Log.ESBulkLevel <= int(zerolog.ErrorLevel) {
			if err := json.NewDecoder(resp.Body).Decode(&blk); err != nil {
				common.LogSampled.Error().Err(err).
					Str("resp", resp.String()).
					Msg("es bulk, parsing the response body")
			} else if blk.Errors {
				for _, d := range blk.Items {
					if d.Index.Status > 201 {
						esBulkErrors.Inc()

						// error: [429] es_rejected_execution_exception
						// 等待一个提交周期, 重新排队
						if utils.InInts(conf.Config.SYSConf.ESReentryCodes, d.Index.Status) {
							time.Sleep(conf.Config.SYSConf.ESPostMaxIntervalDuration)
							// 拷贝数据重新提交
							esBodyNew := bytespool.New(len(esBody))
							copy(esBodyNew, esBody)
							submitESBulk(esBodyNew)
						}

						// Warn 级别时, 抽样数据详情
						if conf.Config.SYSConf.Log.ESBulkLevel <= int(zerolog.WarnLevel) {
							common.LogSampled.Error().
								Msgf("error: [%d] %s; %s; %s; %s\n||\n%s||",
									d.Index.Status,
									d.Index.Error.Type,
									d.Index.Error.Reason,
									d.Index.Error.Cause.Type,
									d.Index.Error.Cause.Reason,
									utils.B2S(esBody),
								)
						} else {
							common.LogSampled.Error().
								Msgf("error: [%d] %s; %s; %s; %s",
									d.Index.Status,
									d.Index.Error.Type,
									d.Index.Error.Reason,
									d.Index.Error.Cause.Type,
									d.Index.Error.Cause.Reason,
								)
						}
					}
				}
			}
		}
	}
}
