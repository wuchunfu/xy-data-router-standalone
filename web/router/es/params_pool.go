package es

import (
	"sync"
)

type params struct {
	Index      string         `json:"index"`
	DocumentID string         `json:"document_id"`
	Scroll     int            `json:"scroll"`
	ScrollID   string         `json:"scroll_id"`
	Body       map[string]any `json:"body"`
	ClientIP   string         `json:"client_ip"`
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
	p.Body = nil
	p.ClientIP = ""
	p.Scroll = 0
	p.ScrollID = ""
	paramsPool.Put(p)
}
