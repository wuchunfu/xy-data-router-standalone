package middleware

import (
	"net/http"

	"github.com/fufuok/utils"
	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/common"
)

var apiSuccessNil = utils.MustJSON(common.APISuccessNil())

// 通用异常处理
func APIException(c *gin.Context, code int, msg string) {
	if msg == "" {
		msg = "错误的请求"
	}
	c.JSON(code, common.APIFailureData(msg))
	c.Abort()
}

// 返回失败, 状态码: 200
func APIFailure(c *gin.Context, msg string) {
	APIException(c, http.StatusOK, msg)
}

// 返回成功, 状态码: 200
func APISuccess(c *gin.Context, data interface{}, count int) {
	c.JSON(http.StatusOK, common.APISuccessData(data, count))
	c.Abort()
}

// 返回成功, 无数据, 状态码: 200
func APISuccessNil(c *gin.Context) {
	c.Data(http.StatusOK, "application/json; charset=utf-8", apiSuccessNil)
	c.Abort()
}

// 返回文本消息
func TxtMsg(c *gin.Context, msg string) {
	c.String(http.StatusOK, msg)
	c.Abort()
}
