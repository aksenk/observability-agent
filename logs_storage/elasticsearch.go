package logs_storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"observability-agent/core"
	"observability-agent/logger"
)

type ElasticSearchClient struct {
	client    *elasticsearch.Client
	indexName string
	log       logger.Logger
}

func NewElasticSearchClient(ctx context.Context, addresses []string, username, password, indexName string, log logger.Logger) (*ElasticSearchClient, error) {
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
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	client := &ElasticSearchClient{
		client:    es,
		indexName: indexName,
		log:       log,
	}

	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("Error get elasticsearch info: %v", err)
	}

	err = client.Prepare(ctx)
	if err != nil {
		log.Fatalf("Error preparing storage :%v", err)
	}

	return client, nil
}

func (c *ElasticSearchClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping()
	return err
}

func (c *ElasticSearchClient) Close(ctx context.Context) error {
	return nil
}

func (c *ElasticSearchClient) Prepare(ctx context.Context) error {
	_, err := c.client.Indices.Create(c.indexName)
	return err
}

func (c *ElasticSearchClient) Save(ctx context.Context, logs core.LogsRequest) error {
	document := struct {
		UserID int64  `json:"id"`
		Log    string `json:"log"`
	}{
		UserID: 123,
		Log:    string(logs.Data),
	}
	d, _ := json.Marshal(document)
	resp, err := c.client.Index(c.indexName, bytes.NewReader(d))
	fmt.Print(resp)
	return err
}
