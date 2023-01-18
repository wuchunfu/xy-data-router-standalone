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
	Refresh    string         `json:"refresh"`
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
	paramsPool.Put(p)
}

// 只允许 ?refresh=true 或默认
func fixedRefresh(s string) string {
	s = utils.ToLower(s)
	if s != "true" {
		s = ""
	}
	return s
}
