package bootstrap

import (
	"judge/jConfig"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func bootstrapLogger(config *jConfig.LoggerConfig) *zap.Logger {
	level := zap.ErrorLevel
	levelValid := true

	switch config.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		levelValid = false
	}

	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSizeInMegabytes,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAgeInDays,
			Compress:   config.Compress,
		}),
		level,
	))

	if !levelValid {
		logger.Fatal("Invalid log level. Choose one from debug, info, warn, error, dpanic, panic, fatal in logger.level.")
	}

	return logger
}
