package frontend

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"observability-agent/config"
	"observability-agent/core"
	"observability-agent/logger"
)

// Frontend
// Интерфейс для реализации фронтенда
type Frontend interface {
	Start() error
}

// HTTPFrontend
// Реализация HTTP фронтенда
type HTTPFrontend struct {
	log    logger.Logger
	agent  *core.Agent
	config *config.Config
}

// Start
// Запуск фронтенда
func (f *HTTPFrontend) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
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
