package json

import (
	"unsafe"
)

// MustJSONIndent 转 json 返回 []byte
func MustJSONIndent(v interface{}) []byte {
	js, _ := MarshalIndent(v, "", "  ")
	return js
}

// MustJSON 转 json 返回 []byte
func MustJSON(v interface{}) []byte {
	js, _ := Marshal(v)
	return js
}

// MustJSONString 转 json 返回 string
func MustJSONString(v interface{}) string {
	js := MustJSON(v)
	return *(*string)(unsafe.Pointer(&js))
}
