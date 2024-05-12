package metrics_storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"observability-agent/internal/config"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
	"observability-agent/internal/sampler"
	"strings"
	"time"
)

// VMAgentClient
// Клиент для работы с victoriametrics agent
type VMAgentClient struct {
	client  *http.Client
	url     string
	log     logger.Logger
	sampler *sampler.Sampler
	timeout time.Duration
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

	URL := fmt.Sprintf("%v%v", c.url, metrics.UserID)

	c.log.Debugf("URL = %v", URL)

	request, err := http.NewRequest(http.MethodPost, URL, bytes.NewReader(metrics.Data))
	if err != nil {
		return err
	}

	if metrics.Gzip {
		request.Header.Set("Content-Encoding", "gzip")
	}

	response, err := c.client.Do(request)
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
func NewVMAgentClient(vmConfig *config.VictoriaMetricsConfig, log logger.Logger, sampler *sampler.Sampler) (*VMAgentClient, error) {
	if vmConfig.URL == "" {
		return nil, fmt.Errorf("url is empty")
	}
	if vmConfig.Timeout.Seconds() <= 0 {
		return nil, fmt.Errorf("incorrect timeout: %v", vmConfig.Timeout)
	}

	var URL string
	// Добавляем в URL дополнительные лейблы, указанные в конфиге
	// и дополнительный лейбл gambler_id, который будет добавляться к каждому запросу,
	// но значение этого лейбла будет подставляться только при отправке запроса
	extraLabels := ""
	if vmConfig.ExtraLabels != nil {
		extraLabels = fmt.Sprintf("?extra_label=%v", strings.Join(vmConfig.ExtraLabels, "&extra_label="))
		URL = fmt.Sprintf("%v%v&extra_label=gambler_id=", vmConfig.URL, extraLabels)
	} else {
		URL = fmt.Sprintf("%v?extra_label=gambler_id=", vmConfig.URL)
	}

	return &VMAgentClient{
		client: &http.Client{
			Timeout: vmConfig.Timeout,
		},
		url:     URL,
		log:     log,
		sampler: sampler,
		timeout: vmConfig.Timeout,
	}, nil
}
