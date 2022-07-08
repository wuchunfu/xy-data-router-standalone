package common

import (
	"github.com/imroc/req/v3"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
)

var (
	// ReqUpload HTTP 文件上传客户端 (调试模式不显示上传文件内容, 无超时时间)
	ReqUpload *req.Client

	// ReqDownload HTTP 文件下载客户端 (调试模式不显示下载文件内容, 无超时时间)
	ReqDownload *req.Client

	// HTTP 客户端调试模式
	reqDebug bool
)

func initReq() {
	newReq()
	loadReq()
}

func loadReq() {
	req.SetTimeout(conf.Config.DataConf.APIClientTimeoutDuration)
	if reqDebug == conf.Debug {
		return
	}
	reqDebug = conf.Debug
	req.SetLogger(NewAppLogger())
	ReqUpload.SetLogger(NewAppLogger())
	ReqDownload.SetLogger(NewAppLogger())
	if reqDebug {
		req.EnableDumpAll().EnableDebugLog().EnableTraceAll()
		ReqUpload.EnableDumpAllWithoutRequestBody().EnableDebugLog().EnableTraceAll()
		ReqDownload.EnableDumpAllWithoutResponseBody().EnableDebugLog().EnableTraceAll()
	} else {
		req.DisableDumpAll().DisableDebugLog().DisableTraceAll()
		ReqUpload.DisableDumpAll().DisableDebugLog().DisableTraceAll()
		ReqDownload.DisableDumpAll().DisableDebugLog().DisableTraceAll()
	}
}

func newReq() {
	req.SetUserAgent(conf.ReqUserAgent).
		SetJsonMarshal(json.Marshal).
		SetJsonUnmarshal(json.Unmarshal).
		SetLogger(NewAppLogger())
	ReqUpload = req.C().
		SetUserAgent(conf.ReqUserAgent).
		SetJsonMarshal(json.Marshal).
		SetJsonUnmarshal(json.Unmarshal).
		SetLogger(NewAppLogger())
	ReqDownload = req.C().
		SetUserAgent(conf.ReqUserAgent).
		SetJsonMarshal(json.Marshal).
		SetJsonUnmarshal(json.Unmarshal).
		SetLogger(NewAppLogger())
}
