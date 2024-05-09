package frontend

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"observability-agent/internal/auth"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
	"time"
)

const contextUserIDKey = "user_id"

// HTTPFrontend
// Реализация HTTP фронтенда
type HTTPFrontend struct {
	log     logger.Logger
	agent   *core.Agent
	config  *config.Config
	metrics *PromMetrics
	auth    auth.Verifier
}

// Start
// Запуск HTTP фронтенда
func (f *HTTPFrontend) Start() error {
	f.metrics = PreparePrometheusMetrics()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(f.config.Server.Timeout))
	r.Use(logger.Middleware(f.log))
	r.Use(f.AuthMiddleware)
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
func NewHTTP(agent *core.Agent, log logger.Logger, cfg *config.Config, auth auth.Verifier) (Frontend, error) {
	var front Frontend
	front = &HTTPFrontend{
		agent:  agent,
		log:    log,
		config: cfg,
		auth:   auth,
	}
	return front, nil
}

// AuthMiddleware
// Мидлвар для проверки авторизации и определения user ID
func (f *HTTPFrontend) AuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// для подсчета времени выполнения запроса
		reqStart := time.Now()
		// статус ответа по умолчанию
		reqStatus := http.StatusOK

		authHeader := f.config.Auth.Header
		allowUnauthorized := f.config.Auth.AllowUnauthorized

		defer func() {
			f.metrics.incomingRequests.WithLabelValues(
				fmt.Sprintf("%d", reqStatus),
				"logs",
			).Observe(time.Since(reqStart).Seconds())
		}()

		// Пробуем получить user ID
		userID, err := f.auth.GetUserID(r.Header.Get(authHeader))
		if err != nil {
			f.log.Warnf("Error getting user ID: %v", err)
		}

		// Если запрещены запросы от неавторизованных пользователей и не смогли получить user UD, то возвращаем ошибку
		if !allowUnauthorized && userID == 0 {
			reqStatus = http.StatusForbidden
			f.log.Warnf("Unauthorized requests is forbidden")
			http.Error(w, "Unauthorized requests is forbidden", reqStatus)
			return
		}

		// Добавляем user ID в контекст
		ctx := context.WithValue(r.Context(), contextUserIDKey, userID)
		// Добавляем контекст в запрос
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
