package frontend

import (
	"github.com/prometheus/client_golang/prometheus"
)

// metricsPrefix
// Префикс для всех отдаваемых prometheus-метрик
const metricsPrefix string = "observability_agent_"

// histogramBuckets
// Бакеты для хранения метрик типа histogram (при необходимости можно будет вынести в конфиг)
var histogramBuckets = []float64{0.1, 0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3, 3.5, 4, 5, 6, 7, 8, 9, 10, 15}

// PromMetrics
// Объект для хранения метрик Prometheus
type PromMetrics struct {
	incomingRequests *prometheus.HistogramVec
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
