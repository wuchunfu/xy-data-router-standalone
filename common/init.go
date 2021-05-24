package common

import (
	"context"
	"time"

	"github.com/panjf2000/gnet/pool/goroutine"
	"github.com/tidwall/gjson"
)

var (
	IPv4Zero = "0.0.0.0"
	CtxBG    = context.Background()

	// 同步时间差值
	GTimeSub time.Duration

	// 协程池
	Pool = goroutine.Default()
)

func PoolRelease() {
	Pool.Release()
}

// 统一时间: TODO: 原子钟
func GetGlobalTime() time.Time {
	return time.Now().Add(GTimeSub)
}

// 统一时间并格式化
func GetGlobalDataTime(layout string) string {
	if layout == "" {
		layout = time.RFC3339
	}

	return GetGlobalTime().Format(layout)
}

// 检查必有字段: 只要存在该字段即可, 值可为空
func CheckRequiredField(body *[]byte, fields []string) bool {
	for _, field := range fields {
		if !gjson.GetBytes(*body, field).Exists() {
			return false
		}
	}

	return true
}
