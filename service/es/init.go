package es

import (
	"github.com/fufuok/xy-data-router/common"
)

// InitMain 程序启动时初始化配置
func InitMain() {
	initES()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	if err := loadES(); err != nil {
		common.Log.Error().Err(err).Msg("Failed to update elasticsearch connection")
	}
}

func Stop() {}
