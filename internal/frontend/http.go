package frontend

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"net/http"
	"observability-agent/internal/auth"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
)

// HTTPFrontend
// Реализация HTTP фронтенда
type HTTPFrontend struct {
	log                   logger.Logger
	agent                 *core.Agent
	config                *config.Config
	metrics               *PromMetrics
	auth                  auth.Verifier
	globalLimiter         *stdlib.Middleware
	perUserLimiterLogs    *stdlib.Middleware
	perUserLimiterMetrics *stdlib.Middleware
}

// Start
// Запуск HTTP фронтенда
func (f *HTTPFrontend) Start() error {
	f.metrics = PreparePrometheusMetrics()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(f.config.Server.Timeout))
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
	f.log.Info("Starting agent")
	return http.ListenAndServe(fmt.Sprintf("%v:%v", f.config.Server.Host, f.config.Server.Port), r)
}

// NewHTTP
// Конструктор HTTPFrontend
func NewHTTP(agent *core.Agent,
	log logger.Logger,
	cfg *config.Config,
	auth auth.Verifier,
	globalLimiter, perUserLimiterMetrics, perUserLimiterLogs *stdlib.Middleware,
) (Frontend, error) {
	var front Frontend
	front = &HTTPFrontend{
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
