package limiter

import (
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"observability-agent/internal/frontend"
	"time"
)

func NewPerUserLimiterMiddleware(period time.Duration, limit int64, key string) *stdlib.Middleware {
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

	// Функция получения ключа для лимитера,
	// возвращаем ключ вида userID
	middleware.KeyGetter = func(r *http.Request) string {
		userID, _ := frontend.GetUserID(r)
		return string(userID)
	}

	// Указываем свой обработчик для ситуаций превышения количества запросов
	middleware.OnLimitReached = UserLimitReachedHandler

	return middleware
}

func UserLimitReachedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "User limit exceeded", http.StatusTooManyRequests)
}
