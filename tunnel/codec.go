package tunnel

import (
	"errors"

	"github.com/fufuok/xy-data-router/common"
)

var (
	codecError = errors.New("invalid GenDataItem")
)

// genCodec wraps GenDataItem
type genCodec struct{}

// Marshal wraps GenDataItem.Marshal
func (j *genCodec) Marshal(v interface{}) ([]byte, error) {
	if d, ok := v.(*common.GenDataItem); ok {
		return d.Marshal(nil)
	}

	return nil, codecError
}

// Unmarshal wraps GenDataItem.Unmarshal
func (j *genCodec) Unmarshal(data []byte, v interface{}) error {
	if d, ok := v.(*common.GenDataItem); ok {
		_, err := d.Unmarshal(data)
		return err
	}

	return codecError
}
