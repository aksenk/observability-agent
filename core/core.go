package core

import (
	"context"
	"fmt"
)

// LogsStorage
// Интерфейс для работы с хранилищем логов
type LogsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, logs *LogsRequest) error
}

// MetricsStorage
// Интерфейс для работы с хранилищем метрик
type MetricsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, metrics *MetricsRequest) error
}

// MetricsRequest
// Объект для хранения данных из входящего запроса с метриками
type MetricsRequest struct {
	Data []byte
	Gzip bool
}

// LogsRequest
// Объект для хранения данных из входящего запроса с логами
type LogsRequest struct {
	Data   []byte
	Gzip   bool
	UserID int64
}

// Agent
// Объект основного приложения
type Agent struct {
	metricsStorage MetricsStorage
	logsStorage    LogsStorage
}

// NewAgent
// Конструктор для объекта Agent
func NewAgent(metricsStorage MetricsStorage, logsStorage LogsStorage) (*Agent, error) {
	return &Agent{
		metricsStorage: metricsStorage,
		logsStorage:    logsStorage,
	}, nil
}

// SaveMetrics
// Сохранение метрик в хранилище
func (a *Agent) SaveMetrics(ctx context.Context, request *MetricsRequest) error {
	return a.metricsStorage.Save(ctx, request)
}

// SaveLogs
// Сохранение логов в хранилище
func (a *Agent) SaveLogs(ctx context.Context, request *LogsRequest) error {
	// Проверяем что данные не закодированы, если нет заголовка 'Content-Encoding: gzip'
	if isGzipped(request.Data) && !request.Gzip {
		return fmt.Errorf("request body is gzipped but header 'Content-Encoding: gzip' is not exist")
	}

	// Проверяем что данные закодированы, если есть заголовок 'Content-Encoding: gzip'
	if !isGzipped(request.Data) && request.Gzip {
		return fmt.Errorf("request body is not gzipped but header 'Content-Encoding: gzip' is exist")
	}

	return a.logsStorage.Save(ctx, request)
}

//func (a *Agent) CheckJWT(ctx context.Context) error {
//	return nil
//}

// isGzipped
// Проверка данных на сжатие форматом gzip.
func isGzipped(data []byte) bool {
	// Проверяем, является ли первый байт 0x1f и второй байт 0x8b, что характерно для gzip формата.
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}
