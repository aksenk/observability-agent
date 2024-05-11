package core

import (
	"context"
	"fmt"
)

// LogsRequest
// Объект для хранения данных из входящего запроса с логами
type LogsRequest struct {
	Data   []byte
	Gzip   bool
	UserID int64
}

// LogsSave
// Сохранение логов в хранилище
func (a *Agent) LogsSave(ctx context.Context, request *LogsRequest) error {
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

// LogsIsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (a *Agent) LogsIsSampled() bool {
	return a.logsStorage.IsSampled()
}
