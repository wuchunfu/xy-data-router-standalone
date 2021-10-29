//go:build !cgo
// +build !cgo

// Package zstd Ref: VictoriaMetrics/lib/encoding/zstd/
package zstd

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/klauspost/compress/zstd"
)

var (
	encoder *zstd.Encoder
	decoder *zstd.Decoder

	mu sync.Mutex
	av atomic.Value
)

type registry map[int]*zstd.Encoder

func init() {
	r := make(registry)
	av.Store(r)

	var err error
	decoder, err = zstd.NewReader(nil)
	if err != nil {
		log.Panicf("BUG: failed to create ZSTD reader: %s", err)
	}
	encoder = newEncoder(3)
}

// Decompress appends decompressed src to dst and returns the result.
func Decompress(dst, src []byte) ([]byte, error) {
	return decoder.DecodeAll(src, dst)
}

// Compress appends compressed src to dst and returns the result.
//
// The given compressionLevel is used for the compression.
func Compress(dst, src []byte) []byte {
	return encoder.EncodeAll(src, dst)
}

// CompressLevel appends compressed src to dst and returns the result.
//
// The given compressionLevel is used for the compression.
func CompressLevel(dst, src []byte, compressionLevel int) []byte {
	e := getEncoder(compressionLevel)
	return e.EncodeAll(src, dst)
}

func getEncoder(compressionLevel int) *zstd.Encoder {
	r := av.Load().(registry)
	e := r[compressionLevel]
	if e != nil {
		return e
	}

	mu.Lock()
	// Create the encoder under lock in order to prevent from wasted work
	// when concurrent goroutines create encoder for the same compressionLevel.
	r1 := av.Load().(registry)
	if e = r1[compressionLevel]; e == nil {
		e = newEncoder(compressionLevel)
		r2 := make(registry)
		for k, v := range r1 {
			r2[k] = v
		}
		r2[compressionLevel] = e
		av.Store(r2)
	}
	mu.Unlock()

	return e
}

func newEncoder(compressionLevel int) *zstd.Encoder {
	level := zstd.EncoderLevelFromZstd(compressionLevel)
	e, err := zstd.NewWriter(nil,
		zstd.WithEncoderCRC(false), // Disable CRC for performance reasons.
		zstd.WithEncoderLevel(level))
	if err != nil {
		log.Panicf("BUG: failed to create ZSTD writer: %s", err)
	}
	return e
}
