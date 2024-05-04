package frontend

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"observability-agent/core"
	"observability-agent/logger"
)

type Frontend interface {
	Start() error
}

type HTTPFrontend struct {
	log   logger.Logger
	agent *core.Agent
}

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

func NewHTTP(agent *core.Agent, log logger.Logger) (Frontend, error) {
	var front Frontend
	front = &HTTPFrontend{
		agent: agent,
		log:   log,
	}
	return front, nil
}
