package config

import (
	"context"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Server  ServerConfig  `env:", prefix=SERVER_"`
	Logs    LogsConfig    `env:", prefix=LOGS_"`
	Metrics MetricsConfig `env:", prefix=METRICS_"`
	Auth    AuthConfig    `env:", prefix=AUTH_"`
	Log     LogConfig     `env:", prefix=LOG_"`
}

type LogConfig struct {
	Level string `env:"LEVEL, default=info"`
}

type ServerConfig struct {
	Host string `env:"HOST, default=0.0.0.0"`
	Port int    `env:"PORT, default=8080"`
}

type LogsConfig struct {
	Type             string              `env:"TYPE, default=elasticsearch"`
	Elastic          ElasticSearchConfig `env:", prefix=ELASTIC_"`
	MaximumBytesSize int64               `env:"MAXIMUM_BYTES_SIZE, default=5242880"`
}

type ElasticSearchConfig struct {
	URL      string `env:"URL"`
	Index    string `env:"INDEX"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
}

type MetricsConfig struct {
	Type             string                `env:"TYPE, default=victoriametrics"`
	Victoria         VictoriaMetricsConfig `env:", prefix=VICTORIA_"`
	MaximumBytesSize int64                 `env:"MAXIMUM_BYTES_SIZE, default=5242880"`
}

type VictoriaMetricsConfig struct {
	URL         string   `env:"URL"`
	ExtraLabels []string `env:"EXTRA_LABELS"`
}

type AuthConfig struct {
	Enable bool   `env:"PORT, default=true"`
	Secret string `env:"SECRET"`
}

func Get(ctx context.Context) (*Config, error) {
	var c Config
	err := envconfig.Process(ctx, &c)
	return &c, err
}
