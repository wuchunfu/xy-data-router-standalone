package common

import (
	"github.com/imroc/req/v3"

	"github.com/fufuok/xy-data-router/conf"
)

func InitReq() {
	if conf.Debug {
		req.EnableDumpAll().EnableDebugLog().EnableTraceAll()
	} else {
		req.DisableDumpAll().DisableDebugLog().DisableTraceAll()
	}
	req.SetTimeout(conf.Config.DataConf.APIClientTimeoutDuration).SetUserAgent(conf.ReqUserAgent)
}
