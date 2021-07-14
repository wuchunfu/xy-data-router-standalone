package tunnel

import (
	"github.com/rs/zerolog"

	"github.com/fufuok/xy-data-router/common"
)

// zerolog of arpc
type logger struct{}

func (l *logger) SetLevel(lvl int) {}

func (l *logger) SetLogger(logger zerolog.Logger) {
	common.LogSampled = logger
}

func (l *logger) Debug(format string, v ...interface{}) {
	common.LogSampled.Debug().Msgf(format, v...)
}

func (l *logger) Info(format string, v ...interface{}) {
	common.LogSampled.Info().Msgf(format, v...)
}

func (l *logger) Warn(format string, v ...interface{}) {
	common.LogSampled.Warn().Msgf(format, v...)
}

func (l *logger) Error(format string, v ...interface{}) {
	common.LogSampled.Error().Msgf(format, v...)
}
