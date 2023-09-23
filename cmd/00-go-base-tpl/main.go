package main

import (
	"context"
	"os/signal"
	"syscall"
)

func main() {
	var (
		ctx, stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

		config = MustReadConfig()

		//log = CreateLogger(config)
	)

	//zap.RedirectStdLog(log) // todo

	//defer log.Sync() // nolint skip sync errors as non-important
	defer stop()

	//mustSetMaxProcs(log)

	//app, err := NewAppBuilder().SetConfig(config).SetLog(log).Build()
	app, err := NewAppBuilder().SetConfig(config).Build()
	if err != nil {
		//log.Panic("app build failed", zap.Error(err))
	}

	if err := app.Run(ctx); err != nil {
		//log.Panic("application run error", zap.Error(err))
	}
}
