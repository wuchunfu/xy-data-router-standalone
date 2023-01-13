package master

import (
	"context"
	"reflect"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger"
)

// 初始化获取远端配置
func startRemoteConf(ctx context.Context) {
	// 定时获取远程主配置
	if conf.Config.MainConf.GetConfDuration > 0 {
		go getRemoteConf(ctx, &conf.Config.MainConf)
	}
}

// 执行获取远端配置
func getRemoteConf(ctx context.Context, c *conf.FilesConf) {
	v := reflect.ValueOf(c)
	m := v.MethodByName(c.Method)
	if m.Kind() != reflect.Func {
		logger.Error().Str("method", c.Method).Msg("skip init get remote conf(func error)")
		return
	}

	logger.Info().
		Str("path", c.Path).Str("method", c.Method).Dur("duration", c.GetConfDuration).
		Msg("start get remote conf")

	ticker := common.TWs.NewTicker(c.GetConfDuration)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			logger.Warn().
				Str("path", c.Path).Str("method", c.Method).
				Msg("exit get remote conf")
			return
		default:
			res := m.Call(nil)
			if !res[0].IsNil() {
				logger.Error().
					Str("path", c.Path).Str("method", c.Method).
					Msgf("get remote conf err: %s", res[0])
			}
		}
	}
}
