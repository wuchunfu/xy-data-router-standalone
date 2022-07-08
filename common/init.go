package common

import (
	"time"

	"github.com/fufuok/chanx"
	"github.com/fufuok/timewheel"
	"github.com/fufuok/utils/myip"
	"github.com/fufuok/utils/xsync"
	"github.com/tidwall/gjson"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	// GTimeSub TODO: 同步时间差值
	GTimeSub time.Duration

	// TWms 时间轮, 精度 50ms, 1s, 1m
	TWms *timewheel.TimeWheel
	TWs  *timewheel.TimeWheel
	TWm  *timewheel.TimeWheel

	// Now3399UTC 当前时间
	Now3399UTC = time.Now().Format("2006-01-02T15:04:05Z")
	Now3399    = time.Now().Format(time.RFC3339)

	IPv4Zero = "0.0.0.0"

	// HTTPRequestCount HTTP 请求计数
	HTTPRequestCount    xsync.Counter
	HTTPBadRequestCount xsync.Counter

	// Start 系统启动时间
	Start = time.Now()

	// InternalIPv4 服务器 IP
	InternalIPv4 string
	ExternalIPv4 string
)

// InitMain 程序启动时初始化
func InitMain() {
	// 180 秒
	TWms, _ = timewheel.NewTimeWheel(100*time.Millisecond, 1800)
	TWms.Start()

	// 30 分
	TWs, _ = timewheel.NewTimeWheel(time.Second, 1800)
	TWs.Start()

	// 24 时
	TWm, _ = timewheel.NewTimeWheel(time.Minute, 1440)
	TWm.Start()

	// 同步时间字段
	go syncNow()

	// 初始化本机 IP
	go initServerIP()

	// 初始化日志环境
	initLogger()

	// 池相关设置
	initPool()

	// 初始化 HTTP 客户端连接配置
	initReq()

	// 初始化代理参数
	initProxy()

	// 初始化 ES 连接
	initES()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	loadLogger()
	loadReq()

	// 重新连接 ES
	if err := loadES(); err != nil {
		Log.Error().Err(err).Msg("Failed to update elasticsearch connection")
	}
}

func Stop() {
	twStop()
	poolRelease()
}

func twStop() {
	TWms.Stop()
	TWs.Stop()
	TWm.Stop()
}

// GTimeNow 统一时间
func GTimeNow() time.Time {
	return time.Now().Add(GTimeSub)
}

// GTimeNowString 统一时间并格式化
func GTimeNowString(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}
	return GTimeNow().Format(layout)
}

// CheckRequiredField 检查必有字段: 只要存在该字段即可, 值可为空
func CheckRequiredField(body []byte, fields []string) bool {
	for _, field := range fields {
		if !gjson.GetBytes(body, field).Exists() {
			return false
		}
	}
	return true
}

// NewChanx 初始化无限缓冲信道
func NewChanx() *chanx.UnboundedChan {
	return chanx.NewUnboundedChan(conf.Config.DataConf.ChanSize, conf.Config.DataConf.ChanMaxBufCap)
}

// 周期性更新全局时间字段
func syncNow() {
	ticker := TWms.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		now := GTimeNow()
		Now3399UTC = now.Format("2006-01-02T15:04:05Z")
		Now3399 = now.Format(time.RFC3339)
	}
}

func initServerIP() {
	go func() {
		InternalIPv4 = myip.InternalIPv4()
	}()
	go func() {
		ExternalIPv4 = myip.ExternalIPAny(10)
	}()
}
