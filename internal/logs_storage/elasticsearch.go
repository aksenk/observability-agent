package logs_storage

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"net/http"
	"observability-agent/internal/core"
	"observability-agent/internal/logger"
	"observability-agent/internal/sampler"
	"time"
)

// ElasticSearchClient
// Объект для работы с ElasticSearch
type ElasticSearchClient struct {
	client    *elasticsearch.Client
	indexName string
	log       logger.Logger
	sampler   *sampler.Sampler
}

// ElasticSearchDocumentMetadataIndex
// Контент секции Index документа ElasticSearch
type ElasticSearchDocumentMetadataIndex struct {
	Index *string `json:"_index"`
	ID    *string `json:"_id"`
}

// ElasticSearchDocumentMetadata
// Объект для описания секции Index документа ElasticSearch
type ElasticSearchDocumentMetadata struct {
	Index *ElasticSearchDocumentMetadataIndex `json:"index"`
}

// ElasticSearchDocumentPayload
// Объект содержащий структуру итогового лога, который будет записан в ElasticSearch
type ElasticSearchDocumentPayload struct {
	UserID *int64  `json:"user_id"`
	Log    *string `json:"log"`
}

// NewElasticSearchClient
// Конструктор для ElasticSearchClient
func NewElasticSearchClient(ctx context.Context, addresses []string, username, password, indexName string, timeout time.Duration, log logger.Logger, sampler *sampler.Sampler) (*ElasticSearchClient, error) {
	if len(addresses) == 0 || addresses[0] == "" {
		return nil, fmt.Errorf("addresses is not defined")
	}
	if username == "" {
		return nil, fmt.Errorf("username is not defined")
	}
	if password == "" {
		return nil, fmt.Errorf("password is not defined")
	}
	if indexName == "" {
		return nil, fmt.Errorf("index name is not defined")
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout,
			//DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
		},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	client := &ElasticSearchClient{
		client:    es,
		indexName: indexName,
		log:       log,
		sampler:   sampler,
	}

	log.Info("Preparing elasticsearch index")
	err = client.Prepare(ctx)
	if err != nil {
		log.Fatalf("Error preparing storage :%v", err)
	}

	return client, nil
}

// IsSampled
// Проверка на семлирование (допуск только определенного процента трафика). Возвращает true если запрос должен быть отброшен
func (c *ElasticSearchClient) IsSampled() bool {
	return c.sampler.IsSampled()
}

// Ping
// Проверка соединения
func (c *ElasticSearchClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping()
	return err
}

// Close
// Закрытие соединения
func (c *ElasticSearchClient) Close(ctx context.Context) error {
	return nil
}

// Prepare
// Подготовка к работе
func (c *ElasticSearchClient) Prepare(ctx context.Context) error {
	_, err := c.client.Indices.Create(c.indexName)
	return err
}

// Save
// Сохранение логов в хранилище
func (c *ElasticSearchClient) Save(ctx context.Context, request *core.LogsRequest) error {
	bulkBody, err := c.prepareBulkBody(request)
	if err != nil {
		return fmt.Errorf("prepare bulk body: %s", err)
	}

	resp, err := c.client.Bulk(
		bytes.NewReader(bulkBody.Bytes()),
		c.client.Bulk.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	c.log.Debugf(resp.String())
	resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("%s", resp.Status())
	}

	return nil
}

// prepareBulkBody
// Формирование тела запроса для отправки данных в ElasticSearch методом Bulk
func (c *ElasticSearchClient) prepareBulkBody(request *core.LogsRequest) (*bytes.Buffer, error) {
	var bulkBody bytes.Buffer

	scanner := bufio.NewScanner(bytes.NewReader(request.Data))

	// считываем данные построчно и записываем их в буфер для bulkload
	// для каждой строки генерируем строку с метаданными и строку с логом
	// Elastic Bulk API: https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
	for scanner.Scan() {
		line := scanner.Text()

		c.log.Debugf("Processing line: %s", line)

		meta := &ElasticSearchDocumentMetadata{
			Index: &ElasticSearchDocumentMetadataIndex{
				Index: &c.indexName,
				ID:    generateRandomString(20),
			},
		}

		payload := ElasticSearchDocumentPayload{
			UserID: &request.UserID,
			Log:    &line,
		}

		if err := json.NewEncoder(&bulkBody).Encode(meta); err != nil {
			return nil, fmt.Errorf("error metadata encode: %s", err)
		}

		if err := json.NewEncoder(&bulkBody).Encode(payload); err != nil {
			return nil, fmt.Errorf("error body encode: %s", err)
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading data: %v", err)
	}

	// Данные для отправки методом Bulk должны заканчиваться символом перевода строки
	bulkBody.WriteByte('\n')

	return &bulkBody, nil
}

// generateRandomString
// Генерация случайной строки длиной length
func generateRandomString(length int) *string {
	// Вычисляем размер среза байтов, необходимый для указанной длины строки
	bytes := make([]byte, length)

	// Читаем криптографически безопасные случайные байты в срез bytes
	rand.Read(bytes)

	// Конвертируем срез байтов в строку base64
	s := base64.URLEncoding.EncodeToString(bytes)

	return &s
}
