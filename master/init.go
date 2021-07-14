package master

import (
	"context"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/controller"
	"github.com/fufuok/xy-data-router/service"
	"github.com/fufuok/xy-data-router/tunnel"
)

var (
	// 重启信号
	restartChan = make(chan bool)

	// 配置重载信息
	reloadChan = make(chan bool)
)

func Start() {
	// 初始化
	initMaster()

	go func() {
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

func initMaster() {
	// 优先初始化公共变量
	common.InitCommon()

	// 启动数据服务
	service.InitService()

	// 启动 Tunnel 服务
	go tunnel.InitTunServer()
	go tunnel.InitTunClient()

	// 启动 Web 服务
	go controller.InitWebServer()

	// 统计和性能工具
	go startPProf()
}

// 程序退出时清理
func Stop() {
	common.TWStop()
	common.PoolRelease()
	service.PoolRelease()
}
