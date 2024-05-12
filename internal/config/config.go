package config

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"time"
)

type Config struct {
	Server  ServerConfig  `env:", prefix=SERVER_"`
	Storage StorageConfig `env:", prefix=STORAGE_"`
	Auth    AuthConfig    `env:", prefix=AUTH_"`
	Log     LogConfig     `env:", prefix=LOG_"`
}

type RateLimitConfig struct {
	Requests int64         `env:"REQUESTS, default=0"`
	Period   time.Duration `env:"PERIOD, default=1s"`
}

type StorageConfig struct {
	Logs    LogsConfig    `env:", prefix=LOGS_"`
	Metrics MetricsConfig `env:", prefix=METRICS_"`
}

type LogConfig struct {
	Level string `env:"LEVEL, default=info"`
	Type  string `env:"TYPE, default=plain"`
}

type ServerConfig struct {
	Host              string          `env:"HOST, default=0.0.0.0"`
	Port              string          `env:"PORT, default=8080"`
	WriteTimeout      time.Duration   `env:"WRITE_TIMEOUT, default=10s"`
	ReadTimeout       time.Duration   `env:"READ_TIMEOUT, default=10s"`
	IdleTimeout       time.Duration   `env:"IDLE_TIMEOUT, default=120s"`
	ReadHeaderTimeout time.Duration   `env:"READ_HEADER_TIMEOUT, default=10s"`
	GlobalRateLimit   RateLimitConfig `env:", prefix=GLOBAL_RATE_LIMIT_"`
}

type LogsConfig struct {
	Type             string              `env:"TYPE, default=elasticsearch"`
	Elastic          ElasticSearchConfig `env:", prefix=ELASTIC_"`
	MaximumBytesSize int64               `env:"MAXIMUM_BYTES_SIZE, default=5242880"` // 5 мегабайт
	SamplingRate     float64             `env:"SAMPLING_RATE, default=1.0"`          // 1.0 означает приём 100% трафика
	PerUserRateLimit RateLimitConfig     `env:", prefix=PER_USER_RATE_LIMIT_"`
}

type ElasticSearchConfig struct {
	Addresses              []string      `env:"ADDRESSES"`
	Index                  string        `env:"INDEX"`
	User                   string        `env:"USER"`
	Password               string        `env:"PASSWORD"`
	Timeout                time.Duration `env:"TIMEOUT, default=10s"`
	CreateIndex            bool          `env:"CREATE_INDEX, default=true"`
	StartupCheckConnection bool          `env:"STARTUP_CHECK_CONNECTION, default=true"`
}

type MetricsConfig struct {
	Type             string                `env:"TYPE, default=victoriametrics"`
	Victoria         VictoriaMetricsConfig `env:", prefix=VICTORIA_"`
	MaximumBytesSize int64                 `env:"MAXIMUM_BYTES_SIZE, default=5242880"` // 5 megabytes
	SamplingRate     float64               `env:"SAMPLING_RATE, default=1.0"`          // 1.0 означает приём 100% трафика
	PerUserRateLimit RateLimitConfig       `env:", prefix=PER_USER_RATE_LIMIT_"`
}

type VictoriaMetricsConfig struct {
	URL         string        `env:"URL"`
	ExtraLabels []string      `env:"EXTRA_LABELS"`
	Timeout     time.Duration `env:"TIMEOUT, default=10s"`
}

type AuthConfig struct {
	AllowUnauthorized bool   `env:"ALLOW_UNAUTHORIZED, default=true"`
	Secret            string `env:"SECRET"`
	Header            string `env:"HEADER, default=x-access-token"`
}

func Get(ctx context.Context, useLocalEnvFile string) (*Config, error) {
	var c Config

	if useLocalEnvFile == "true" {
		// Загрузка переменных окружения из файла .env
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("error loading .env file: %v", err)
		}
	}

	err := envconfig.Process(ctx, &c)
	envconfig.OsLookuper()
	return &c, err
}
