package schema

import (
	"testing"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"
)

func TestDataItem(t *testing.T) {
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
	buf := bytespool.New64(data.Size())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = data.Marshal(buf)
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
		buf := bytespool.New64(data.Size())
		for pb.Next() {
			_, _ = data.Marshal(buf)
		}
	})
}

func BenchmarkDataItem_Unmarshal_Parallel(b *testing.B) {
	data := New("test", "7.7.7.7", utils.FastRandBytes(512))
	dec, _ := data.Marshal(nil)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			item := Make()
			_, _ = item.Unmarshal(dec)
			item.Release()
		}
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=BenchmarkDataItem_
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/schema
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkDataItem_Marshal-4                     40991911                29.34 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItem_Marshal-4                     41093400                29.39 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItem_Unmarshal-4                   17382360                69.21 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItem_Unmarshal-4                   17725569                70.28 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItem_Marshal_Parallel-4           152554408                7.997 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItem_Marshal_Parallel-4           151097313                7.862 ns/op            0 B/op          0 allocs/op
// BenchmarkDataItem_Unmarshal_Parallel-4          35022432                33.98 ns/op           16 B/op          2 allocs/op
// BenchmarkDataItem_Unmarshal_Parallel-4          34755963                34.48 ns/op           16 B/op          2 allocs/op
