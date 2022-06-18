package common

import (
	"time"

	"github.com/fufuok/timewheel"
	"github.com/tidwall/gjson"
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
)

func InitCommon() {
	// 180 秒
	TWms, _ = timewheel.NewTimeWheel(100*time.Millisecond, 1800)
	TWms.Start()

	// 30 分
	TWs, _ = timewheel.NewTimeWheel(time.Second, 1800)
	TWs.Start()

	// 24 时
	TWm, _ = timewheel.NewTimeWheel(time.Minute, 1440)
	TWm.Start()

	// 池相关设置
	initPool()

	// 同步时间字段
	go syncNow()

	// 初始化日志环境
	initLogger()

	// 初始化 HTTP 客户端连接配置
	initReq()

	// 初始化代理参数
	initProxy()

	// 初始化 ES 连接
	initES()
}

func TWStop() {
	TWms.Stop()
	TWs.Stop()
	TWm.Stop()
}

// GetGlobalTime 统一时间
func GetGlobalTime() time.Time {
	return time.Now().Add(GTimeSub)
}

// GetGlobalDataTime 统一时间并格式化
func GetGlobalDataTime(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return GetGlobalTime().Format(layout)
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
		now := GetGlobalTime()
		Now3399UTC = now.Format("2006-01-02T15:04:05Z")
		Now3399 = now.Format(time.RFC3339)
	}
}
