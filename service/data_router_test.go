package service

import (
	"strconv"
	"testing"

	"github.com/fufuok/xy-data-router/conf"
)

// 测试 PushDataToChanx 时加载接口配置的场景
func BenchmarkDataRouterLoad(b *testing.B) {
	// 模拟 Config 构建以接口名为键的配置集合
	apiNamePrefix := "TestAPI.Name."
	apiConfig := make(map[string]*conf.TAPIConf)
	apiname := apiNamePrefix + "777"
	for i := 0; i < 5000; i++ {
		apiname = apiNamePrefix + strconv.Itoa(i)
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

	// 模拟 InitDataRouter 初始化数据分发处理器
	for apiname, cfg := range apiConfig {
		apiConf := cfg
		apiConf.ESBulkHeader = []byte(`{"index":{"_index":"` + apiname + `","_type":"_doc"}}`)
		dr := &tDataRouter{
			apiConf: apiConf,
			drOut:   &tDataRouterOut{},
		}
		dataRouters.Store(apiname, dr)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dr, ok := dataRouters.Load(apiname)
		if !ok {
			b.Fatal("expected ok")
		}
		if dr.(*tDataRouter).apiConf.APIName != apiname {
			b.Fatalf("expected apiname is %s", apiname)
		}
	}
}

// go test -bench=BenchmarkDataRouterLoad -benchtime=1s -count=3
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/service
// cpu: Intel(R) Xeon(R) CPU E5-2667 v2 @ 3.30GHz
// BenchmarkDataRouterLoad-4       55250872                21.60 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad-4       54071024                28.00 ns/op            0 B/op          0 allocs/op
// BenchmarkDataRouterLoad-4       55929714                21.69 ns/op            0 B/op          0 allocs/op
