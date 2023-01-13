package tunnel

import (
	"github.com/lesismal/arpc/log"
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var (
	logType   = " LogSampled"
	arpcDebug bool
)

// zerolog of arpc
// 注意: 受抽样日志影响, 日志可能不会被全部输出
type arpcLogger struct {
	log zerolog.Logger
}

func initLogger() {
	log.SetLogger(newLogger())
}

func loadLogger() {
	if arpcDebug == conf.Debug {
		return
	}
	arpcDebug = conf.Debug
	log.SetLogger(newLogger())
}

func newLogger() *arpcLogger {
	l := &arpcLogger{
		log: common.LogSampled,
	}
	if conf.Debug {
		l.SetLogger(common.Log)
		logType = ""
	}
	return l
}

func (l *arpcLogger) SetLevel(lvl int) {}

func (l *arpcLogger) SetLogger(appLogger zerolog.Logger) {
	l.log = appLogger
}

func (l *arpcLogger) Debug(format string, v ...any) {
	l.log.Debug().Msgf(format, v...)
}

func (l *arpcLogger) Info(format string, v ...any) {
	l.log.Info().Msgf(format, v...)
}

func (l *arpcLogger) Warn(format string, v ...any) {
	l.log.Warn().Msgf(format, v...)
}

func (l *arpcLogger) Error(format string, v ...any) {
	l.log.Error().Msgf(format, v...)
}
