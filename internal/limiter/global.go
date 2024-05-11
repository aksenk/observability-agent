package limiter

import (
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"time"
)

func NewGlobalLimiterMiddleware(period time.Duration, limit int64) *stdlib.Middleware {
	// Если лимит выставлен в 0, то считаем, что ограничение запросов не требуется
	if limit == 0 {
		return nil
	}
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	// Храним в памяти
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	middleware := stdlib.NewMiddleware(instance)

	// Функция получения ключа для лимитера
	// Для всех запросов возвращаем единый ключ (т.к. это лимитер всех запросов приложения)
	middleware.KeyGetter = func(r *http.Request) string {
		return "request"
	}

	// Указываем свой обработчик для ситуаций превышения количества запросов
	middleware.OnLimitReached = GlobalLimitReachedHandler

	return middleware
}

func GlobalLimitReachedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Global limit exceeded", http.StatusTooManyRequests)
}
