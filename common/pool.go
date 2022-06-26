package common

import (
	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	// GoPool 协程池
	GoPool = goroutine.Default()
)

func initPool() {
	bufferpool.SetMaxSize(conf.BufferMaxCapacity)
	bytespool.InitDefaultPools(2, conf.BufferMaxCapacity)
}

func poolRelease() {
	GoPool.Release()
}
