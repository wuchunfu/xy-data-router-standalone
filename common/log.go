package common

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/imroc/req"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	Log        zerolog.Logger
	LogSampled zerolog.Logger
)

func initLogger() {
	if err := InitLogger(); err != nil {
		log.Fatalln("Failed to initialize logger:", err, "\nbye.")
	}

	// 路径脱敏, 日志格式规范, 避免与自定义字段名冲突: {"E":"is Err(error)","error":"is Str(error)"}
	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.ErrorFieldName = "E"
	zerolog.CallerFieldName = "F"
	zerolog.ErrorStackFieldName = "S"
	zerolog.DurationFieldInteger = true
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
}

func InitLogger() error {
	if err := LogConfig(); err != nil {
		return err
	}

	// 抽样的日志记录器
	sampler := &zerolog.BurstSampler{
		Burst:  conf.Config.SYSConf.Log.Burst,
		Period: conf.Config.SYSConf.Log.PeriodDur,
	}
	LogSampled = Log.Sample(&zerolog.LevelSampler{
		TraceSampler: sampler,
		DebugSampler: sampler,
		InfoSampler:  sampler,
		WarnSampler:  sampler,
		ErrorSampler: sampler,
	})

	req.Debug = conf.Config.SYSConf.Debug

	return nil
}

// LogConfig 日志配置
// 1. 开发环境时, 日志高亮输出到控制台
// 2. 生产环境时, 日志输出到文件(可选关闭高亮, 保存最近 10 个 30 天内的日志)
func LogConfig() error {
	basicLog := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "0102 15:04:05"}

	if !conf.Config.SYSConf.Debug {
		basicLog.NoColor = conf.Config.SYSConf.Log.NoColor
		basicLog.Out = &lumberjack.Logger{
			Filename:   conf.Config.SYSConf.Log.File,
			MaxSize:    conf.Config.SYSConf.Log.MaxSize,
			MaxAge:     conf.Config.SYSConf.Log.MaxAge,
			MaxBackups: conf.Config.SYSConf.Log.MaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
	}

	Log = zerolog.New(basicLog).With().Timestamp().Caller().Logger()
	Log = Log.Level(zerolog.Level(conf.Config.SYSConf.Log.Level))

	return nil
}
