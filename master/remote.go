package master

import (
	"context"
	"reflect"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

// 初始化获取远端配置
func startRemoteConf(ctx context.Context) {
	// 定时获取远程主配置
	if conf.Config.SYSConf.MainConfig.GetConfDuration > 0 {
		go getRemoteConf(ctx, &conf.Config.SYSConf.MainConfig)
	}
}

// 执行获取远端配置
func getRemoteConf(ctx context.Context, c *conf.TFilesConf) {
	v := reflect.ValueOf(c)
	m := v.MethodByName(c.Method)
	if m.Kind() != reflect.Func {
		common.Log.Error().Str("method", c.Method).Msg("skip init get remote conf(func error)")
		return
	}

	common.Log.Info().
		Str("path", c.Path).Str("method", c.Method).Dur("duration", c.GetConfDuration).
		Msg("start get remote conf")

	ticker := common.TW.NewTicker(c.GetConfDuration)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			common.Log.Warn().
				Str("path", c.Path).Str("method", c.Method).
				Msg("exit get remote conf")
			return
		default:
			res := m.Call(nil)
			if !res[0].IsNil() {
				common.Log.Error().
					Str("path", c.Path).Str("method", c.Method).
					Msgf("get remote conf err: %s", res[0])
			}
		}
	}
}
