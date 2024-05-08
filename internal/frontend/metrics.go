package frontend

import (
	"fmt"
	"io"
	"net/http"
	"observability-agent/internal/core"
	"time"
)

// metricsReceiverHandler
// Обработчик запроса на запись метрик
func (f *HTTPFrontend) metricsReceiverHandler(w http.ResponseWriter, r *http.Request) {
	// для подсчета времени выполнения запроса
	reqStart := time.Now()
	// статус ответа по умолчанию
	reqStatus := http.StatusOK

	// по завершении запроса регистрируем метрику
	defer func() {
		sinceStart := time.Since(reqStart).Seconds()
		f.metrics.incomingRequests.WithLabelValues(
			fmt.Sprintf("%d", reqStatus),
			"metrics",
		).Observe(sinceStart)
	}()

	// Сразу проверяем что заявленный размер запроса не превышает максимально допустимый
	if r.ContentLength > f.config.Metrics.MaximumBytesSize {
		reqStatus = http.StatusRequestEntityTooLarge
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", reqStatus)
		return
	}

	// Пробуем получить user ID
	userID, err := f.agent.GetUserID(r.Header.Get("user-id"))
	if err != nil {
		f.log.Warnf("Error getting user ID: %v", err)
	}

	// Если запрещены запросы от неавторизованных пользователей и не смогли получить user UD, то возвращаем ошибку
	if !f.config.Auth.AllowUnauthorized && userID == 0 {
		reqStatus = http.StatusForbidden
		f.log.Warnf("Unauthorized requests is forbidden")
		http.Error(w, "Unauthorized requests is forbidden", reqStatus)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		reqStatus = http.StatusInternalServerError
		f.log.Errorf("Error reading body: %v", err)
		http.Error(w, "Error reading body", reqStatus)
		return
	}
	defer r.Body.Close()

	// Если запрос пустой, то возвращаем ошибку
	if len(body) == 0 {
		reqStatus = http.StatusBadRequest
		f.log.Warn("Request with empty body")
		http.Error(w, "empty body", reqStatus)
		return
	}

	// Если фактический размер запроса превышает максимально разрешенный размер, то возвращаем ошибку
	if int64(len(body)) > f.config.Metrics.MaximumBytesSize {
		reqStatus = http.StatusRequestEntityTooLarge
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", reqStatus)
		return
	}

	// Подготавливаем объект для передачи в агент
	metrics := &core.MetricsRequest{
		Data: body,
	}

	// Если запрос закодирован, то устанавливаем флаг, указывающий на это
	if r.Header.Get("Content-Encoding") == "gzip" {
		metrics.Gzip = true
	}

	// Сохраняем полученные данные
	err = f.agent.SaveMetrics(r.Context(), metrics)
	if err != nil {
		reqStatus = http.StatusInternalServerError
		f.log.Errorf("Error receiving metrics: %v", err)
		http.Error(w, "Error receiving metrics", reqStatus)
		return
	}

	w.Write([]byte("Metrics received successfully\n"))
	w.WriteHeader(reqStatus)
	return
}
