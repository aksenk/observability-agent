package core

import "context"

type LogsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, logs LogsRequest) error
}

type MetricsStorage interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Prepare(ctx context.Context) error
	Save(ctx context.Context, metrics MetricsRequest) error
}

type MetricsRequest struct {
	Data []byte
	Gzip bool
}

type LogsRequest struct {
	Data []byte
}

type Agent struct {
	metricsStorage MetricsStorage
	logsStorage    LogsStorage
}

func NewAgent(metricsStorage MetricsStorage, logsStorage LogsStorage) (*Agent, error) {
	return &Agent{
		metricsStorage: metricsStorage,
		logsStorage:    logsStorage,
	}, nil
}

func (a *Agent) SaveMetrics(ctx context.Context, metrics MetricsRequest) error {
	return a.metricsStorage.Save(ctx, metrics)
}

func (a *Agent) SaveLogs(ctx context.Context, logs LogsRequest) error {
	return a.logsStorage.Save(ctx, logs)
}

//func (a *Agent) CheckJWT(ctx context.Context) error {
//	return nil
//}
