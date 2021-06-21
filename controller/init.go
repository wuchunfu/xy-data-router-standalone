package controller

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
)

func InitWebServer() {
	app := fiber.New(fiber.Config{
		ServerHeader:          conf.WebAPPName,
		BodyLimit:             conf.Config.SYSConf.LimitBody,
		DisableStartupMessage: true,
		// Immutable:             true,
	})

	// 限流 (有一定的 CPU 占用)
	if conf.Config.SYSConf.LimitExpiration > 0 && conf.Config.SYSConf.LimitRequest > 0 {
		app.Use(middleware.IPLimiter())
	}

	app.Use(middleware.RecoverLogger(), middleware.HTTPCounter(), compress.New())
	setupRouter(app)

	common.Log.Info().Str("addr", conf.Config.SYSConf.WebServerAddr).Msg("Listening and serving HTTP")
	if err := app.Listen(conf.Config.SYSConf.WebServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}
