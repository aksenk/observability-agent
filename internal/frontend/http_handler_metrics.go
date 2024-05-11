package frontend

import (
	"io"
	"net/http"
	"observability-agent/internal/core"
)

// metricsReceiverHandler
// Обработчик запроса на запись метрик
func (f *HTTPFrontend) metricsReceiverHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Пробуем получить user ID из заголовка (добавляется в middleware)
	userID, err := GetUserID(r)
	if err != nil {
		f.log.Error("Error getting user ID: %v", err)
		http.Error(w, "Error getting user ID", http.StatusInternalServerError)
		return
	}

	// проверяем, нужно ли отбросить запрос на основе семплирования
	if f.agent.MetricsIsSampled() {
		f.log.Debugf("Metrics request are sampled")
		http.Error(w, "Metrics request are sampled", http.StatusTooManyRequests)
		return
	}

	// Сразу проверяем что заявленный размер запроса не превышает максимально допустимый
	if r.ContentLength > f.config.Storage.Metrics.MaximumBytesSize {
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", http.StatusRequestEntityTooLarge)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		f.log.Errorf("Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Если запрос пустой, то возвращаем ошибку
	if len(body) == 0 {
		f.log.Warn("Request with empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Если фактический размер запроса превышает максимально разрешенный размер, то возвращаем ошибку
	if int64(len(body)) > f.config.Storage.Metrics.MaximumBytesSize {
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", http.StatusRequestEntityTooLarge)
		return
	}

	// Подготавливаем объект для передачи в агент
	metrics := &core.MetricsRequest{
		Data:   body,
		UserID: userID,
	}

	// Если запрос закодирован, то устанавливаем флаг, указывающий на это
	if r.Header.Get("Content-Encoding") == "gzip" {
		metrics.Gzip = true
	}

	// Сохраняем полученные данные
	err = f.agent.MetricsSave(ctx, metrics)
	if err != nil {
		f.log.Errorf("Error receiving metrics: %v", err)
		http.Error(w, "Error receiving metrics", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Metrics received successfully\n"))
	w.WriteHeader(http.StatusOK)
	return
}
