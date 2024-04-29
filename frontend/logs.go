package frontend

import (
	"io"
	"net/http"
	"observability-agent/core"
)

func (f *HTTPFrontend) logsReceiverHandler(w http.ResponseWriter, r *http.Request) {
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

	logs := core.LogsRequest{
		Data: body,
	}

	err = f.agent.SaveLogs(r.Context(), logs)
	if err != nil {
		f.log.Error("Error receiving logs: %v", err)
		http.Error(w, "Error receiving logs", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Logs received successfully\n"))
	w.WriteHeader(http.StatusOK)
	return
}
