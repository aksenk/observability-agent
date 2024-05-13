package frontend

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"log"
	"net/http"
	"observability-agent/internal/auth"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
)

// HTTPFrontend
// Реализация HTTP фронтенда
type HTTPFrontend struct {
	server                *http.Server
	log                   logger.Logger
	agent                 *core.App
	config                *config.Config
	metrics               *PromMetrics
	auth                  auth.Verifier
	globalLimiter         *stdlib.Middleware
	perUserLimiterLogs    *stdlib.Middleware
	perUserLimiterMetrics *stdlib.Middleware
}

func (f *HTTPFrontend) prepareRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(logger.Middleware(f.log))
	r.Use(f.DetectRequestTypeMiddleware)
	r.Use(f.PrometheusMetricsMiddleware)
	r.Use(f.AuthMiddleware)
	r.Use(f.LimiterMiddleware(f.globalLimiter, f.perUserLimiterLogs, f.perUserLimiterMetrics))
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthcheck", f.HealthcheckHandler)

	r.Route("/api/v1/", func(r chi.Router) {
		r.Post("/logs/elasticsearch/bulk", f.logsReceiverHandler)
		r.Put("/metrics/victoriametrics/import", f.metricsReceiverHandler)
	})
	return r
}

// Start
// Запуск HTTP фронтенда
func (f *HTTPFrontend) Start(ctx context.Context) {
	f.metrics = PreparePrometheusMetrics()
	f.server.Handler = f.prepareRouter()

	f.log.Info("Starting web server")

	err := f.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Error start frontend: %v", err)
	}
}

// Stop
// Остановка HTTP фронтенда
func (f *HTTPFrontend) Stop(ctx context.Context) error {
	err := f.server.Shutdown(ctx)
	return err
}

// NewHTTP
// Конструктор HTTPFrontend
func NewHTTP(agent *core.App,
	log logger.Logger,
	cfg *config.Config,
	auth auth.Verifier,
	globalLimiter, perUserLimiterMetrics, perUserLimiterLogs *stdlib.Middleware,
) (Frontend, error) {
	var front Frontend

	server := &http.Server{
		Addr:              cfg.Server.Host + ":" + cfg.Server.Port,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	front = &HTTPFrontend{
		server:                server,
		agent:                 agent,
		log:                   log,
		config:                cfg,
		auth:                  auth,
		globalLimiter:         globalLimiter,
		perUserLimiterMetrics: perUserLimiterMetrics,
		perUserLimiterLogs:    perUserLimiterLogs,
	}
	return front, nil
}
