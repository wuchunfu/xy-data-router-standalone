package service

import (
	"time"

	"github.com/fufuok/bytespool"
	"github.com/imroc/req"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/schema"
)

func apiWorker(dr *tDataRouter) {
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
		case v, ok := <-dr.drOut.apiChan.Out:
			if !ok {
				// 消费所有数据
				if dis.count > 0 {
					postAPI(dis, dr.apiConf.PostAPI.API)
				}
				common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("apiWorker exited")
				return
			}
			dis.add(v.(*schema.DataItem))
		}
	}
}

// 推送数据到 API
func postAPI(dis *tDataItems, api []string) {
	// [json,json]
	apiBody := bytespool.Make(dis.size + dis.count + 1)
	apiBody = append(apiBody, '[')
	for i := 0; i < dis.count; i++ {
		apiBody = append(apiBody, dis.items[i].Body...)
		apiBody = append(apiBody, ',')
	}
	apiBody[len(apiBody)-1] = ']'
	dis.release()
	_ = common.GoPool.Submit(func() {
		for _, u := range api {
			if _, err := req.Post(u, req.BodyJSON(apiBody), conf.ReqUserAgent); err != nil {
				common.LogSampled.Error().Err(err).Str("url", u).Msg("apiWorker")
			}
		}
	})
}
