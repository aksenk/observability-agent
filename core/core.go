package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strconv"
)

// LogsStorage
// Интерфейс для работы с хранилищем логов
type LogsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, request *LogsRequest) error
}

// MetricsStorage
// Интерфейс для работы с хранилищем метрик
type MetricsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, request *MetricsRequest) error
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

	// Если данные закодированы, то распаковываем их
	if request.Gzip {
		var err error
		request.Data, err = unGzip(request.Data)
		if err != nil {
			return fmt.Errorf("error ungzip request body: %w", err)
		}
	}

	return a.logsStorage.Save(ctx, request)
}

func (a *Agent) GetUserID(rawData string) (int64, error) {
	if rawData == "" {
		return 0, fmt.Errorf("user-id is not found")
	}
	userID, err := strconv.ParseInt(rawData, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("user-id is not a number")
	}
	return userID, nil
}

// isGzipped
// Проверка данных на сжатие форматом gzip.
func isGzipped(data []byte) bool {
	// Проверяем, является ли первый байт 0x1f и второй байт 0x8b, что характерно для gzip формата.
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// unGzip
// Распаковка данных из gzip формата.
func unGzip(inputData []byte) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(inputData))
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	outputData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	return outputData, nil
}
