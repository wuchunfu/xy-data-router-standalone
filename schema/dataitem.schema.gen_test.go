package schema

import (
	"testing"

	"github.com/fufuok/utils"
)

func TestDataItem_Marshal(t *testing.T) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	dec, err := data.Marshal(nil)
	utils.AssertEqual(t, true, err == nil)

	item := Make()
	_, err = item.Unmarshal(dec)
	utils.AssertEqual(t, true, err == nil)
	utils.AssertEqual(t, item, data)
}

func BenchmarkDataItem_Marshal(b *testing.B) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = data.Marshal(nil)
	}
}

func BenchmarkDataItem_Unmarshal(b *testing.B) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	dec, _ := data.Marshal(nil)
	item := Make()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = item.Unmarshal(dec)
	}
}

func BenchmarkDataItem_Marshal_Parallel(b *testing.B) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = data.Marshal(nil)
		}
	})
}

func BenchmarkDataItem_Unmarshal_Parallel(b *testing.B) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	dec, _ := data.Marshal(nil)
	item := Make()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = item.Unmarshal(dec)
		}
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkDataItem_
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/schema
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkDataItem_Marshal-4                      8294629               136.6 ns/op           576 B/op          1 allocs/op
// BenchmarkDataItem_Marshal-4                      9026492               135.6 ns/op           576 B/op          1 allocs/op
// BenchmarkDataItem_Unmarshal-4                   18264658                65.75 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItem_Unmarshal-4                   18336759                65.68 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItem_Marshal_Parallel-4            12356250               100.3 ns/op           576 B/op          1 allocs/op
// BenchmarkDataItem_Marshal_Parallel-4            12638331               101.1 ns/op           576 B/op          1 allocs/op
// BenchmarkDataItem_Unmarshal_Parallel-4           7288165               165.4 ns/op            16 B/op          2 allocs/op
// BenchmarkDataItem_Unmarshal_Parallel-4           7345234               165.7 ns/op            16 B/op          2 allocs/op
