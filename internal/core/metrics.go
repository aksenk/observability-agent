package core

import "context"

// MetricsRequest
// Объект для хранения данных из входящего запроса с метриками
type MetricsRequest struct {
	Data   []byte
	Gzip   bool
	UserID int64
}

// MetricsSave
// Сохранение метрик в хранилище
func (a *Agent) MetricsSave(ctx context.Context, request *MetricsRequest) error {
	return a.metricsStorage.Save(ctx, request)
}

// MetricsIsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (a *Agent) MetricsIsSampled() bool {
	return a.metricsStorage.IsSampled()
}
