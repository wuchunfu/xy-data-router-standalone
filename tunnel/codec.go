package tunnel

import (
	"errors"

	"github.com/fufuok/xy-data-router/schema"
)

var (
	codecError = errors.New("invalid GenDataItem")
)

// genCodec wraps GenDataItem
type genCodec struct{}

// Marshal wraps GenDataItem.Marshal
func (j *genCodec) Marshal(v interface{}) ([]byte, error) {
	if d, ok := v.(*schema.DataItem); ok {
		return d.Marshal(nil)
	}

	return nil, codecError
}

// Unmarshal wraps GenDataItem.Unmarshal
func (j *genCodec) Unmarshal(data []byte, v interface{}) error {
	if d, ok := v.(*schema.DataItem); ok {
		_, err := d.Unmarshal(data)
		return err
	}

	return codecError
}
