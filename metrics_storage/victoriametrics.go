package metrics_storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"observability-agent/core"
	"observability-agent/logger"
)

type VMAgentClient struct {
	url string
	log logger.Logger
}

func (c *VMAgentClient) Ping(ctx context.Context) error {
	return nil
}

func (c *VMAgentClient) Close(ctx context.Context) error {
	return nil
}

func (c *VMAgentClient) Prepare(ctx context.Context) error {
	return nil
}

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

	//c.log.Debugf("Sending following body:\n%v", string(metrics.Data))

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %v", response.StatusCode)
	}

	return nil
}

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
