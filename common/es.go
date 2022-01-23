package common

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v6"
	"github.com/tidwall/gjson"
	"github.com/valyala/bytebufferpool"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	ES *elasticsearch.Client

	// ESVersionServer ESVersionClient 版本信息
	ESVersionServer string
	ESVersionClient string
)

func initES() {
	// 首次初始化 ES 连接, PING 失败时允许启动程序
	es, cfgErr, _ := newES()
	if cfgErr != nil {
		log.Fatalln("Failed to initialize ES:", cfgErr, "\nbye.")
	}

	ES = es
}

// InitES 重新初始化 ES 连接, PING 成功则更新连接
func InitES() error {
	es, cfgErr, esErr := newES()
	if cfgErr != nil || esErr != nil {
		return fmt.Errorf("%s%s", cfgErr, esErr)
	}

	ES = es

	return nil
}

func newES() (es *elasticsearch.Client, cfgErr error, esErr error) {
	Log.Info().Strs("hosts", conf.Config.SYSConf.ESAddress).Msg("Initialize ES connection")
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

	res, err := es.Info()
	if err != nil {
		return nil, nil, err
	}
	if res.IsError() {
		err = fmt.Errorf("ES info error, status: %s", res.Status())
		Log.Error().Err(err).Msg("es.Info")
		return nil, nil, err
	}

	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)
	n, _ := bb.ReadFrom(res.Body)
	if n == 0 {
		err = fmt.Errorf("ES info error: nil")
		Log.Error().Err(err).Msg("es.Info")
		return nil, nil, err
	}
	ESVersionServer = gjson.GetBytes(bb.Bytes(), "version.number").String()
	ESVersionClient = elasticsearch.Version

	return
}
