package main

import (
	"context"
	"fmt"
	"log"
	"observability-agent/config"
	"observability-agent/core"
	"observability-agent/frontend"
	"observability-agent/logger"
	"observability-agent/logs_storage"
	"observability-agent/metrics_storage"
	"os"
)

func main() {
	ctx := context.Background()

	// Инициализация конфига.
	cfg, err := config.Get(ctx)
	if err != nil {
		log.Fatalf("Error init config: %v", err)
	}

	// Инициализация логгера.
	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		fmt.Printf("Logger init error: %v", err)
		os.Exit(1)
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
			log)
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
		metricsStorage, err = metrics_storage.NewClient(cfg.Metrics.Victoria.URL, cfg.Metrics.Victoria.ExtraLabels, log)
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
