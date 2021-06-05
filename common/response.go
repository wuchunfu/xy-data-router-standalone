package common

// API 标准返回, 内部规范
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

// API 请求失败返回值
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

// API 请求成功返回值
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

// API 请求成功返回, 无数据
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
