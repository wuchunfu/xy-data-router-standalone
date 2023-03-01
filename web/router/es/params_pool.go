package es

import (
	"sync"

	"github.com/fufuok/utils"
)

type params struct {
	Index      string         `json:"index"`
	DocumentID string         `json:"document_id"`
	Scroll     int            `json:"scroll"`
	ScrollID   string         `json:"scroll_id"`
	Body       map[string]any `json:"body"`
	ClientIP   string         `json:"client_ip"`

	// ES 更新/删除时可选指定是否刷新数据, 只允许 ?refresh=true 或默认
	Refresh string `json:"refresh"`

	// 请求 esapi 接口超时毫秒数, 用于提前返回响应, 会自动转换为 ?timeout=99ms
	TimeoutMs int `json:"timeout_ms"`
}

var paramsPool = sync.Pool{
	New: func() any {
		return new(params)
	},
}

func getParams() *params {
	return paramsPool.Get().(*params)
}

func putParams(p *params) {
	p.Index = ""
	p.Scroll = 0
	p.ScrollID = ""
	p.ClientIP = ""
	p.Body = nil
	p.Refresh = ""
	p.TimeoutMs = 0
	paramsPool.Put(p)
}

func fixedRefresh(s string) string {
	s = utils.ToLower(s)
	if s != "true" {
		s = ""
	}
	return s
}
