package frontend

import "context"

// Frontend
// Интерфейс для реализации фронтенда
type Frontend interface {
	Start(ctx context.Context)
	Stop(ctx context.Context) error
}
