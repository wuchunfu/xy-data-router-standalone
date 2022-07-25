package datarouter

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/conf"
)

// 测试 PushDataToChanx 时加载接口配置的场景
func BenchmarkDataRouterLoad(b *testing.B) {
	// 模拟 Config 构建以接口名为键的配置集合
	apiNamePrefix := "TestAPI.Name."
	apiConfig := make(map[string]*conf.TAPIConf)
	for i := 0; i < 5000; i++ {
		apiname := apiNamePrefix + strconv.Itoa(i)
		apiConfig[apiname] = &conf.TAPIConf{
			APIName:       apiname,
			ESIndex:       apiname,
			ESIndexSplit:  "day",
			RequiredField: []string{"timestamp", "name", "msg", "more"},
			PostAPI: conf.TPostAPIConf{
				API: []string{"https://test.demo.com/v1/apiname",
					"http://localhost/api",
				},
			},
		}
	}

	// 原子存放方案
	var av atomic.Value
	type avDataRouter map[string]*tDataRouter
	avmap := make(avDataRouter)

	var sm sync.Map

	// 模拟 initDataRouter 初始化数据分发处理器
	for apiname, cfg := range apiConfig {
		apiConf := cfg
		apiConf.ESBulkHeader = []byte(`{"index":{"_index":"` + apiname + `","_type":"_doc"}}`)
		dr := &tDataRouter{
			apiConf: apiConf,
		}
		dataRouters.Store(apiname, dr)

		avmap[apiname] = dr

		sm.Store(apiname, dr)
	}

	av.Store(avmap)

	apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))

	b.ReportAllocs()
	b.ResetTimer()
	b.Run("xsync.Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
			dr, ok := dataRouters.Load(apiname)
			if !ok {
				b.Fatal("expected ok")
			}
			if dr.(*tDataRouter).apiConf.APIName != apiname {
				b.Fatalf("expected apiname is %s", apiname)
			}
		}
	})
	b.Run("sync.Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
			dr, ok := sm.Load(apiname)
			if !ok {
				b.Fatal("expected ok")
			}
			if dr.(*tDataRouter).apiConf.APIName != apiname {
				b.Fatalf("expected apiname is %s", apiname)
			}
		}
	})
	b.Run("av.Load", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
			m := av.Load().(avDataRouter)
			dr, ok := m[apiname]
			if !ok {
				b.Fatal("expected ok")
			}
			if dr.apiConf.APIName != apiname {
				b.Fatalf("expected apiname is %s", apiname)
			}
		}
	})
	b.Run("xsync.Map.Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
				dr, ok := dataRouters.Load(apiname)
				if !ok {
					b.Fatal("expected ok")
				}
				if dr.(*tDataRouter).apiConf.APIName != apiname {
					b.Fatalf("expected apiname is %s", apiname)
				}
			}
		})
	})
	b.Run("sync.Map.Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
				dr, ok := sm.Load(apiname)
				if !ok {
					b.Fatal("expected ok")
				}
				if dr.(*tDataRouter).apiConf.APIName != apiname {
					b.Fatalf("expected apiname is %s", apiname)
				}
			}
		})
	})
	b.Run("av.Load.Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// apiname := apiNamePrefix + strconv.Itoa(utils.FastIntn(5000))
				m := av.Load().(avDataRouter)
				dr, ok := m[apiname]
				if !ok {
					b.Fatal("expected ok")
				}
				if dr.apiConf.APIName != apiname {
					b.Fatalf("expected apiname is %s", apiname)
				}
			}
		})
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkDataRouterLoad
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/service
// cpu: Intel(R) Xeon(R) CPU E5-2667 v2 @ 3.30GHz
// BenchmarkDataRouterLoad/xsync.Map-4              46649872               25.96 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/xsync.Map-4              45526196               33.19 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/sync.Map-4               29399697               51.21 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/sync.Map-4               30169599               51.02 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/av.Load-4                46885378               25.35 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/av.Load-4                47382720               25.39 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/xsync.Map.Parallel-4    180850470               6.629 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/xsync.Map.Parallel-4    176905102               6.658 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/sync.Map.Parallel-4     100000000               10.34 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/sync.Map.Parallel-4     100000000               10.24 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/av.Load.Parallel-4      179036335               6.553 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad/av.Load.Parallel-4      177899280               6.798 ns/op            0 B/op          0 allocs/op
