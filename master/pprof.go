package master

import (
	"net/http"
	"time"

	"github.com/arl/statsviz"

	"github.com/fufuok/xy-data-router/conf"
)

// 统计和性能工具
func startPProf() {
	if conf.Config.SYSConf.PProfAddr != "" {
		go func() {
			_ = statsviz.RegisterDefault(statsviz.SendFrequency(time.Second * 5))
			_ = http.ListenAndServe(conf.Config.SYSConf.PProfAddr, nil)
		}()
	}
}
