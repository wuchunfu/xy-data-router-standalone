package controller

import (
	"log"
	"sync/atomic"

	"github.com/fufuok/utils/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

func InitWebServer() {
	app := fiber.New(fiber.Config{
		ServerHeader:          conf.WebAPPName,
		BodyLimit:             conf.Config.SYSConf.LimitBody,
		JSONEncoder:           json.Marshal,
		DisableStartupMessage: true,
		StrictRouting:         true,
		DisableKeepalive:      !conf.Config.SYSConf.EnableKeepalive,
		ReduceMemoryUsage:     conf.Config.SYSConf.ReduceMemoryUsage,
		ErrorHandler:          errorHandler,
		// Immutable:             true,
	})

	// 限流 (有一定的 CPU 占用)
	if conf.Config.SYSConf.LimitExpiration > 0 && conf.Config.SYSConf.LimitRequest > 0 {
		app.Use(middleware.IPLimiter())
	}

	app.Use(middleware.CheckESBlackList(true), middleware.RecoverLogger(), middleware.HTTPCounter(), compress.New())
	setupRouter(app)

	common.Log.Info().Str("addr", conf.Config.SYSConf.WebServerAddr).Msg("Listening and serving HTTP")
	if err := app.Listen(conf.Config.SYSConf.WebServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}

// 请求错误处理
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	atomic.AddUint64(&service.HTTPBadRequestCounters, 1)
	if conf.Config.SYSConf.Debug {
		common.LogSampled.Error().Err(err).
			Str("client_ip", c.IP()).Str("uri", c.OriginalURL()).
			Msg(c.Method())
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return c.Status(code).SendString(err.Error())
}
