package common

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fufuok/utils"
	"github.com/natefinch/lumberjack/v3"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/conf"
	"github.com/fufuok/xy-data-router/internal/json"
)

const (
	// 文件滚动单位
	megabyte = 1024 * 1024
	days     = 24 * time.Hour
)

var (
	// Log 通用日志, Debug 时输出到控制台, 否则写入日志文件
	Log zerolog.Logger

	// LogSampled 抽样日志
	LogSampled zerolog.Logger

	// LogAlarm 报警日志, 写入通用日志并发送报警
	LogAlarm zerolog.Logger

	logCurrentConf conf.LogConf
)

func initLogger() {
	if err := loadLogger(); err != nil {
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
	zerolog.InterfaceMarshalFunc = json.Marshal
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		i := strings.LastIndexByte(file, '/')
		if i == -1 {
			return file
		}
		i = strings.LastIndexByte(file[:i], '/')
		if i == -1 {
			return file
		}
		return file[i+1:] + ":" + strconv.Itoa(line)
	}
	Log.Info().Msg("Logger initialized successfully")
}

func loadLogger() error {
	if logCurrentConf == conf.Config.LogConf {
		return nil
	}
	logCurrentConf = conf.Config.LogConf

	if err := newLogger(); err != nil {
		return err
	}

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
	return nil
}

// 日志配置
// 1. 开发环境时, 日志高亮输出到控制台
// 2. 生产环境时, 日志输出到文件(可选关闭高亮, 保存最近 10 个 30 天内的日志)
func newLogger() error {
	basicLog := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    conf.Config.LogConf.NoColor,
		TimeFormat: "0102 15:04:05",
	}
	if !conf.Debug {
		fh, err := lumberjack.NewRoller(
			conf.Config.LogConf.File,
			// 以 MiB 为单位
			conf.Config.LogConf.MaxSize*megabyte,
			&lumberjack.Options{
				// 以 天 为单位
				MaxAge:     time.Duration(conf.Config.LogConf.MaxAge) * days,
				MaxBackups: conf.Config.LogConf.MaxBackups,
				LocalTime:  true,
				Compress:   true,
			})
		if err != nil {
			return err
		}
		basicLog.Out = fh
	}

	Log = zerolog.New(basicLog).With().Timestamp().Caller().Logger()
	Log = Log.Level(zerolog.Level(conf.Config.LogConf.Level))
	LogAlarm = zerolog.New(zerolog.MultiLevelWriter(basicLog, newAlarmWriter(zerolog.ErrorLevel))).
		With().Timestamp().Caller().Logger().
		Level(zerolog.Level(conf.Config.LogConf.Level))
	return nil
}

// 指定级别及以上日志发送到报警接口
type alarmWriter struct {
	lv zerolog.Level
}

func newAlarmWriter(lv zerolog.Level) *alarmWriter {
	return &alarmWriter{
		lv: lv,
	}
}

// Write 发送报警消息到接口
func (w *alarmWriter) Write(p []byte) (n int, err error) {
	if conf.Config.LogConf.PostAlarmAPI != "" {
		p := utils.CopyBytes(p)
		_ = GoPool.Submit(func() {
			sendAlarm(p)
		})
	}
	return len(p), nil
}

// WriteLevel 日志级别过滤
func (w *alarmWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= w.lv && l < zerolog.NoLevel {
		return w.Write(p)
	}
	return len(p), nil
}

type appLogger struct {
	log zerolog.Logger
}

// NewAppLogger 类库日志实现: Req / Ants
// 注意: 受抽样日志影响, 日志可能不会被全部输出
func NewAppLogger() *appLogger {
	if conf.Debug {
		return &appLogger{
			log: Log,
		}
	}
	return &appLogger{
		log: LogSampled,
	}
}

func (l *appLogger) Debugf(format string, v ...any) {
	l.log.Debug().Msgf(format, v...)
}

func (l *appLogger) Warnf(format string, v ...any) {
	l.log.Warn().Msgf(format, v...)
}

func (l *appLogger) Printf(format string, v ...any) {
	l.log.Warn().Msgf(format, v...)
}

func (l *appLogger) Errorf(format string, v ...any) {
	l.log.Error().Msgf(format, v...)
}
