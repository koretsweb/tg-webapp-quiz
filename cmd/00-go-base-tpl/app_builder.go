package main

import (
	"00-go-base-tpl-sv/cmd/00-go-base-tpl/handler"
	"00-go-base-tpl-sv/internal/player"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AppBuilder struct {
	config         *Config
	log            *zap.Logger
	serverListener net.Listener
}

func NewAppBuilder() *AppBuilder {
	return &AppBuilder{}
}

func (a *AppBuilder) SetConfig(c *Config) *AppBuilder {
	a.config = c

	return a
}

func (a *AppBuilder) SetLog(log *zap.Logger) *AppBuilder {
	a.log = log

	return a
}

func (a *AppBuilder) SetServerListener(listener net.Listener) *AppBuilder {
	a.serverListener = listener

	return a
}

func (a *AppBuilder) Build() (*App, error) {
	if a.config == nil {
		return nil, errors.New("config must be defined")
	}

	if a.log == nil {
		a.log = zap.NewNop()
	}

	if a.serverListener == nil {
		listener, err := net.Listen("tcp", a.config.HTTP.Listen)
		if err != nil {
			return nil, fmt.Errorf("create server listener: %w", err)
		}

		a.serverListener = listener
	}

	return a.createApp(), nil
}

func (a *AppBuilder) createApp() *App {
	var (
		playerStorage = a.createPlayerStorage()
	)

	var (
		playerSv      = a.createPlayerService(playerStorage)
		playerHandler = handler.NewPlayers(playerSv, a.log)
	)

	var (
		router = mux.NewRouter()
		server = a.createHTTPServer(router)
	)

	a.registerHTTPHandlers(router, playerHandler)

	return &App{
		//log: a.log.Named("app"),
		//
		startupTimeout:  a.config.App.StartupTimeout,
		shutdownTimeout: a.config.App.ShutdownTimeout,
		//
		setuppers: []Setupper{playerStorage},
		//
		server:         server,
		serverListener: a.serverListener,
	}
}

func (a *AppBuilder) createHTTPServer(h http.Handler) *http.Server {
	return &http.Server{
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

func (a *AppBuilder) createPlayerStorage() *player.StorageMongo {
	return player.NewStorageMongo()
}

func (a *AppBuilder) createPlayerService(storage player.Storage) player.Service {
	return player.NewService(
		a.config.App.ServiceName,
		storage,
	)
}

func (a *AppBuilder) registerHTTPHandlers(
	router *mux.Router,
	playerHandler *handler.Players,
) {
	playerHandler.Register(router)
}
