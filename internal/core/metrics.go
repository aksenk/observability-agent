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
func (a *App) MetricsSave(ctx context.Context, request *MetricsRequest) error {
	return a.metricsStorage.Save(ctx, request)
}

// MetricsIsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (a *App) MetricsIsSampled() bool {
	return a.metricsStorage.IsSampled()
}

func (a *App) PingMetricsStorage(ctx context.Context) error {
	return a.metricsStorage.Ping(ctx)
}
