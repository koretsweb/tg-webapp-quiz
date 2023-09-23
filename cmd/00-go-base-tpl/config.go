package main

import (
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type appConfig struct {
	Debug        bool   `mapstructure:"debug"`
	ServiceName  string `mapstructure:"service-name"`
	PProf        bool   `mapstructure:"pprof"`
	PyroscopeDSN string `mapstructure:"pyroscope-dsn"`

	StartupTimeout  time.Duration `mapstructure:"startup-timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout"`
}

type mongoConfig struct {
	DSN              string `mapstructure:"mongo-dsn"`
	Database         string `mapstructure:"mongo-db"`
	PlayerCollection string `mapstructure:"mongo-player-collection"`
}

type rmqConfig struct {
	DSN               string        `mapstructure:"rabbitmq-dsn"`
	FallbackDelay     time.Duration `mapstructure:"rabbitmq-fallback-delay"`
	MaxFailedAttempts int           `mapstructure:"rabbitmq-max-failed-attempt"`
	Heartbeat         time.Duration `mapstructure:"rabbitmq-heartbeat"`
}

type httpConfig struct {
	Listen string `mapstructure:"listen"`
}

type Config struct {
	App   appConfig   `mapstructure:",squash"`
	Mongo mongoConfig `mapstructure:",squash"`
	RMQ   rmqConfig   `mapstructure:",squash"`
	HTTP  httpConfig  `mapstructure:",squash"`
}

func ReadConfig() (*Config, error) {
	_ = godotenv.Load() // nolint

	pflag.Bool("debug", false, "Enable debug")
	pflag.String("service-name", "00-go-base-tpl", "Service name")
	pflag.String("telegram-bot-token", "", "")
	pflag.Duration("startup-timeout", 10*time.Second, "Timeout until application should be started")
	pflag.Duration("shutdown-timeout", 15*time.Second, "Timeout until application should be stopped")

	pflag.String("mongo-dsn", "mongodb://127.0.0.1:27017", "Mongo DSN")
	pflag.String("mongo-db", "00_go_base_tpl", "Mongo database for player") // TODO rename it
	pflag.String("mongo-player-collection", "player", "Mongo collection name for players")

	pflag.String("rabbitmq-dsn", "amqp://127.0.0.1:5672//", "RabbitMQ connection DSN")

	pflag.Duration("rabbitmq-fallback-delay", time.Second, "RabbitMQ delay before reconnection retry")
	pflag.Int("rabbitmq-max-failed-attempt", 5, "RabbitMQ max serial connection attempts before fail")
	pflag.Duration("rabbitmq-heartbeat", 5*time.Second, "RabbitMQ heartbeat duration")

	pflag.StringP("listen", "l", ":80", "HTTP binding address")

	pflag.Bool("pprof", false, "Enable pprof profiling")
	pflag.String("pyroscope-dsn", "http://localhost:4040", "Pyroscope DSN")

	pflag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}

	var config Config

	err = viper.Unmarshal(&config)

	return &config, err
}
