package frontend

import (
	"context"
	"net/http"
	"strings"
)

func (f *HTTPFrontend) DetectRequestTypeMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path

		var requestType RequestType

		if strings.Contains(url, "/api/v1/metrics/") {
			requestType = TypeMetrics
		} else if strings.Contains(url, "/api/v1/logs/") {
			requestType = TypeLogs
		} else if url == "/metrics" || url == "/healthcheck" {
			requestType = TypeService
		} else {
			requestType = TypeUnknown
		}

		ctx := context.WithValue(r.Context(), RequestTypeContextField, requestType)
		next.ServeHTTP(w, r.WithContext(ctx))

	}
	return http.HandlerFunc(fn)
}
