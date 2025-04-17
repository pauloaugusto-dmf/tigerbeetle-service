package logger

import (
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func Init(isProduction bool) {
	var logger *zap.Logger
	var err error

	if isProduction {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(err)
	}

	log = logger.Sugar()
}

func Info(msg string, fields ...interface{}) {
	log.Infow(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	log.Errorw(msg, fields...)
}

func Debug(msg string, fields ...interface{}) {
	log.Debugw(msg, fields...)
}
