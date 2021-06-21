package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fufuok/xy-data-router/middleware"
)

func setupRouter(app *gin.Engine) *gin.Engine {
	// 动态接口
	v1 := app.Group("/v1/:apiname", middleware.WebAPILogger())
	{
		v1.POST("/bulk/gzip", V1APIHandler)
		v1.POST("/bulk", V1APIHandler)
		v1.POST("/gzip", V1APIHandler)
		v1.POST("", V1APIHandler)
		v1.GET("", V1APIHandler)
	}

	// 兼容旧 ES 上报接口
	oldAPI := []string{"/start/", "/stop/", "/tp2cn/", "/pubg_proxy/bulk/", "/tcp_proxy/bulk/"}
	for _, u := range oldAPI {
		app.POST(u, oldAPIHandler(nil))
	}

	// 测速数据上报 JSON 修正 (临时方案)
	app.POST("/speed_report/", oldAPIHandler([]string{"data.node_line_type"}))

	// ES 相关接口
	es := app.Group("/es", middleware.CheckESWhiteList(true))
	{
		// ES 通用查询
		es.POST("/search", ESSearchHandler)
		// ES Scroll
		es.POST("/scroll", ESScrollHandler)
	}

	// 健康检查
	app.GET("/heartbeat", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	app.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "PONG")
	})

	// 服务器状态
	app.GET("/sys/status", runningStatusHandler)
	app.GET("/sys/status/queue", runningQueueStatusHandler)
	app.GET("/sys/check", middleware.CheckESWhiteList(false), func(c *gin.Context) {
		remoteIP, _ := c.RemoteIP()
		c.String(http.StatusOK, c.ClientIP()+" - "+remoteIP.String())
	})

	// 异常请求
	app.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404")
	})

	app.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "405")
	})

	return app
}
