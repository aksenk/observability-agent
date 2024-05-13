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

// App
// Объект основного приложения, содержит в себе методы для работы с входящими логами и метриками
type App struct {
	metricsStorage MetricsStorage
	logsStorage    LogsStorage
}

// NewApp
// Конструктор для объекта App
func NewApp(metricsStorage MetricsStorage, logsStorage LogsStorage) (*App, error) {
	return &App{
		metricsStorage: metricsStorage,
		logsStorage:    logsStorage,
	}, nil
}
