package limiter

import (
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	redisLimiterDriver "github.com/ulule/limiter/v3/drivers/store/redis"
	"net/http"
	"observability-agent/internal/frontend"
	"time"
)

func NewPerUserLimiterMiddleware(period time.Duration, limit int64, redisClient *redis.Client) (*stdlib.Middleware, error) {
	// Если лимит выставлен в 0, то считаем, что ограничение запросов не требуется
	if limit == 0 {
		return nil, nil
	}

	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	// Храним в redis
	store, err := redisLimiterDriver.NewStore(redisClient)
	if err != nil {
		return nil, err
	}

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

	return middleware, nil
}

func UserLimitReachedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "User limit exceeded", http.StatusTooManyRequests)
}
