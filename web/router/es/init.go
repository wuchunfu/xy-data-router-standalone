package es

import (
	"time"

	"github.com/fufuok/xy-data-router/conf"
)

const (
	// ES 查询请求代理时的超时秒数附加值
	esProxyTimeoutAdd = 2 * time.Second
)

var (
	// 代理请求超时时间
	esProxyTimeout time.Duration

	// ESAPI 超时时间
	defaultESAPITimeout time.Duration
)

// InitMain 程序启动时初始化
func InitMain() {
	initTimeout()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	initTimeout()
}

func Stop() {}

func initTimeout() {
	// 代理超时 > HTTP 请求超时 = ES 超时参数(规定时间内结束查询)
	defaultESAPITimeout = conf.Config.WebConf.ESAPITimeout
	esProxyTimeout = defaultESAPITimeout + esProxyTimeoutAdd
}

// 解析参数中的 timeout_ms, 快速返回, 注意判断 ES 结果中 timed_out: true/false
func getTimeoutParams(n int) time.Duration {
	if n < 1 {
		return defaultESAPITimeout
	}
	dur := time.Duration(n) * time.Millisecond
	if dur > defaultESAPITimeout {
		return defaultESAPITimeout
	}
	return dur
}
