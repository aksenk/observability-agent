package frontend

import (
	"io"
	"net/http"
	"observability-agent/internal/core"
)

// metricsReceiverHandler
// Обработчик запроса на запись метрик
func (f *HTTPFrontend) metricsReceiverHandler(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > f.config.Metrics.MaximumBytesSize {
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

	// TODO аус

	metrics := &core.MetricsRequest{
		Data: body,
	}

	if r.Header.Get("Content-Encoding") == "gzip" {
		metrics.Gzip = true
	}

	err = f.agent.SaveMetrics(r.Context(), metrics)
	if err != nil {
		f.log.Errorf("Error receiving metrics: %v", err)
		http.Error(w, "Error receiving metrics", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Metrics received successfully\n"))
	w.WriteHeader(http.StatusOK)
	return
}
