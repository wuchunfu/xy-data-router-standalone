package tunnel

import (
	"testing"

	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/schema"
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
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coder.Marshal(data)
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
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coder.Marshal(data)
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
		for pb.Next() {
			_, _ = coder.Marshal(data)
		}
	})
}

func BenchmarkCodec_Unmarshal_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	coder := new(genCodec)
	dec, _ := coder.Marshal(data)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
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
		for pb.Next() {
			_, _ = coder.Marshal(data)
		}
	})
}

func BenchmarkCodec_Unmarshal_Compress_Parallel(b *testing.B) {
	data := schema.New("test", "7.7.7.7", utils.FastRandBytes(512))
	data.Flag = 1
	coder := new(genCodec)
	dec, _ := coder.Marshal(data)
	item := schema.Make()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
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
// BenchmarkCodec_Marshal-4                         5919986               204.3 ns/op           576 B/op          1 allocs/op
// BenchmarkCodec_Marshal-4                         5932430               201.0 ns/op           576 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal-4                      16941763               70.85 ns/op            16 B/op          2 allocs/op
// BenchmarkCodec_Unmarshal-4                      17128773               70.66 ns/op            16 B/op          2 allocs/op
// BenchmarkCodec_Marshal_Compress-4                 170994                6624 ns/op           480 B/op          1 allocs/op
// BenchmarkCodec_Marshal_Compress-4                 179944                6470 ns/op           480 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal_Compress-4               241256                4697 ns/op         16425 B/op          5 allocs/op
// BenchmarkCodec_Unmarshal_Compress-4               275373                4639 ns/op         16425 B/op          5 allocs/op
// BenchmarkCodec_Marshal_Parallel-4               10878728               118.0 ns/op           576 B/op          1 allocs/op
// BenchmarkCodec_Marshal_Parallel-4               10519323               109.9 ns/op           576 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal_Parallel-4              6563950               174.4 ns/op            16 B/op          2 allocs/op
// BenchmarkCodec_Unmarshal_Parallel-4              7055163               182.6 ns/op            16 B/op          2 allocs/op
// BenchmarkCodec_Marshal_Compress_Parallel-4        708220                1993 ns/op           480 B/op          1 allocs/op
// BenchmarkCodec_Marshal_Compress_Parallel-4        710470                2008 ns/op           480 B/op          1 allocs/op
// BenchmarkCodec_Unmarshal_Compress_Parallel-4      344793                3157 ns/op         16425 B/op          5 allocs/op
// BenchmarkCodec_Unmarshal_Compress_Parallel-4      422097                3117 ns/op         16425 B/op          5 allocs/op
