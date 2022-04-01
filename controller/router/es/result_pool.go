package es

import (
	"sync"
)

type tResult struct {
	result []byte
	err    error
	count  int
	errMsg string
}

var resultPool = sync.Pool{
	New: func() interface{} {
		return new(tResult)
	},
}

func getResult() *tResult {
	return resultPool.Get().(*tResult)
}

func putResult(r *tResult) {
	r.result = r.result[:0]
	r.err = nil
	r.count = 0
	r.errMsg = ""
	resultPool.Put(r)
}
