package es

import (
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/schema"
	"github.com/fufuok/xy-data-router/service"
)

type tLog struct {
	*tParams
	*tResult
}

func log(params *tParams, ret *tResult) {
	res := &tLog{params, ret}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	if err := json.NewEncoder(buf).Encode(res); err != nil {
		return
	}

	// 查询语句转换为字符串
	data := buf.Bytes()
	data, _ = sjson.SetBytes(data, "body", gjson.GetBytes(data, "body").String())

	item := schema.New(conf.Config.SYSConf.ESAPILogIndex, service.ExternalIPv4, data)
	service.PushDataToChanx(item)
	if conf.Debug {
		common.Log.Debug().RawJSON("query", buf.Bytes()).Msg("es query")
	}
}
