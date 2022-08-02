package tunnel

import (
	"errors"

	"github.com/fufuok/bytespool"

	"github.com/fufuok/xy-data-router/internal/zstd"
	"github.com/fufuok/xy-data-router/service/schema"
)

const (
	// 普通数据和压缩数据标识字符
	flagData byte = 'd'
	flagZstd byte = 'z'
)

var codecError = errors.New("invalid schema.DataItem")

// genCodec wraps schema.DataItem
type genCodec struct{}

// Marshal wraps schema.DataItem.Marshal
func (j *genCodec) Marshal(v any) ([]byte, error) {
	d, ok := v.(*schema.DataItem)
	if !ok {
		return nil, codecError
	}

	n := d.Size() + 1
	bs := bytespool.New64(n)
	buf := bs[1:]

	// 编码数据
	buf, _ = d.Marshal(buf)

	if schema.FlagType(d.Flag) != schema.FlagZstd {
		bs[0] = flagData
		return bs, nil
	}

	// 压缩数据
	dec := bytespool.Make64(n)
	dec = append(dec, flagZstd)
	dec = zstd.Compress(dec, buf)
	bytespool.Release(bs)

	return dec, nil
}

// Unmarshal wraps schema.DataItem.Unmarshal
func (j *genCodec) Unmarshal(data []byte, v any) (err error) {
	d, ok := v.(*schema.DataItem)
	if !ok {
		return codecError
	}

	if data[0] != flagZstd {
		_, err = d.Unmarshal(data[1:])
		return
	}

	// 先解压
	enc := bytespool.Make(len(data) * 2)
	enc, err = zstd.Decompress(enc, data[1:])

	// 再解码
	_, err = d.Unmarshal(enc)
	bytespool.Release(enc)

	return
}
