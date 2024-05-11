package metrics_storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
	"observability-agent/internal/sampler"
	"strings"
	"time"
)

// VMAgentClient
// Клиент для работы с victoriametrics agent
type VMAgentClient struct {
	url         string
	log         logger.Logger
	sampler     *sampler.Sampler
	extraLabels []string
	timeout     time.Duration
}

// IsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (c *VMAgentClient) IsSampled() bool {
	return c.sampler.IsSampled()
}

// Ping
// Проверка соединения
func (c *VMAgentClient) Ping(ctx context.Context) error {
	client := http.Client{}
	client.Timeout = c.timeout

	request, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		return err
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

	client := http.Client{}
	client.Timeout = c.timeout

	extraLabels := ""
	if c.extraLabels != nil {
		extraLabels = fmt.Sprintf("&extra_label=%v", strings.Join(c.extraLabels, "&extra_label="))
	}

	url := fmt.Sprintf("%v?extra_label=gambler_id=%v%v", c.url, metrics.UserID, extraLabels)

	c.log.Debugf("url: %v", url)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(metrics.Data))
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

// NewVMAgentClient
// Конструктор для VMAgentClient
func NewVMAgentClient(url string, extraLabels []string, timeout time.Duration, log logger.Logger, sampler *sampler.Sampler) (*VMAgentClient, error) {
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}
	// TODO добавить поддержку extra labels
	return &VMAgentClient{
		url:         url,
		log:         log,
		sampler:     sampler,
		extraLabels: extraLabels,
		timeout:     timeout,
	}, nil
}
