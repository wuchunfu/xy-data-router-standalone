package service

import (
	"bytes"
	"time"

	"github.com/imroc/req"
	bbPool "github.com/valyala/bytebufferpool"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
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

	var buf bytes.Buffer
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
				// 指定时间间隔推送数据
				postAPI(&buf, dr.apiConf.PostAPI.API)
			}
		case js, ok := <-dr.drOut.apiChan.Out:
			if !ok {
				// 消费所有数据
				postAPI(&buf, dr.apiConf.PostAPI.API)
				common.Log.Warn().Str("apiname", dr.apiConf.APIName).Msg("apiWorker exited")
				return
			}
			buf.WriteByte(',')
			buf.Write(js.([]byte))
		}
	}
}

// 推送数据到 API
func postAPI(buf *bytes.Buffer, api []string) {
	if buf.Len() == 0 {
		return
	}

	defer buf.Reset()

	body := bbPool.Get()
	_, _ = body.Write(jsArrLeft)
	_, _ = body.Write(buf.Bytes()[1:])
	_, _ = body.Write(jsArrRight)

	_ = common.Pool.Submit(func() {
		defer bbPool.Put(body)
		for _, u := range api {
			if _, err := req.Post(u, req.BodyJSON(body.Bytes()), conf.ReqUserAgent); err != nil {
				common.LogSampled.Error().Err(err).Str("url", u).Msg("apiWorker")
			}
		}
	})
}
