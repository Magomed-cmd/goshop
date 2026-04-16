package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level       string
	Encoding    string
	Development bool
}

func NewFromGinMode(ginMode string) *zap.Logger {
	if ginMode == "release" {
		return NewProduction()
	}
	return NewDevelopment()
}

func NewDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, _ := config.Build()
	return logger
}

func NewProduction() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}
