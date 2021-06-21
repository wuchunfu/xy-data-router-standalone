package controller

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
)

func InitWebServer() {
	app := gin.New()

	if conf.Config.SYSConf.Debug {
		gin.SetMode(gin.DebugMode)
		app.Use(gin.Logger())
	} else {
		// 生产环境不记录请求日志
		gin.SetMode(gin.ReleaseMode)
	}

	app.Use(
		gzip.DefaultHandler().Gin,
		middleware.RecoverLogger(true),
		middleware.HTTPCounter(),
		middleware.ResponseHeaders(),
	)

	setupRouter(app)

	common.Log.Info().Str("addr", conf.Config.SYSConf.WebServerAddr).Msg("Listening and serving HTTP")
	if err := app.Run(conf.Config.SYSConf.WebServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}
