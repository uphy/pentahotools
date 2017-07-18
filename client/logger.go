package client

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the logger of pentahotools.
type Logger interface {
	// Debug records the debug logs
	Debug(msg string, fields ...zapcore.Field)
	// Warn records the warning logs
	Warn(msg string, fields ...zapcore.Field)
	// Error records the error logs
	Error(msg string, fields ...zapcore.Field) error
}

// CompositeLogger merges multiple loggers.
type compositeLogger struct {
	Logger []zap.Logger
}

func NewCompositeLogger() Logger {
	fileConfig := zap.NewDevelopmentConfig()
	fileConfig.OutputPaths = []string{"./pentahotools.log"}
	fileConfig.Level.SetLevel(zap.DebugLevel)
	fileLogger, _ := fileConfig.Build()

	consoleConfig := zap.NewDevelopmentConfig()
	consoleConfig.Encoding = "console"
	consoleConfig.EncoderConfig.TimeKey = ""
	consoleConfig.EncoderConfig.StacktraceKey = ""
	consoleConfig.EncoderConfig.CallerKey = ""
	consoleConfig.OutputPaths = []string{"stdout"}
	consoleConfig.Level.SetLevel(zap.WarnLevel)
	consoleLogger, _ := consoleConfig.Build()

	return newCompositeLogger(*consoleLogger, *fileLogger)
}

func NewConsoleLogger() Logger {
	consoleConfig := zap.NewDevelopmentConfig()
	consoleConfig.Encoding = "console"
	consoleConfig.EncoderConfig.TimeKey = ""
	consoleConfig.EncoderConfig.StacktraceKey = ""
	consoleConfig.OutputPaths = []string{"stdout"}
	consoleConfig.Level.SetLevel(zap.WarnLevel)
	consoleLogger, _ := consoleConfig.Build()
	return newCompositeLogger(*consoleLogger)
}

func newCompositeLogger(logger ...zap.Logger) *compositeLogger {
	return &compositeLogger{logger}
}

// Debug records the debug logs
func (l *compositeLogger) Debug(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Debug(msg, fields...)
	}
}

// Warn records the warning logs
func (l *compositeLogger) Warn(msg string, fields ...zapcore.Field) {
	for _, logger := range l.Logger {
		logger.Warn(msg, fields...)
	}
}

// Error records the error logs
func (l *compositeLogger) Error(msg string, fields ...zapcore.Field) error {
	for _, logger := range l.Logger {
		logger.Error(msg, fields...)
	}
	return errors.New(msg)
}
