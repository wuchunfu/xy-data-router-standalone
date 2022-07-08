package service

import (
	"github.com/fufuok/xy-data-router/service/datarouter"
	"github.com/fufuok/xy-data-router/service/schema"
	"github.com/fufuok/xy-data-router/service/tunnel"
)

// InitMain 程序启动时初始化
func InitMain() {
	// 启动数据收集服务
	schema.InitMain()

	// 启动数据分发服务
	datarouter.InitMain()

	// 启动 Tunnel 服务
	tunnel.InitMain()

	// 心跳服务
	go initHeartbeat()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	datarouter.InitRuntime()
	tunnel.InitRuntime()
	schema.InitRuntime()
}

func Stop() {
	datarouter.Stop()
	tunnel.Stop()
	schema.Stop()
}
