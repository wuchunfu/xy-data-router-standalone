package es

import (
	"sync"
)

type result struct {
	result     []byte
	StatusCode int    `json:"status_code"`
	Took       int64  `json:"took"`
	Count      int    `json:"count"`
	Error      string `json:"err"`
	ErrMsg     string `json:"err_msg"`
	ReqUri     string `json:"req_uri"`
	ReqTime    string `json:"req_time"`
	ReqType    string `json:"type"`
}

var resultPool = sync.Pool{
	New: func() any {
		return new(result)
	},
}

func getResult() *result {
	return resultPool.Get().(*result)
}

func putResult(r *result) {
	r.result = r.result[:0]
	r.Count = 0
	r.Error = ""
	r.ErrMsg = ""
	r.ReqUri = ""
	resultPool.Put(r)
}
