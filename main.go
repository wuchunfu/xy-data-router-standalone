package main

import (
	"github.com/zh-five/xdaemon"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/master"
)

func main() {
	defer master.Stop()

	if !conf.Config.SYSConf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	master.Start()
	master.Watcher()

	select {}
}
