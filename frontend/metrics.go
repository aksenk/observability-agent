package frontend

import (
	"io"
	"net/http"
	"observability-agent/core"
)

func (f *HTTPFrontend) metricsReceiverHandler(w http.ResponseWriter, r *http.Request) {
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

}
