package main

import (
	"context"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/frontend"
	"observability-agent/internal/logger"
	"observability-agent/internal/logs_storage"
	"observability-agent/internal/metrics_storage"
	"observability-agent/internal/sampler"
)

/*
TODO list
+ receiving metrics
+ receiving logs
- validating logs
- add user label to metrics from jwt
+ logger
- circuit breaker
+ sampling
- distributed rate limit
  - per user
  - total?
- jwt
+ prometheus metrics
- opentelemetry metrics?
- logs contract
- metrics contract
- graceful shutdown
- tests
- autotests
*/

func main() {
	ctx := context.Background()

	// Инициализация логгера.
	log := logger.New()

	// Инициализация конфига.
	cfg, err := config.Get(ctx)
	if err != nil {
		log.Fatalf("Error init config: %v", err)
	}

	if err := log.SetLevel(cfg.Log.Level); err != nil {
		log.Fatalf("Incorrect log level: %v", err)
	}

	if err := log.SetFormatter(cfg.Log.Type); err != nil {
		log.Fatalf("Incorrect log formatter: %v", err)
	}

	// Инициализация механизма семплирования для логов
	logsSampler, err := sampler.New(cfg.Logs.SamplingRate)
	if err != nil {
		log.Fatalf("Error init logs sampler: %v", err)
	}

	// Инициализация механизма семплирования для метрик
	metricsSampler, err := sampler.New(cfg.Metrics.SamplingRate)
	if err != nil {
		log.Fatalf("Error init metrics sampler: %v", err)
	}

	// Инициализация хранилища для логов.
	var logsStorage core.LogsStorage
	switch cfg.Logs.Type {
	case "elasticsearch":
		logsStorage, err = logs_storage.NewElasticSearchClient(
			ctx,
			[]string{cfg.Logs.Elastic.URL},
			cfg.Logs.Elastic.User,
			cfg.Logs.Elastic.Password,
			cfg.Logs.Elastic.Index,
			log,
			logsSampler)
		if err != nil {
			log.Fatalf("Error init log storage: %v", err)
		}
	default:
		log.Fatalf("Unknown logs storage type: %v", cfg.Logs.Type)
	}

	// Инициализация хранилища для метрик.
	var metricsStorage core.MetricsStorage
	switch cfg.Metrics.Type {
	case "victoriametrics":
		metricsStorage, err = metrics_storage.NewVMAgentClient(
			cfg.Metrics.Victoria.URL,
			cfg.Metrics.Victoria.ExtraLabels,
			log,
			metricsSampler)
	default:
		log.Fatalf("Unknown metrics storage type: %v", cfg.Metrics.Type)
	}

	// Инициализация основного приложения
	agent, err := core.NewAgent(metricsStorage, logsStorage)
	if err != nil {
		log.Fatalf("Error init agent: %v", err)
	}

	// Инициализация фронтенда для приложения.
	front, err := frontend.NewHTTP(agent, log, cfg)
	if err != nil {
		log.Fatalf("Error init frontend: %v", err)
	}

	// Старт фронтенда.
	err = front.Start()
	if err != nil {
		log.Fatalf("Error start frontend: %v", err)
	}
}
