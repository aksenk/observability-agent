package frontend

import (
	"fmt"
	"io"
	"net/http"
	"observability-agent/core"
	"strconv"
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

	userID, err := getUserID(r)
	if err != nil {
		f.log.Warnf("Error getting user ID: %v", err)
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

func getUserID(r *http.Request) (int64, error) {
	i := r.Header.Get("user-id")
	if i == "" {
		return 0, fmt.Errorf("user-id is not found")
	}
	userID, err := strconv.ParseInt(i, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("user-id is not a number")
	}
	return userID, nil
}
