package master

import (
	"net/http"

	"github.com/fufuok/xy-data-router/conf"
)

// 统计和性能工具
func startPProf() {
	if conf.Config.SYSConf.PProfAddr != "" {
		_ = http.ListenAndServe(conf.Config.SYSConf.PProfAddr, nil)
	}
}
