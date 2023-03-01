package common

import (
	"time"

	"github.com/fufuok/timewheel"
	"github.com/fufuok/utils/myip"
	"github.com/fufuok/utils/xjson/gjson"
	"github.com/fufuok/utils/xsync"
)

var (
	// GTimeSub TODO: 同步时间差值
	GTimeSub time.Duration

	// TWms 时间轮, 精度 50ms, 1s, 1m
	TWms *timewheel.TimeWheel
	TWs  *timewheel.TimeWheel
	TWm  *timewheel.TimeWheel

	// Start 系统启动时间
	Start = time.Now()

	Now = &CurrentTime{
		Start.Format("2006-01-02T15:04:05Z"),
		Start.Format(time.RFC3339),
		Start,
	}

	IPv4Zero = "0.0.0.0"

	// HTTPRequestCount HTTP 请求计数
	HTTPRequestCount    xsync.Counter
	HTTPBadRequestCount xsync.Counter

	// InternalIPv4 服务器 IP
	InternalIPv4 string
	ExternalIPv4 string
)

// CurrentTime 当前时间, 预格式化的字符串形式 (秒级)
type CurrentTime struct {
	// 强制 0 时区的 8 时区时间 (仅特定场景展示使用)
	Str3339Z string
	// 带时区的正确时间值
	Str3339 string
	// 当前时间
	Time time.Time
}

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
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	if err := loadLogger(); err != nil {
		Log.Error().Err(err).Msg("Unable to reinitialize logger")
	}
	loadReq()
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

// 周期性更新全局时间字段
func syncNow() {
	ticker := TWms.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		t := GTimeNow()
		c := &CurrentTime{
			t.Format("2006-01-02T15:04:05Z"),
			t.Format(time.RFC3339),
			t,
		}
		Now = c
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
