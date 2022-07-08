package tunnel

import (
	"testing"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/service/schema"
)

func TestCodec(t *testing.T) {
	coder := new(genCodec)
	src := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	dec, err := coder.Marshal(src)
	utils.AssertEqual(t, true, err == nil)
	// t.Log(len(dec), src.Size())

	item := schema.Make()
	err = coder.Unmarshal(dec, item)
	utils.AssertEqual(t, true, err == nil)
	utils.AssertEqual(t, item, src)
}

func TestCodecCompress(t *testing.T) {
	coder := new(genCodec)
	src := schema.New("test", "7.7.7.7", utils.FastRandBytes(1512))
	src.Flag = 1
	dec, err := coder.Marshal(src)
	// t.Log(len(dec), src.Size())
	utils.AssertEqual(t, true, err == nil)

	item := schema.Make()
	err = coder.Unmarshal(dec, item)
	utils.AssertEqual(t, true, err == nil)
	utils.AssertEqual(t, item, src)
}

func BenchmarkCodec_Marshal(b *testing.B) {
	var dec []byte
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec, _ = coder.Marshal(data)
		bytespool.Release(dec)
	}
}

func BenchmarkCodec_Unmarshal(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	dec, _ := coder.Marshal(data)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = coder.Unmarshal(dec, item)
	}
}

func BenchmarkCodec_Marshal_Compress(b *testing.B) {
	var dec []byte
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec, _ = coder.Marshal(data)
		bytespool.Release(dec)
	}
}

func BenchmarkCodec_Unmarshal_Compress(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	dec, _ := coder.Marshal(data)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = coder.Unmarshal(dec, item)
	}
}

func BenchmarkCodec_Marshal_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var dec []byte
		for pb.Next() {
			dec, _ = coder.Marshal(data)
			bytespool.Release(dec)
		}
	})
}

func BenchmarkCodec_Unmarshal_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		dec, _ := coder.Marshal(data)
		for pb.Next() {
			_ = coder.Unmarshal(dec, item)
		}
	})
}

func BenchmarkCodec_Marshal_Compress_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var dec []byte
		for pb.Next() {
			dec, _ = coder.Marshal(data)
			bytespool.Release(dec)
		}
	})
}

func BenchmarkCodec_Unmarshal_Compress_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		dec, _ := coder.Marshal(data)
		for pb.Next() {
			_ = coder.Unmarshal(dec, item)
		}
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=.
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/tunnel
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkCodec_Marshal-4                        10929298               110.2 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Marshal-4                        10828400               111.1 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Unmarshal-4                       9805401               120.5 ns/op            45 B/op          2 allocs/op
// BenchmarkCodec_Unmarshal-4                      10154618               121.8 ns/op            46 B/op          2 allocs/op
// BenchmarkCodec_Marshal_Compress-4                 139848                8582 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Marshal_Compress-4                 140238                8599 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Unmarshal_Compress-4              2446064               481.6 ns/op            79 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal_Compress-4              2520304               654.8 ns/op            84 B/op          1 allocs/op
// BenchmarkCodec_Marshal_Parallel-4               42292669               28.15 ns/op            0 B/op           0 allocs/op
// BenchmarkCodec_Marshal_Parallel-4               41839184               28.13 ns/op            0 B/op           0 allocs/op
// BenchmarkCodec_Unmarshal_Parallel-4              6645628               180.7 ns/op            35 B/op          2 allocs/op
// BenchmarkCodec_Unmarshal_Parallel-4              6321154               179.3 ns/op            36 B/op          2 allocs/op
// BenchmarkCodec_Marshal_Compress_Parallel-4        504944                2316 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Marshal_Compress_Parallel-4        503425                2374 ns/op             0 B/op          0 allocs/op
// BenchmarkCodec_Unmarshal_Compress_Parallel-4     7723346               178.5 ns/op            69 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal_Compress_Parallel-4     6893208               219.7 ns/op            72 B/op          1 allocs/op
