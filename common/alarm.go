package common

import (
	"github.com/imroc/req/v3"

	"github.com/fufuok/xy-data-router/conf"
)

// 发送报警消息
func sendAlarm(bs []byte) {
	if _, err := req.SetBodyJsonBytes(bs).Post(conf.Config.LogConf.PostAlarmAPI); err != nil {
		LogSampled.Warn().Err(err).Str("url", conf.Config.LogConf.PostAlarmAPI).Msg("sendAlarm")
	}
}
