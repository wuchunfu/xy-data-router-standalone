package main

import (
	"flag"
	"fmt"

	"github.com/fufuok/utils/xdaemon"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/master"
	"github.com/fufuok/xy-data-router/service/es"
)

var version, daemon bool

func init() {
	flag.StringVar(&conf.ForwardHost, "f", "", "指定上联服务器地址(可选) IP / 域名, 如: -f=hk.upstream.cn")
	flag.StringVar(&conf.ConfigFile, "c", conf.ConfigFile, "配置文件绝对路径(可选)")
	flag.BoolVar(&conf.Debug, "debug", false, "指定该参数进入调试模式, 日志级别 Debug")
	flag.BoolVar(&daemon, "d", false, "启动后台守护进程")
	flag.BoolVar(&version, "v", false, "版本信息")
	flag.Parse()
}

func main() {
	if version {
		fmt.Println(">>>", conf.APPName, conf.Version, conf.GoVersion)
		fmt.Println(">>>", conf.GitCommit)
		fmt.Println(">>> Elasticsearch client version:", es.ClientVer)
		return
	}

	conf.InitMain()

	if daemon || !conf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	defer master.Stop()
	master.Start()
	master.Watcher()

	select {}
}
