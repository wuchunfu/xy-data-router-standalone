package es

import (
	"sync"
)

var responsePool = sync.Pool{
	New: func() any {
		return new(Response)
	},
}

func GetResponse() *Response {
	resp := responsePool.Get().(*Response)
	if ServerLessThan7 {
		resp.TotalPath = "hits.total"
	} else {
		resp.TotalPath = "hits.total.value"
	}
	return resp
}

func PutResponse(r *Response) {
	r.Response = nil
	r.Err = nil
	r.TotalPath = ""
	responsePool.Put(r)
}
