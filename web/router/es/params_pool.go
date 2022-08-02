package es

import (
	"sync"
)

type tParams struct {
	Index    string         `json:"index"`
	Scroll   int            `json:"scroll"`
	ScrollID string         `json:"scroll_id"`
	Body     map[string]any `json:"body"`
	ClientIP string         `json:"client_ip"`
}

var paramsPool = sync.Pool{
	New: func() any {
		return new(tParams)
	},
}

func getParams() *tParams {
	return paramsPool.Get().(*tParams)
}

func putParams(p *tParams) {
	p.Index = ""
	p.Body = nil
	p.ClientIP = ""
	p.Scroll = 0
	p.ScrollID = ""
	paramsPool.Put(p)
}
