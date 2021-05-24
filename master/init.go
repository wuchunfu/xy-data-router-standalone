package master

import (
	"context"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/controller"
	"github.com/fufuok/xy-data-router/service"
)

var (
	// 重启信号
	restartChan = make(chan bool)

	// 配置重载信息
	reloadChan = make(chan bool)
)

func Start() {
	go func() {
		// 启动 Web 服务
		go controller.InitWebServer()

		// 统计和性能工具
		go startPProf()

		for {
			cancelCtx, cancel := context.WithCancel(common.CtxBG)
			ctx := context.WithValue(cancelCtx, "start", time.Now())

			// 获取远程配置
			go startRemoteConf(ctx)

			select {
			case <-restartChan:
				// 强制退出, 由 Daemon 重启程序
				common.Log.Warn().Msg("restart <-restartChan")
				os.Exit(0)
			case <-reloadChan:
				// 重载配置及相关服务
				cancel()
				common.Log.Warn().Msg("reload <-reloadChan")
			}
		}
	}()
}

// 程序退出时清理
func Stop() {
	common.PoolRelease()
	service.PoolRelease()
}
