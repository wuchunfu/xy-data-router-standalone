package common

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	Log        zerolog.Logger
	LogSampled zerolog.Logger
)

// Logger 注意: 受抽样日志影响, 日志可能不会被全部输出
type Logger struct {
	log zerolog.Logger
}

func initLogger() {
	loadLogger()
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

func loadLogger() {
	newLogger()
	// 抽样的日志记录器
	sampler := &zerolog.BurstSampler{
		Burst:  conf.Config.LogConf.Burst,
		Period: conf.Config.LogConf.PeriodDur,
	}
	LogSampled = Log.Sample(&zerolog.LevelSampler{
		TraceSampler: sampler,
		DebugSampler: sampler,
		InfoSampler:  sampler,
		WarnSampler:  sampler,
		ErrorSampler: sampler,
	})
}

// 日志配置
// 1. 开发环境时, 日志高亮输出到控制台
// 2. 生产环境时, 日志输出到文件(可选关闭高亮, 保存最近 10 个 30 天内的日志)
func newLogger() {
	basicLog := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "0102 15:04:05"}

	if !conf.Debug {
		basicLog.NoColor = conf.Config.LogConf.NoColor
		basicLog.Out = &lumberjack.Logger{
			Filename:   conf.Config.LogConf.File,
			MaxSize:    conf.Config.LogConf.MaxSize,
			MaxAge:     conf.Config.LogConf.MaxAge,
			MaxBackups: conf.Config.LogConf.MaxBackups,
			LocalTime:  true,
			Compress:   true,
		}
	}
	Log = zerolog.New(basicLog).With().Timestamp().Caller().Logger()
	Log = Log.Level(zerolog.Level(conf.Config.LogConf.Level))
}

// NewAppLogger 类库日志实现: Req / Ants
func NewAppLogger() *Logger {
	if conf.Debug {
		return &Logger{
			log: Log,
		}
	}
	return &Logger{
		log: LogSampled,
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.log.Debug().Msgf(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.log.Warn().Msgf(format, v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.log.Warn().Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.log.Error().Msgf(format, v...)
}
