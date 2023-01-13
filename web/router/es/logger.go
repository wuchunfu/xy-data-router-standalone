package es

import (
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/service/schema"
)

type logData struct {
	*params
	*result
}

func log(params *params, ret *result) {
	res := &logData{params, ret}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)

	if err := json.NewEncoder(buf).Encode(res); err != nil {
		return
	}

	// 查询语句转换为字符串
	data := buf.Bytes()
	data, _ = sjson.SetBytes(data, "body", gjson.GetBytes(data, "body").String())

	item := schema.New(conf.Config.LogConf.ESIndex, common.ExternalIPv4, data)
	schema.PushDataToChanx(item)
	if conf.Debug {
		common.Log.Debug().RawJSON("query", buf.Bytes()).Msg("es query")
	}
}
