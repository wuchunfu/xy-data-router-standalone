package common

import (
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
)

var (
	// GoPool 协程池
	GoPool = goroutine.Default()
)

func PoolRelease() {
	GoPool.Release()
}
