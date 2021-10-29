//go:build cgo
// +build cgo

package zstd

import (
	"bytes"
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
				if !bytes.Equal(src, bs) {
					b.Fatal(src, bs)
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
				if !bytes.Equal(src, bs) {
					b.Fatal(src, bs)
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
				if !bytes.Equal(src, bs) {
					b.Fatal(src, bs)
				}
				// _ = dec
			}
		})
	})
}

// go test -run=^$ -benchmem -benchtime=1s -count=2 -bench=.
// goos: linux
// goarch: amd64
// pkg: github.com/fufuok/internal/zstd
// cpu: Intel(R) Xeon(R) CPU E5-2667 v2 @ 3.30GHz
// BenchmarkName/gzip-4              187963              6692 ns/op            1748 B/op          2 allocs/op
// BenchmarkName/gzip-4              183778              6285 ns/op            1827 B/op          2 allocs/op
// BenchmarkName/pure-4              338545              3632 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/pure-4              342086              3606 ns/op            1024 B/op          2 allocs/op
// BenchmarkName/cgo-4               440370              2562 ns/op            1216 B/op          2 allocs/op
// BenchmarkName/cgo-4               487807              2587 ns/op            1216 B/op          2 allocs/op
