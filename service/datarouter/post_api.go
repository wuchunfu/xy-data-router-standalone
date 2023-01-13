package datarouter

import (
	"sync"
	"time"

	"github.com/fufuok/bytespool"
	"github.com/imroc/req/v3"

	"github.com/fufuok/xy-data-router/common"
)

func apiWorker(dr *router) {
	// 默认 1s 空转
	interval := 1
	if dr.apiConf.PostAPI.Interval > 0 {
		interval = dr.apiConf.PostAPI.Interval
		common.Log.Info().
			Int("interval", interval).Str("apiname", dr.apiConf.APIName).
			Msg("apiWorker start")
	}
	ticker := common.TWs.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	dis := getDataItems()
	for {
		select {
		case <-ticker.C:
			if dr.apiConf.PostAPI.Interval > 0 {
				// 校准推送周期
				if dr.apiConf.PostAPI.Interval != interval {
					interval = dr.apiConf.PostAPI.Interval
					ticker.Reset(time.Duration(interval) * time.Second)
					common.Log.Warn().
						Int("interval", interval).Str("apiname", dr.apiConf.APIName).
						Msg("apiWorker restart")
				}
				if dis.count == 0 {
					continue
				}
				// 指定时间间隔推送数据
				postAPI(dis, dr.apiConf.PostAPI.API)
				dis = getDataItems()
			}
		case item, ok := <-dr.apiChan.Out:
			if !ok {
				// 消费所有数据
				if dis.count > 0 {
					postAPI(dis, dr.apiConf.PostAPI.API)
				}
				common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("apiWorker exited")
				return
			}

			dis.add(item)

			// 达到限定数量或大小时推送数据
			if dis.count%dr.apiConf.PostAPI.BatchNum == 0 || dis.size > dr.apiConf.PostAPI.BatchBytes {
				postAPI(dis, dr.apiConf.PostAPI.API)
				dis = getDataItems()
			}
		}
	}
}

// 推送数据到 API
func postAPI(dis *dataItems, api []string) {
	defer dis.release()
	if len(api) == 0 {
		return
	}
	// [json,json]
	apiBody := bytespool.Make(dis.size + dis.count + 1)
	apiBody = append(apiBody, '[')
	for i := 0; i < dis.count; i++ {
		apiBody = append(apiBody, dis.items[i].Body...)
		apiBody = append(apiBody, ',')
	}
	apiBody[len(apiBody)-1] = ']'
	// 分发数据到接口 POST JSON
	_ = common.GoPool.Submit(func() {
		defer bytespool.Release(apiBody)
		var wg sync.WaitGroup
		for _, u := range api {
			wg.Add(1)
			apiUrl := u
			_ = common.GoPool.Submit(func() {
				defer wg.Done()
				if _, err := req.SetBodyJsonBytes(apiBody).Post(apiUrl); err != nil {
					common.LogSampled.Error().Err(err).Str("url", apiUrl).Msg("apiWorker")
				}
			})
		}
		wg.Wait()
	})
}
