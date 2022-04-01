package es

import (
	"sync"

	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/fufuok/xy-data-router/common"
)

type tResponse struct {
	response  *esapi.Response
	err       error
	totalPath string
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return new(tResponse)
	},
}

func getResponse() *tResponse {
	resp := responsePool.Get().(*tResponse)
	if common.ESLessThan7 {
		resp.totalPath = "hits.total"
	} else {
		resp.totalPath = "hits.total.value"
	}
	return resp
}

func putResponse(r *tResponse) {
	r.response = nil
	r.err = nil
	r.totalPath = ""
	responsePool.Put(r)
}
