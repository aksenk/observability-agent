package frontend

import (
	"net/http"
	"strings"
)

func (f *HTTPFrontend) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	logsErr := f.agent.PingLogsStorage(r.Context())
	metricsErr := f.agent.PingMetricsStorage(r.Context())

	w.Header().Set("Content-Type", "application/json")

	responseStatus := http.StatusOK
	responseText := strings.Builder{}

	responseText.WriteString("{")

	if logsErr != nil {
		responseText.WriteString("\"logs\": \"ERROR\",")
		responseStatus = http.StatusInternalServerError
	} else {
		responseText.WriteString("\"logs\": \"OK\",")
	}

	if metricsErr != nil {
		responseText.WriteString("\"metrics\": \"ERROR\"")
		responseStatus = http.StatusInternalServerError
	} else {
		responseText.WriteString("\"metrics\": \"OK\"")
	}

	responseText.WriteString("}\n")

	w.WriteHeader(responseStatus)
	w.Write([]byte(responseText.String()))
}
