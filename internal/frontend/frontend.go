package frontend

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
)

const metricsPrefix string = "observability_agent_"

var histogramBuckets = []float64{0.1, 0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3, 3.5, 4, 5, 6, 7, 8, 9, 10, 15}

// PromMetrics
// Объект для хранения метрик Prometheus
type PromMetrics struct {
	incomingRequests *prometheus.HistogramVec
}

// Frontend
// Интерфейс для реализации фронтенда
type Frontend interface {
	Start() error
}

// HTTPFrontend
// Реализация HTTP фронтенда
type HTTPFrontend struct {
	log     logger.Logger
	agent   *core.Agent
	config  *config.Config
	metrics *PromMetrics
}

// Start
// Запуск HTTP фронтенда
func (f *HTTPFrontend) Start() error {
	f.metrics = PreparePrometheusMetrics()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(f.config.Server.Timeout))
	//r.Use(middleware.Logger)
	r.Use(logger.Middleware(f.log))
	r.Handle("/metrics", promhttp.Handler())
	r.Route("/api/v1/", func(r chi.Router) {
		r.Post("/logs/elasticsearch/bulk", f.logsReceiverHandler)
		r.Put("/metrics/victoriametrics/import", f.metricsReceiverHandler)
	})
	f.log.Info("Starting agent")
	return http.ListenAndServe(":8080", r)
}

// NewHTTP
// Конструктор HTTPFrontend
func NewHTTP(agent *core.Agent, log logger.Logger, cfg *config.Config) (Frontend, error) {
	var front Frontend
	front = &HTTPFrontend{
		agent:  agent,
		log:    log,
		config: cfg,
	}
	return front, nil
}

// PreparePrometheusMetrics
// Функция для инициализации метрик Prometheus
func PreparePrometheusMetrics() *PromMetrics {
	incomingRequests := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metricsPrefix + "incoming_requests",
			Buckets: histogramBuckets,
		},
		[]string{"status", "type"},
	)

	prometheus.MustRegister(incomingRequests)

	return &PromMetrics{
		incomingRequests: incomingRequests,
	}
}
