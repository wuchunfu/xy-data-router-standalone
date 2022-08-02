package json

import (
	"unsafe"
)

// MustJSONIndent 转 json 返回 []byte
func MustJSONIndent(v any) []byte {
	js, _ := MarshalIndent(v, "", "  ")
	return js
}

// MustJSON 转 json 返回 []byte
func MustJSON(v any) []byte {
	js, _ := Marshal(v)
	return js
}

// MustJSONString 转 json 返回 string
func MustJSONString(v any) string {
	js := MustJSON(v)
	return *(*string)(unsafe.Pointer(&js))
}
