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

	cfg, err := config.Get(ctx)
	if err != nil {
		log.Fatalf("Error init config: %v", err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Logger init error: %v", err)
		os.Exit(1)
	}

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

	var metricsStorage core.MetricsStorage
	switch cfg.Metrics.Type {
	case "victoriametrics":
		metricsStorage, err = metrics_storage.NewClient(cfg.Metrics.Victoria.URL, cfg.Metrics.Victoria.ExtraLabels, log)
	default:
		log.Fatalf("Unknown metrics storage type: %v", cfg.Metrics.Type)
	}

	agent, err := core.NewAgent(metricsStorage, logsStorage)
	if err != nil {
		log.Fatalf("Error init agent: %v", err)
	}

	front, err := frontend.NewHTTP(agent, log)
	if err != nil {
		log.Fatalf("Error init frontend: %v", err)
	}

	err = front.Start()
	if err != nil {
		log.Fatalf("Error start frontend: %v", err)
	}
}
