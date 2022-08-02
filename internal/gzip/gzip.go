package gzip

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
)

var (
	writerPool = newGzipWriterPool()
	readerPool = sync.Pool{
		New: func() any {
			return new(gzip.Reader)
		},
	}
	bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(nil)
		},
	}
)

func Zip(data []byte) ([]byte, error) {
	return ZipLevel(data, gzip.BestSpeed)
}

func ZipLevel(data []byte, level int) (dst []byte, err error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	idx := getWriterPoolIndex(level)
	zw := writerPool[idx].Get().(*gzip.Writer)
	zw.Reset(buf)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
		writerPool[idx].Put(zw)
	}()

	_, err = zw.Write(data)
	if err != nil {
		return
	}
	err = zw.Flush()
	if err != nil {
		return
	}
	err = zw.Close()
	if err != nil {
		return
	}

	dst = buf.Bytes()
	return
}

func Unzip(data []byte) (src []byte, err error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	_, err = buf.Write(data)
	if err != nil {
		return
	}

	zr := readerPool.Get().(*gzip.Reader)
	defer func() {
		readerPool.Put(zr)
	}()

	err = zr.Reset(buf)
	if err != nil {
		return
	}
	defer func() {
		_ = zr.Close()
	}()

	src, err = io.ReadAll(zr)
	if err != nil {
		return
	}
	return
}

func newGzipWriterPool() (pools []*sync.Pool) {
	for i := 0; i < 12; i++ {
		level := i - 2
		pools = append(pools, &sync.Pool{
			New: func() any {
				zw, _ := gzip.NewWriterLevel(nil, level)
				return zw
			},
		})
	}
	return
}

func getWriterPoolIndex(level int) int {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	return level + 2
}
