package frontend

import (
	"io"
	"net/http"
	"observability-agent/core"
)

// logsReceiverHandler
// Обработчик запроса на получение логов
func (f *HTTPFrontend) logsReceiverHandler(w http.ResponseWriter, r *http.Request) {
	// Сразу проверяем что размер запроса не превышает максимально допустимый
	if r.ContentLength > f.config.Logs.MaximumBytesSize {
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
		f.log.Warnf("Unauthorized requests is forbidden")
		http.Error(w, "Unauthorized requests is forbidden", http.StatusForbidden)
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
		f.log.Errorf("Error receiving request: %v", err)
		http.Error(w, "Error receiving request", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Logs received successfully\n"))
	w.WriteHeader(http.StatusOK)
	return
}
