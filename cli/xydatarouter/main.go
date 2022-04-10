package main

import (
	"flag"
	"fmt"

	"github.com/fufuok/utils/xdaemon"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/master"
)

var version bool

func init() {
	flag.StringVar(&conf.ConfigFile, "c", conf.ConfigFile, "配置文件绝对路径(可选)")
	flag.StringVar(&conf.ForwardHost, "f", "", "指定上联服务器地址(可选) IP / 域名, 如: -f=hk.upstream.cn")
	flag.BoolVar(&version, "v", false, "版本信息")
	flag.Parse()
}

func main() {
	if version {
		fmt.Println(">>>", conf.APPName, conf.Version, conf.GoVersion)
		fmt.Println(">>>", conf.GitCommit)
		return
	}

	conf.InitConfig()

	if !conf.Debug {
		xdaemon.NewDaemon(conf.LogDaemon).Run()
	}

	defer master.Stop()
	master.Start()
	master.Watcher()

	select {}
}
