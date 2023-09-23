package main

import (
	"fmt"
)

func MustReadConfig() *Config {
	config, err := ReadConfig()
	if err != nil {
		panic(fmt.Sprintf("read config: %s", err))
	}

	return config
}

//func CreateLogger(config *Config) *zap.Logger {
//	if config.App.Debug {
//		return logger.NewZapWithOptions(
//			logger.WithDebug(true),
//			logger.WithLogLevel(zapcore.DebugLevel),
//		)
//	}
//
//	return logger.NewZapWithOptions(
//		logger.WithLogLevel(zapcore.InfoLevel),
//	)
//} todo
