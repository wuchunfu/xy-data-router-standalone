package common

import (
	"time"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/panjf2000/ants/v2"

	"github.com/fufuok/xy-data-router/conf"
)

const (
	// antsPoolSize sets up the capacity of worker pool, 256 * 1024.
	antsPoolSize = 1 << 18

	// expiryDuration is the interval time to clean up those expired workers.
	expiryDuration = 10 * time.Second
)

var (
	// GoPool 协程池
	GoPool *ants.Pool
)

func initPool() {
	bufferpool.SetMaxSize(conf.BufferMaxCapacity)
	bytespool.InitDefaultPools(2, conf.BufferMaxCapacity)
	GoPool, _ = ants.NewPool(
		antsPoolSize,
		ants.WithExpiryDuration(expiryDuration),
		ants.WithNonblocking(true),
		ants.WithLogger(NewAppLogger()),
	)
}

func poolRelease() {
	GoPool.Release()
}
