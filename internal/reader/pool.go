package reader

import (
	"bytes"
	"sync"
)

var defaultPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewReader(nil)
	},
}

func New(b []byte) *bytes.Reader {
	r := defaultPool.Get().(*bytes.Reader)
	r.Reset(b)
	return r
}

func Release(r *bytes.Reader) {
	r.Reset(nil)
	defaultPool.Put(r)
}
