package es

import (
	"sync"
)

var responsePool = sync.Pool{
	New: func() any {
		return new(TResponse)
	},
}

func GetResponse() *TResponse {
	resp := responsePool.Get().(*TResponse)
	if ServerLessThan7 {
		resp.TotalPath = "hits.total"
	} else {
		resp.TotalPath = "hits.total.value"
	}
	return resp
}

func PutResponse(r *TResponse) {
	r.Response = nil
	r.Err = nil
	r.TotalPath = ""
	responsePool.Put(r)
}
