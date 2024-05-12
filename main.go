package main

import (
	"context"
	"errors"
	"observability-agent/internal/auth"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/frontend"
	"observability-agent/internal/limiter"
	"observability-agent/internal/logger"
	"observability-agent/internal/logs_storage"
	"observability-agent/internal/metrics_storage"
	"observability-agent/internal/sampler"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
TODO list
+ receiving metrics
+ receiving logs
+ sampling
+ jwt
+ prometheus metrics
+ add user label to metrics from jwt
+ custom metric labels
+ global rate limit for instance
+ rate limit per user (in-memory)
+ global prometheus metrics middleware
+ healthcheck
+ graceful shutdown
+\- timeouts (в эластике работает странно, фактический таймаут х4 от указанного)
+\- logger (кривой какой-то)

- distributed rate limit per user (redis)
- circuit breaker
- client metrics histogram support
- open telemetry metrics?
- logs contract + validation
- metrics contract + validation
- tests
- autotests
*/

func main() {
	ctx, stopCtx := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Инициализация логгера.
	log := logger.New()

	useLocalEnvFile := os.Getenv("USE_LOCAL_ENV_FILE")
	if useLocalEnvFile == "true" {
		log.Info("Read environment variables from .env file")
	}

	// Инициализация конфига.
	cfg, err := config.Get(ctx, useLocalEnvFile)
	if err != nil {
		log.Fatalf("Error init config: %v", err)
	}

	if err := log.SetLevel(cfg.Log.Level); err != nil {
		log.Fatalf("Incorrect log level: %v", err)
	}

	if err := log.SetFormatter(cfg.Log.Type); err != nil {
		log.Fatalf("Incorrect log formatter: %v", err)
	}

	// Инициализация механизма авторизации.
	jwtVerifier, err := auth.NewJWTVerifier(cfg.Auth.Secret)
	if err != nil {
		log.Fatalf("Error init jwt verifier: %v", err)
	}

	// Инициализация механизма семплирования для логов
	logsSampler, err := sampler.New(cfg.Storage.Logs.SamplingRate)
	if err != nil {
		log.Fatalf("Error init logs sampler: %v", err)
	}

	// Инициализация механизма семплирования для метрик
	metricsSampler, err := sampler.New(cfg.Storage.Metrics.SamplingRate)
	if err != nil {
		log.Fatalf("Error init metrics sampler: %v", err)
	}

	// Инициализация глобального ограничителя запросов
	globalRateLimiter := limiter.NewGlobalLimiterMiddleware(
		cfg.Server.GlobalRateLimit.Period,
		cfg.Server.GlobalRateLimit.Requests)
	if globalRateLimiter == nil {
		log.Info("Global rate limiter is not configured")
	} else {
		log.Info("Global rate limiter: maximum %v requests per %v",
			cfg.Server.GlobalRateLimit.Requests, cfg.Server.GlobalRateLimit.Period)
	}

	// Инициализация ограничителя запросов по пользователям для логов
	logsRateLimiter := limiter.NewPerUserLimiterMiddleware(
		cfg.Storage.Logs.PerUserRateLimit.Period,
		cfg.Storage.Logs.PerUserRateLimit.Requests,
		frontend.UserIDContextField)
	if logsRateLimiter == nil {
		log.Info("Per user logs rate limiter is not configured")
	} else {
		log.Info("Per user logs rate limiter: maximum %v requests per %v",
			cfg.Storage.Logs.PerUserRateLimit.Requests, cfg.Storage.Logs.PerUserRateLimit.Period)
	}

	// Инициализация ограничителя запросов по пользователям для логов
	metricsRateLimiter := limiter.NewPerUserLimiterMiddleware(
		cfg.Storage.Metrics.PerUserRateLimit.Period,
		cfg.Storage.Metrics.PerUserRateLimit.Requests,
		frontend.UserIDContextField)
	if metricsRateLimiter == nil {
		log.Info("Per user metrics rate limiter is not configured")
	} else {
		log.Info("Per user metrics rate limiter: maximum %v requests per %v",
			cfg.Storage.Metrics.PerUserRateLimit.Requests, cfg.Storage.Metrics.PerUserRateLimit.Period)
	}

	// Инициализация хранилища для логов.
	var logsStorage core.LogsStorage
	switch cfg.Storage.Logs.Type {
	case "elasticsearch":
		logsStorage, err = logs_storage.NewElasticSearchClient(
			ctx,
			&cfg.Storage.Logs.Elastic,
			log,
			logsSampler)
		if err != nil {
			log.Fatalf("Error init elastic storage: %v", err)
		}
	default:
		log.Fatalf("Unknown logs storage type: %v", cfg.Storage.Logs.Type)
	}

	// Инициализация хранилища для метрик.
	var metricsStorage core.MetricsStorage
	switch cfg.Storage.Metrics.Type {
	case "victoriametrics":
		metricsStorage, err = metrics_storage.NewVMAgentClient(
			&cfg.Storage.Metrics.Victoria,
			log,
			metricsSampler)
		if err != nil {
			log.Fatalf("Error init victoria metrics storage: %v", err)
		}
	default:
		log.Fatalf("Unknown metrics storage type: %v", cfg.Storage.Metrics.Type)
	}

	// Инициализация основного приложения
	agent, err := core.NewAgent(metricsStorage, logsStorage)
	if err != nil {
		log.Fatalf("Error init agent: %v", err)
	}

	// Инициализация фронтенда для приложения.
	front, err := frontend.NewHTTP(agent, log, cfg, jwtVerifier, globalRateLimiter, metricsRateLimiter, logsRateLimiter)
	if err != nil {
		log.Fatalf("Error init frontend: %v", err)
	}

	go func() {
		<-signals
		log.Info("Starting graceful shutdown")

		// контекст для graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// если таймаут истек, то завершаем приложение с ошибкой
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal("Graceful shutdown timeout. Forcing exit")
			}
		}()

		// специально переобъявляем эту переменную, что избежать возможного data race с такой же переменной с основной функции
		if err := front.Stop(shutdownCtx); err != nil {
			log.Fatalf("Shutdown error: %v", err)
		}
		log.Info("Graceful shutdown completed")

		stopCtx()
	}()

	// Старт фронтенда.
	go front.Start(ctx)

	<-ctx.Done()

	log.Info("Application is stopped")
}
