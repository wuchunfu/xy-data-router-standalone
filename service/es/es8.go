//go:build !es7

package es

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	Client *elasticsearch.Client

	// ClientVer 客户端版本
	ClientVer = elasticsearch.Version
)

type tESClient struct {
	client *elasticsearch.Client
}

// TResponse 通用请求响应体
type TResponse struct {
	Response  *esapi.Response
	TotalPath string
	Err       error
}

func newES() (client tESClient, cfgErr error, esErr error) {
	common.Log.Info().Strs("hosts", conf.Config.DataConf.ESAddress).Msg("Initialize ES connection")
	client.client, cfgErr = elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     conf.Config.DataConf.ESAddress,
		RetryOnStatus: conf.Config.DataConf.ESRetryOnStatus,
		MaxRetries:    conf.Config.DataConf.ESMaxRetries,
		DisableRetry:  conf.Config.DataConf.ESDisableRetry,
		Transport:     &transport{},
	})
	if cfgErr != nil {
		return
	}

	esErr = parseVersion(client)
	return
}
