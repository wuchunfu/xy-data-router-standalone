//go:build !es7

package es

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/logger"
)

var (
	Client *elasticsearch.Client

	// ClientVer 客户端版本
	ClientVer = elasticsearch.Version
)

type esClient struct {
	client *elasticsearch.Client
}

// Response 通用请求响应体
type Response struct {
	Response  *esapi.Response
	TotalPath string
	Err       error
}

func newES() (client esClient, cfgErr error, esErr error) {
	logger.Info().Strs("hosts", conf.Config.DataConf.ESAddress).Msg("Initialize ES connection")
	cfg := elasticsearch.Config{
		Addresses:     conf.Config.DataConf.ESAddress,
		RetryOnStatus: conf.Config.DataConf.ESRetryOnStatus,
		MaxRetries:    conf.Config.DataConf.ESMaxRetries,
		DisableRetry:  conf.Config.DataConf.ESDisableRetry,
		Transport:     &transport{},
	}
	if conf.Config.DataConf.ESAuthPassword != "" {
		cfg.Username = conf.ProjectName
		cfg.Password = conf.Config.DataConf.ESAuthPassword
	}
	client.client, cfgErr = elasticsearch.NewClient(cfg)
	if cfgErr != nil {
		return
	}

	esErr = parseVersion(client)
	return
}
