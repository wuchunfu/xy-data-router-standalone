package controller

import (
	"embed"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
	"github.com/fufuok/xy-data-router/middleware"
	"github.com/fufuok/xy-data-router/service"
)

//go:embed assets/favicon.ico
var fav embed.FS

// InitWebServer 接口服务
func InitWebServer() {
	app := fiber.New(fiber.Config{
		ServerHeader:          conf.APPName,
		BodyLimit:             conf.Config.SYSConf.LimitBody,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
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

	app.Use(
		middleware.RecoverLogger(),
		middleware.CheckESBlackList(true),
		favicon.New(favicon.Config{
			File:       "assets/favicon.ico",
			FileSystem: http.FS(fav),
		}),
		middleware.HTTPCounter(),
		compress.New(),
	)
	setupRouter(app)

	if conf.Config.SYSConf.WebServerHttpsAddr != "" {
		go func() {
			common.Log.Info().Str("addr", conf.Config.SYSConf.WebServerHttpsAddr).Msg("Listening and serving HTTPS")
			if err := app.ListenTLS(conf.Config.SYSConf.WebServerHttpsAddr,
				conf.Config.SYSConf.WebCertFile, conf.Config.SYSConf.WebKeyFile); err != nil {
				log.Fatalln("Failed to start HTTPS Server:", err, "\nbye.")
			}
		}()
	}

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

	service.HTTPBadRequestCount.Inc()
	if conf.Debug {
		common.LogSampled.Error().Err(err).
			Str("client_ip", c.IP()).Str("uri", c.OriginalURL()).
			Msg(c.Method())
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return c.Status(code).SendString(err.Error())
}
