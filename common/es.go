package common

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v6"

	"github.com/fufuok/xy-data-router/conf"
)

var ES *elasticsearch.Client

func initES() {
	// 首次初始化 ES 连接, 连接失败时允许启动程序
	es, cfgErr, esErr := newES()
	if cfgErr != nil {
		log.Fatalln("Failed to initialize ES:", cfgErr, "\nbye.")
	}
	if esErr != nil {
		Log.Error().Err(esErr).Msg("es.Ping")
	}

	ES = es
}

// InitES 重新初始化 ES 连接, 成功则更新连接
func InitES() error {
	es, cfgErr, esErr := newES()
	if cfgErr != nil || esErr != nil {
		return fmt.Errorf("%s%s", cfgErr, esErr)
	}

	ES = es

	return nil
}

func newES() (es *elasticsearch.Client, cfgErr error, esErr error) {
	es, cfgErr = elasticsearch.NewClient(elasticsearch.Config{
		Addresses:    conf.Config.SYSConf.ESAddress,
		DisableRetry: !conf.Config.SYSConf.ESEnableRetry,
	})
	if cfgErr != nil {
		return nil, cfgErr, nil
	}

	// 数据转发时不涉及 ES
	if conf.ForwardTunnel != "" {
		return
	}

	_, esErr = es.Ping()

	return
}
