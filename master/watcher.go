package master

import (
	"log"
	"time"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger"
	"github.com/fufuok/xy-data-router/service"
	"github.com/fufuok/xy-data-router/web"
)

// Watcher 监听程序二进制变化(重启)和配置文件(热加载)
func Watcher() {
	mainFile := utils.Executable(true)
	if mainFile == "" {
		log.Fatalln("Failed to initialize Watcher: miss executable", "\nbye.")
	}

	interval := conf.Config.SYSConf.WatcherIntervalDuration
	intervalStr := interval.String()
	md5Main := utils.MustMD5Sum(mainFile)
	md5Conf := conf.GetFileVer(conf.ConfigFile).MD5
	logger.Info().Str("main", mainFile).Str("interval", intervalStr).Msg("Watching")
	logger.Info().Str("config", conf.ConfigFile).Str("interval", intervalStr).Msg("Watching")

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			// 程序二进制变化时重启
			md5New := utils.MustMD5Sum(mainFile)
			if md5New != md5Main {
				md5Main = md5New
				logger.Warn().Msg(">>>>>>> restart main <<<<<<<")
				restartChan <- true
				continue
			}
			// 配置文件变化时热加载
			md5New = utils.MustMD5Sum(conf.ConfigFile)
			if md5New != md5Conf {
				md5Conf = md5New
				if err := conf.LoadConf(); err != nil {
					logger.Error().Err(err).Msg("reload config")
					continue
				}

				// 重启程序指令
				if conf.Config.SYSConf.RestartMain {
					logger.Warn().Msg(">>>>>>> restart main(config) <<<<<<<")
					restartChan <- true
					continue
				}

				common.InitRuntime()
				service.InitRuntime()
				web.InitRuntime()

				// 更新配置文件监控周期
				if interval != conf.Config.SYSConf.WatcherIntervalDuration {
					interval = conf.Config.SYSConf.WatcherIntervalDuration
					ticker.Reset(interval)
					logger.Warn().Str("interval", interval.String()).Msg("reset ticker")
				}

				logger.Warn().Msg(">>>>>>> reload config <<<<<<<")
				reloadChan <- true
			}
		}
	}()
}
