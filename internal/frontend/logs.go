package frontend

import (
	"fmt"
	"io"
	"net/http"
	"observability-agent/internal/core"
	"time"
)

// logsReceiverHandler
// Обработчик запроса на получение логов
func (f *HTTPFrontend) logsReceiverHandler(w http.ResponseWriter, r *http.Request) {
	// для подсчета времени выполнения запроса
	reqStart := time.Now()
	// статус ответа по умолчанию
	reqStatus := http.StatusOK

	// по завершении запроса регистрируем метрику
	defer func() {
		sinceStart := time.Since(reqStart).Seconds()
		f.metrics.incomingRequests.WithLabelValues(
			fmt.Sprintf("%d", reqStatus),
			"logs",
		).Observe(sinceStart)
	}()

	// Сразу проверяем что размер запроса не превышает максимально допустимый
	if r.ContentLength > f.config.Logs.MaximumBytesSize {
		reqStatus = http.StatusRequestEntityTooLarge
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", reqStatus)
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

	if len(body) == 0 {
		f.log.Warn("Request with empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Пробуем получить user ID
	userID, err := f.agent.GetUserID(r.Header.Get("user-id"))
	if err != nil {
		f.log.Warnf("Error getting user ID: %v", err)
	}

	if !f.config.Auth.AllowUnauthorized && userID == 0 {
		reqStatus = http.StatusForbidden
		f.log.Warnf("Unauthorized incomingRequests is forbidden")
		http.Error(w, "Unauthorized incomingRequests is forbidden", reqStatus)
		return
	}

	request := &core.LogsRequest{
		Data:   body,
		UserID: userID,
	}

	if r.Header.Get("Content-Encoding") == "gzip" {
		request.Gzip = true
	}

	err = f.agent.SaveLogs(r.Context(), request)
	if err != nil {
		reqStatus = http.StatusInternalServerError
		f.log.Errorf("Error receiving request: %v", err)
		http.Error(w, "Error receiving request", reqStatus)
		return
	}

	w.Write([]byte("Logs received successfully\n"))
	w.WriteHeader(reqStatus)
	return
}
