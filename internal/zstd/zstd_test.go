//go:build cgo
// +build cgo

package zstd

import (
	"math/rand"
	"testing"

	"github.com/fufuok/utils"
	pure "github.com/klauspost/compress/zstd"
	cgo "github.com/valyala/gozstd"
)

func TestCompressDecompress(t *testing.T) {
	testCrossCompressDecompress(t, []byte("a"))
	testCrossCompressDecompress(t, []byte("foobarbaz"))

	var b []byte
	for i := 0; i < 64*1024; i++ {
		b = append(b, byte(rand.Int31n(256)))
	}
	testCrossCompressDecompress(t, b)
}

func testCrossCompressDecompress(t *testing.T, b []byte) {
	testCompressDecompress(t, pureCompress, pureDecompress, b)
	testCompressDecompress(t, cgoCompress, cgoDecompress, b)
	testCompressDecompress(t, pureCompress, cgoDecompress, b)
	testCompressDecompress(t, cgoCompress, pureDecompress, b)
}

func testCompressDecompress(t *testing.T, compress compressFn, decompress decompressFn, b []byte) {
	bc, err := compress(nil, b, 5)
	if err != nil {
		t.Fatalf("unexpected error when compressing b=%x: %s", b, err)
	}
	bNew, err := decompress(nil, bc)
	if err != nil {
		t.Fatalf("unexpected error when decompressing b=%x from bc=%x: %s", b, bc, err)
	}
	if string(bNew) != string(b) {
		t.Fatalf("invalid bNew; got\n%x; expecting\n%x", bNew, b)
	}

	prefix := []byte{1, 2, 33}
	bcNew, err := compress(prefix, b, 5)
	if err != nil {
		t.Fatalf("unexpected error when compressing b=%x: %s", bcNew, err)
	}
	if string(bcNew[:len(prefix)]) != string(prefix) {
		t.Fatalf("invalid prefix for b=%x; got\n%x; expecting\n%x", b, bcNew[:len(prefix)], prefix)
	}
	if string(bcNew[len(prefix):]) != string(bc) {
		t.Fatalf("invalid prefixed bcNew for b=%x; got\n%x; expecting\n%x", b, bcNew[len(prefix):], bc)
	}

	bNew, err = decompress(prefix, bc)
	if err != nil {
		t.Fatalf("unexpected error when decompressing b=%x from bc=%x with prefix: %s", b, bc, err)
	}
	if string(bNew[:len(prefix)]) != string(prefix) {
		t.Fatalf("invalid bNew prefix when decompressing bc=%x; got\n%x; expecting\n%x", bc, bNew[:len(prefix)], prefix)
	}
	if string(bNew[len(prefix):]) != string(b) {
		t.Fatalf("invalid prefixed bNew; got\n%x; expecting\n%x", bNew[len(prefix):], b)
	}
}

var pureEncoder, _ = pure.NewWriter(nil,
	pure.WithEncoderCRC(false), // Disable CRC for performance reasons.
	pure.WithEncoderLevel(pure.SpeedDefault),
)

type compressFn func(dst, src []byte, compressionLevel int) ([]byte, error)

func pureCompress(dst, src []byte, _ int) ([]byte, error) {
	return pureEncoder.EncodeAll(src, dst), nil
}

func cgoCompress(dst, src []byte, compressionLevel int) ([]byte, error) {
	return cgo.CompressLevel(dst, src, compressionLevel), nil
}

var pureDecoder, _ = pure.NewReader(nil)

type decompressFn func(dst, src []byte) ([]byte, error)

func pureDecompress(dst, src []byte) ([]byte, error) {
	return pureDecoder.DecodeAll(src, dst)
}

func cgoDecompress(dst, src []byte) ([]byte, error) {
	return cgo.Decompress(dst, src)
}

func BenchmarkName(b *testing.B) {
	bs := utils.FastRandBytes(512)
	b.ResetTimer()
	b.Run("gzip", func(b *testing.B) {
		// dec, _ := utils.Zip(bs)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// bs := utils.FastRandBytes(512)
				dec, _ := utils.Zip(bs)
				src, err := utils.Unzip(dec)
				if err != nil {
					b.Fatal(err)
				}
				if !utils.EqualFoldBytes(src, bs) {
					b.Fatal("src != bs")
				}
				// _ = dec
			}
		})
	})
	b.Run("pure", func(b *testing.B) {
		// dec, _ := pureCompress(nil, bs, 3)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// bs := utils.FastRandBytes(512)
				dec, _ := pureCompress(nil, bs, 3)
				src, err := pureDecompress(nil, dec)
				if err != nil {
					b.Fatal(err)
				}
				if !utils.EqualFoldBytes(src, bs) {
					b.Fatal("src != bs")
				}
				// _ = dec
			}
		})
	})
	b.Run("cgo", func(b *testing.B) {
		// dec, _ := cgoCompress(nil, bs, 3)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// bs := utils.FastRandBytes(512)
				dec, _ := cgoCompress(nil, bs, 3)
				src, err := cgoDecompress(nil, dec)
				if err != nil {
					b.Fatal(err)
				}
				if !utils.EqualFoldBytes(src, bs) {
					b.Fatal("src != bs")
				}
				// _ = dec
			}
		})
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=.
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/xy-data-router/internal/zstd
// cpu: Intel(R) Xeon(R) Gold 6151 CPU @ 3.00GHz
// BenchmarkName/gzip-4              201576              5922 ns/op            1783 B/op          2 allocs/op
// BenchmarkName/gzip-4              202632              6204 ns/op            1759 B/op          2 allocs/op
// BenchmarkName/pure-4              304629              3899 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/pure-4              316008              3897 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/cgo-4               577648              2225 ns/op            1216 B/op          2 allocs/op
// BenchmarkName/cgo-4               565663              2231 ns/op            1216 B/op          2 allocs/op

// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/internal/zstd
// cpu: Intel(R) Xeon(R) Gold 6161 CPU @ 2.20GHz
// BenchmarkName/gzip-8              273086              4276 ns/op            1875 B/op          2 allocs/op
// BenchmarkName/gzip-8              282141              4446 ns/op            1836 B/op          2 allocs/op
// BenchmarkName/pure-8              403152              2943 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/pure-8              380551              2942 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/cgo-8               906164              1521 ns/op            1216 B/op          2 allocs/op
// BenchmarkName/cgo-8               891444              1462 ns/op            1216 B/op          2 allocs/op

// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/internal/zstd
// cpu: Intel(R) Xeon(R) Silver 4114 CPU @ 2.20GHz
// BenchmarkName/gzip-20             447502              2652 ns/op            1898 B/op          2 allocs/op
// BenchmarkName/gzip-20             446084              2660 ns/op            1964 B/op          2 allocs/op
// BenchmarkName/pure-20             662720              1814 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/pure-20             652315              1836 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/cgo-20             1317346              907.7 ns/op           1216 B/op          2 allocs/op
// BenchmarkName/cgo-20             1322730              908.1 ns/op           1216 B/op          2 allocs/op
