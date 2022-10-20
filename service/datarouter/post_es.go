package datarouter

import (
	"time"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/jsongen"
	"github.com/fufuok/utils/pools/readerpool"
	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/service/es"
)

// ES 批量写入协程池初始化
func initESBulkPool() {
	ESBulkPool, _ = ants.NewPoolWithFunc(
		conf.Config.DataConf.ESBulkWorkerSize,
		func(v any) {
			esBulk(v.(*tDataItems))
		},
		ants.WithExpiryDuration(10*time.Second),
		ants.WithMaxBlockingTasks(conf.Config.DataConf.ESBulkMaxWorkerSize),
		ants.WithLogger(common.NewAppLogger()),
		ants.WithPanicHandler(func(r any) {
			common.LogSampled.Error().Msgf("Recovery dataProcessor: %s", r)
		}),
	)
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

	jsIndex := jsongen.NewMap()
	js := jsongen.NewMap()
	js.PutString("_index", esIndex)
	if es.ServerLessThan7 {
		js.PutString("_type", "_doc")
	}
	if apiConf.ESPipeline != "" {
		js.PutString("pipeline", apiConf.ESPipeline)
	}
	jsIndex.PutMap("index", js)
	bs := jsIndex.Serialize(nil)
	bs = append(bs, '\n')
	return bs
}

// 每日更新所有接口 esBluk 索引头
func updateESBulkHeader() {
	time.Sleep(10 * time.Second)
	for {
		now := common.GTimeNow()
		ymd := now.Format("060102")
		dataRouters.Range(func(_ string, dr *tDataRouter) bool {
			dr.apiConf.ESBulkHeader = getESBulkHeader(dr.apiConf, ymd)
			dr.apiConf.ESBulkHeaderLength = len(dr.apiConf.ESBulkHeader)
			return true
		})
		// 等待明天 0 点再执行
		now = common.GTimeNow()
		time.Sleep(utils.BeginOfTomorrow(now).Sub(now))
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
		case item, ok := <-ESChan.Out:
			if !ok {
				// 消费所有数据
				if dis.count > 0 {
					submitESBulk(dis)
				}
				common.Log.Error().Msg("PostES worker exited")
				return
			}

			dis.add(item)

			// 达到限定数量或大小写入 ES
			if dis.count%conf.Config.DataConf.ESPostBatchNum == 0 || dis.size > conf.Config.DataConf.ESPostBatchBytes {
				submitESBulk(dis)
				dis = getDataItems()
			}

			ESDataTotal.Inc()
		}
	}
}

// 提交批量任务, 提交不阻塞, 有执行并发限制, 最大排队数限制
func submitESBulk(dis *tDataItems) {
	_ = common.GoPool.Submit(func() {
		ESBulkTodoCount.Inc()
		if err := ESBulkPool.Invoke(dis); err != nil {
			ESBulkDiscards.Inc()
			ESBulkTodoCount.Dec()
			common.LogSampled.Warn().Err(err).Msg("go esBulk discards")
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
		ESBulkTodoCount.Dec()
		bytespool.Release(esBody)
		readerpool.Release(rBody)
	}()

	return es.BulkRequest(rBody, esBody)
}
