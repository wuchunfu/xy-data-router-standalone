package common

import (
	"strconv"

	"github.com/fufuok/utils"
)

var (
	// 减少 JSON 序列化, 用于拼接 JSON Bytes 数据
	okA = []byte(`{"id":1,"ok":1,"code":0,"msg":"","data":`)
	okB = []byte(`,"count":`)
	okC = []byte(`}`)
)

// TAPIData API 标准返回, 内部规范
// id: 1, ok: 1, code: 0 成功; id: 0, ok: 0, code: 1 失败
// 成功时 msg 必定为空
type TAPIData struct {
	ID    int         `json:"id"`
	OK    int         `json:"ok"`
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Count int         `json:"count"`
}

// APIFailureData API 请求失败返回值
func APIFailureData(msg string) *TAPIData {
	return &TAPIData{
		ID:    0,
		OK:    0,
		Code:  1,
		Msg:   msg,
		Data:  nil,
		Count: 0,
	}
}

// APISuccessData API 请求成功返回值
func APISuccessData(data interface{}, count int) *TAPIData {
	return &TAPIData{
		ID:    1,
		OK:    1,
		Code:  0,
		Msg:   "",
		Data:  data,
		Count: count,
	}
}

// APISuccessBytes API 请求成功返回值(JSON Bytes)
func APISuccessBytes(data []byte, count int) []byte {
	n := utils.S2B(strconv.Itoa(count))
	return utils.JoinBytes(okA, data, okB, n, okC)
}

// APISuccessNil API 请求成功返回, 无数据
func APISuccessNil() *TAPIData {
	return &TAPIData{
		ID:    1,
		OK:    1,
		Code:  0,
		Msg:   "",
		Data:  nil,
		Count: 0,
	}
}
