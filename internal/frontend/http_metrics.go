package frontend

import (
	"fmt"
	"net/http"
	"time"
)

// ExtendedResponseWriter
// Объект http.ResponseWriter с полем для хранения StatusCode
type ExtendedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader
// Вызывает стандартный метод WriteHeader + заполняет StatusCode
func (w *ExtendedResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// PrometheusMetricsMiddleware
// Мидлвар для автоматической регистрации prometheus метрик
func (f *HTTPFrontend) PrometheusMetricsMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// для подсчета времени выполнения запроса
		reqStart := time.Now()

		// Получаем тип запроса из контекста (добавляется в предыдущих middleware)
		requestType := r.Context().Value(RequestTypeContextField).(RequestType)

		rw := &ExtendedResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		// Собираем метрики только для запросов получения логов или метрик
		if requestType == TypeMetrics || requestType == TypeLogs {
			f.metrics.incomingRequests.WithLabelValues(
				fmt.Sprintf("%d", rw.StatusCode),
				fmt.Sprintf("%v", requestType),
			).Observe(time.Since(reqStart).Seconds())
		}
	}
	return http.HandlerFunc(fn)
}
