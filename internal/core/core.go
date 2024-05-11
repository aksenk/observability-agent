package core

import (
	"context"
)

// LogsStorage
// Интерфейс для работы с хранилищем логов
type LogsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, request *LogsRequest) error
	IsSampled() bool
}

// MetricsStorage
// Интерфейс для работы с хранилищем метрик
type MetricsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, request *MetricsRequest) error
	IsSampled() bool
}

// Agent
// Объект основного приложения, содержит в себе методы для работы с входящими логами и метриками
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
