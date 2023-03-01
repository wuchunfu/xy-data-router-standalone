package web

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
	"github.com/fufuok/xy-data-router/internal/logger"
	"github.com/fufuok/xy-data-router/internal/logger/sampler"
	"github.com/fufuok/xy-data-router/web/middleware"
	"github.com/fufuok/xy-data-router/web/response"
	"github.com/fufuok/xy-data-router/web/router/es"
)

//go:embed assets/favicon.ico
var fav embed.FS

var app *fiber.App

// InitMain 程序启动时初始化
func InitMain() {
	go initWeb()
	es.InitMain()
}

// InitRuntime 重新加载或初始化运行时配置
func InitRuntime() {
	es.InitRuntime()
}

func Stop() {
	es.Stop()
}

// 初始化 Web 服务
func initWeb() {
	app = fiber.New(fiber.Config{
		ServerHeader:            conf.APPName,
		BodyLimit:               conf.Config.WebConf.LimitBody,
		ReduceMemoryUsage:       conf.Config.WebConf.ReduceMemoryUsage,
		ProxyHeader:             conf.Config.WebConf.ProxyHeader,
		EnableTrustedProxyCheck: conf.Config.WebConf.EnableTrustedProxyCheck,
		TrustedProxies:          conf.Config.WebConf.TrustedProxies,
		DisableKeepalive:        conf.Config.WebConf.DisableKeepalive,
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		DisableStartupMessage:   true,
		StrictRouting:           true,
		ErrorHandler:            errorHandler,
	})

	// 限流 (有一定的 CPU 占用)
	if conf.Config.WebConf.LimitExpiration > 0 && conf.Config.WebConf.LimitRequest > 0 {
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

	if conf.Config.WebConf.ServerHttpsAddr != "" {
		go func() {
			logger.Info().Str("addr", conf.Config.WebConf.ServerHttpsAddr).Msg("Listening and serving HTTPS")
			if err := app.ListenTLS(conf.Config.WebConf.ServerHttpsAddr,
				conf.Config.WebConf.CertFile, conf.Config.WebConf.KeyFile); err != nil {
				log.Fatalln("Failed to start HTTPS Server:", err, "\nbye.")
			}
		}()
	}

	logger.Info().Str("addr", conf.Config.WebConf.ServerAddr).Msg("Listening and serving HTTP")
	if err := app.Listen(conf.Config.WebConf.ServerAddr); err != nil {
		log.Fatalln("Failed to start HTTP Server:", err, "\nbye.")
	}
}

// 请求错误处理
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	common.HTTPBadRequestCount.Inc()
	if conf.Debug {
		sampler.Error().Err(err).
			Str("client_ip", common.GetClientIP(c)).Str("method", c.Method()).Int("status_code", code).
			Msg(c.OriginalURL())
	}
	return response.APIException(c, code, err.Error(), nil)
}
