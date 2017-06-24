package client

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a logger for this package
var Logger *CompositeLogger

func init() {
	fileConfig := zap.NewDevelopmentConfig()
	fileConfig.OutputPaths = []string{"./pentahotools.log"}
	fileConfig.Level.SetLevel(zap.DebugLevel)
	fileLogger, _ := fileConfig.Build()

	consoleConfig := zap.NewDevelopmentConfig()
	consoleConfig.Encoding = "console"
	consoleConfig.EncoderConfig.TimeKey = ""
	consoleConfig.EncoderConfig.StacktraceKey = ""
	consoleConfig.OutputPaths = []string{"stdout"}
	consoleConfig.Level.SetLevel(zap.WarnLevel)
	consoleLogger, _ := consoleConfig.Build()

	Logger = newCompositeLogger(*consoleLogger, *fileLogger)
}

// CompositeLogger merges multiple loggers.
type CompositeLogger struct {
	Logger []zap.Logger
}

func newCompositeLogger(logger ...zap.Logger) *CompositeLogger {
	return &CompositeLogger{logger}
}

// Debug records the debug logs
func (l *CompositeLogger) Debug(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Debug(msg, fields...)
	}
}

// Warn records the warning logs
func (l *CompositeLogger) Warn(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Warn(msg, fields...)
	}
}

// Error records the error logs
func (l *CompositeLogger) Error(msg string, fields ...zapcore.Field) error {
	for _, logger := range l.Logger {
		logger.Error(msg, fields...)
	}
	return errors.New(msg)
}
