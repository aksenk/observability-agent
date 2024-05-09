package metrics_storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
	"observability-agent/internal/sampler"
)

// VMAgentClient
// Клиент для работы с victoriametrics agent
type VMAgentClient struct {
	url     string
	log     logger.Logger
	sampler *sampler.Sampler
}

// IsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (c *VMAgentClient) IsSampled() bool {
	return c.sampler.IsSampled()
}

// Ping
// Проверка соединения
func (c *VMAgentClient) Ping(ctx context.Context) error {
	return nil
}

// Close
// Закрытие соединения
func (c *VMAgentClient) Close(ctx context.Context) error {
	return nil
}

// Prepare
// Подготовка к работе
func (c *VMAgentClient) Prepare(ctx context.Context) error {
	return nil
}

// Save
// Сохранение логов в хранилище
func (c *VMAgentClient) Save(ctx context.Context, metrics *core.MetricsRequest) error {

	// TODO timeout
	client := http.Client{}

	request, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(metrics.Data))
	if err != nil {
		return err
	}

	if metrics.Gzip {
		request.Header.Set("Content-Encoding", "gzip")
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %v", response.StatusCode)
	}

	return nil
}

// NewClient
// Конструктор для VMAgentClient
func NewClient(url string, extraLabels []string, log logger.Logger) (*VMAgentClient, error) {
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}
	// TODO добавить поддержку extra labels
	return &VMAgentClient{
		url: url,
		log: log,
	}, nil
}
