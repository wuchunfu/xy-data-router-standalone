package tunnel

import (
	"errors"

	"github.com/fufuok/bytespool"
	"github.com/fufuok/utils"

	"github.com/fufuok/xy-data-router/internal/zstd"
	"github.com/fufuok/xy-data-router/schema"
)

const (
	// 普通数据和压缩数据标识字符
	flagData byte = 'd'
	flagZstd byte = 'z'
)

var (
	// 普通数据和压缩数据标识
	flagDatas = []byte{flagData}
	flagZstds = []byte{flagZstd}

	codecError = errors.New("invalid schema.DataItem")
)

// genCodec wraps schema.DataItem
type genCodec struct{}

// Marshal wraps schema.DataItem.Marshal
func (j *genCodec) Marshal(v interface{}) ([]byte, error) {
	d, ok := v.(*schema.DataItem)
	if !ok {
		return nil, codecError
	}

	// 编码数据
	bs := bytespool.Make()
	defer bytespool.Release(bs)
	bs, _ = d.Marshal(bs)

	if schema.FlagType(d.Flag) != schema.FlagZstd {
		return utils.JoinBytes(flagDatas, bs), nil
	}

	// 压缩数据
	dec := bytespool.Make()
	defer bytespool.Release(dec)
	dec = zstd.Compress(dec, bs)

	return utils.JoinBytes(flagZstds, dec), nil
}

// Unmarshal wraps schema.DataItem.Unmarshal
func (j *genCodec) Unmarshal(data []byte, v interface{}) (err error) {
	d, ok := v.(*schema.DataItem)
	if !ok {
		return codecError
	}

	if data[0] != flagZstd {
		_, err = d.Unmarshal(data[1:])
		return
	}

	// 先解压
	enc := bytespool.Make()
	enc, err = zstd.Decompress(enc, data[1:])

	// 再解码
	_, err = d.Unmarshal(enc)
	return
}
