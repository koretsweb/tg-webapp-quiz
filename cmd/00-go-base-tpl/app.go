package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Setupper interface {
	Setup(ctx context.Context) error
}

type App struct {
	log *zap.Logger

	startupTimeout  time.Duration
	shutdownTimeout time.Duration

	setuppers []Setupper

	mongoClient *mongo.Client

	serverListener net.Listener
	server         *http.Server
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := a.boot(ctx); err != nil {
		return fmt.Errorf("boot: %w", err)
	}

	appErrs := make([]error, 0, 2)

	if err := a.run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			appErrs = append(appErrs, fmt.Errorf("run: %w", err))
		}
	}

	shutdownCtx, cancelShutdownTimeout := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancelShutdownTimeout()

	if err := a.shutdown(shutdownCtx); err != nil {
		appErrs = append(appErrs, fmt.Errorf("shutdown: %w", err))
	}

	return errors.Join(appErrs...)
}

func (a *App) boot(ctx context.Context) error {
	// todo
	//if err := a.mongoClient.Connect(ctx); err != nil {
	//	return fmt.Errorf("mongo client: connect: %w", err)
	//}
	//
	//if err := a.mongoClient.Ping(ctx, nil); err != nil {
	//	return fmt.Errorf("mongo client: ping: %w", err)
	//}

	//for i, setupper := range a.setuppers {
	//	if err := setupper.Setup(ctx); err != nil {
	//		return fmt.Errorf("setup[%d]: %w", i, err)
	//	}
	//}

	return nil
}

func (a *App) run(ctx context.Context) error {
	errCh := make(chan error)

	go func() {
		//a.log.Info("http server started", zap.String("addr", a.serverListener.Addr().String()))

		err := a.server.Serve(a.serverListener)
		if err == nil {
			return
		}

		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		select {
		case <-ctx.Done():
		case errCh <- fmt.Errorf("http server: %w", err):
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (a *App) shutdown(ctx context.Context) error {
	shutdownErrs := make([]error, 0, 2)

	if err := a.server.Shutdown(ctx); err != nil {
		shutdownErrs = append(shutdownErrs, fmt.Errorf("shutdown server: %w", err))
	}

	if err := a.mongoClient.Disconnect(ctx); err != nil {
		shutdownErrs = append(shutdownErrs, fmt.Errorf("mongo disconnect: %w", err))
	}

	return errors.Join(shutdownErrs...)
}
