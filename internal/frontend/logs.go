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

	//time.Sleep(10 * time.Second)

	// по завершении запроса регистрируем метрику
	defer func() {
		sinceStart := time.Since(reqStart).Seconds()
		f.metrics.incomingRequests.WithLabelValues(
			fmt.Sprintf("%d", reqStatus),
			"logs",
		).Observe(sinceStart)
	}()

	// проверяем, нужно ли отбросить запрос на основе семплирования
	if f.agent.LogsIsSampled() {
		f.log.Debugf("Logs request are sampled")
		reqStatus = http.StatusTooManyRequests
		http.Error(w, "Logs request are sampled", reqStatus)
		return
	}

	// Сразу проверяем что заявленный размер запроса не превышает максимально допустимый
	if r.ContentLength > f.config.Logs.MaximumBytesSize {
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

	// Вычитываем контент запроса
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
		f.log.Warn("Request with empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Если фактический размер запроса превышает максимально разрешенный размер, то возвращаем ошибку
	if int64(len(body)) > f.config.Logs.MaximumBytesSize {
		reqStatus = http.StatusRequestEntityTooLarge
		f.log.Warnf("Request with size %d is too big", r.ContentLength)
		http.Error(w, "Request is too big", reqStatus)
		return
	}

	// Подготавливаем объект для передачи в агент
	request := &core.LogsRequest{
		Data:   body,
		UserID: userID,
	}

	// Если запрос закодирован, то устанавливаем флаг, указывающий на это
	if r.Header.Get("Content-Encoding") == "gzip" {
		request.Gzip = true
	}

	// Сохраняем полученные данные
	err = f.agent.LogsSave(r.Context(), request)
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
