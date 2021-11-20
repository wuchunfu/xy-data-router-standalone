package tunnel

import (
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/conf"
)

var logType = " LogSampled"

// zerolog of arpc
// 注意: 受抽样日志影响, 日志可能不会被全部输出
type logger struct {
	log zerolog.Logger
}

func newLogger() *logger {
	l := &logger{
		log: common.LogSampled,
	}
	if conf.Debug {
		l.SetLogger(common.Log)
		logType = ""
	}
	return l
}

func (l *logger) SetLevel(lvl int) {}

func (l *logger) SetLogger(logger zerolog.Logger) {
	l.log = logger
}

func (l *logger) Debug(format string, v ...interface{}) {
	l.log.Debug().Msgf(format, v...)
}

func (l *logger) Info(format string, v ...interface{}) {
	l.log.Info().Msgf(format, v...)
}

func (l *logger) Warn(format string, v ...interface{}) {
	l.log.Warn().Msgf(format, v...)
}

func (l *logger) Error(format string, v ...interface{}) {
	l.log.Error().Msgf(format, v...)
}
