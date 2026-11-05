package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})
	Debug(msg string, fields ...map[string]interface{})
	With(fields map[string]interface{}) Logger
}

type ZapLogger struct {
	logger *zap.Logger
}

// APP_ENV=development | staging | production
// LOG_LEVEL=debug | info | warn | error
func NewZapLogger() *ZapLogger {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var zapConfig zap.Config
	if env == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.DisableStacktrace = true
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.DisableStacktrace = true
	}

	switch level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	l, _ := zapConfig.Build(zap.AddCaller())
	return &ZapLogger{logger: l}
}

func (z *ZapLogger) Info(msg string, fields ...map[string]interface{}) {
	logger := z.logger.WithOptions(zap.AddCallerSkip(1))
	if len(fields) > 0 {
		logger.Info(msg, zap.Any("fields", fields[0]))
	} else {
		logger.Info(msg)
	}
}

func (z *ZapLogger) Warn(msg string, fields ...map[string]interface{}) {
	logger := z.logger.WithOptions(zap.AddCallerSkip(1))
	if len(fields) > 0 {
		logger.Warn(msg, zap.Any("fields", fields[0]))
	} else {
		logger.Warn(msg)
	}
}

func (z *ZapLogger) Error(msg string, fields ...map[string]interface{}) {
	logger := z.logger.WithOptions(zap.AddCallerSkip(1))
	if len(fields) > 0 {
		logger.Error(msg, zap.Any("fields", fields[0]))
	} else {
		logger.Error(msg)
	}
}

func (z *ZapLogger) Debug(msg string, fields ...map[string]interface{}) {
	logger := z.logger.WithOptions(zap.AddCallerSkip(1))
	if len(fields) > 0 {
		logger.Debug(msg, zap.Any("fields", fields[0]))
	} else {
		logger.Debug(msg)
	}
}

func (z *ZapLogger) With(fields map[string]interface{}) Logger {
	f := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		f = append(f, zap.Any(k, v))
	}
	return &ZapLogger{logger: z.logger.With(f...)}
}
