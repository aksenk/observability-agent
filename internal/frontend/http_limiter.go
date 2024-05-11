package frontend

import (
	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"net/http"
)

func (f *HTTPFrontend) LimiterMiddleware(globalLimiter, logsLimiter, metricsLimiter *stdlib.Middleware) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Получаем тип запроса из контекста (добавляется в предыдущих middleware)
			requestType := r.Context().Value(RequestTypeContextField)

			var globalContext limiter.Context
			var globalKey string
			// если globalLimiter задан
			if globalLimiter != nil {
				var err error
				// получаем ключ для глобального ограничителя
				globalKey = globalLimiter.KeyGetter(r)
				// получаем текущее значение лимитера
				// пока без увеличения счетчика, потому что запрос может упереться в лимит по логам
				globalContext, err = globalLimiter.Limiter.Peek(r.Context(), globalKey)
				// в случае ошибки - вызываем обработчик для ошибок
				if err != nil {
					globalLimiter.OnError(w, r, err)
					return
				}
			}

			switch requestType.(RequestType) {

			// запросы на приём логов
			case TypeLogs:
				logrus.Infof("type logs")
				// если globalLimiter задан
				if globalLimiter != nil {
					// если глобальный лимит исчерпан
					// вызываем обработчик для таких ситуаций (в данном случае отдать 429)
					if globalContext.Reached {
						globalLimiter.OnLimitReached(w, r)
						return
					}
				}

				// если logsLimiter задан
				if logsLimiter != nil {
					// получаем ключ для ограничителя запросов логов
					logsKey := logsLimiter.KeyGetter(r)
					// получаем текущее значение лимитера и инкрементим его
					logsContext, err := logsLimiter.Limiter.Get(r.Context(), logsKey)
					// в случае ошибки - вызываем обработчик для ошибок
					if err != nil {
						logsLimiter.OnError(w, r, err)
						return
					}
					// если лимит по ключу (id пользователя) исчерпан
					// вызываем обработчик для таких ситуаций (в нашем случае отдать 429)
					if logsContext.Reached {
						logsLimiter.OnLimitReached(w, r)
						return
					}
				}

				if globalLimiter != nil {
					// Если запрос не упёрся в лимитер логов, то также увеличиваем счетчик глобального лимитера
					globalLimiter.Limiter.Increment(r.Context(), globalKey, 1)
				}

				// если лимиты еще не исчерпаны или лимитеры не заданы - передаём управление дальше
				next.ServeHTTP(w, r)

			// запросы на приём метрик
			case TypeMetrics:
				logrus.Infof("type metrics")

				// если globalLimiter задан
				if globalLimiter != nil {
					// если глобальный лимит исчерпан
					// вызываем обработчик для таких ситуаций (в нашем случае отдать 429)
					if globalContext.Reached {
						globalLimiter.OnLimitReached(w, r)
						return
					}
				}

				// если metricsLimiter задан
				if metricsLimiter != nil {
					// получаем ключ для ограничителя запросов логов
					metricsKey := metricsLimiter.KeyGetter(r)
					// получаем текущее значение лимитера и инкрементим его
					metricsContext, err := metricsLimiter.Limiter.Get(r.Context(), metricsKey)
					// это не должно никогда выполняться
					if err != nil {
						metricsLimiter.OnError(w, r, err)
						return
					}
					// если лимит по ключу (id пользователя) исчерпан
					// вызываем обработчик для таких ситуаций (в данном случае отдать 429)
					if metricsContext.Reached {
						metricsLimiter.OnLimitReached(w, r)
						return
					}
				}

				if globalLimiter != nil {
					// Если запрос не упёрся в лимитер метрик, то также увеличиваем счетчик глобального лимитера
					globalLimiter.Limiter.Increment(r.Context(), globalKey, 1)
				}

				// если лимиты еще не исчерпаны или лимитеры не заданы - передаём управление дальше
				next.ServeHTTP(w, r)

			// для всех остальных определенных запросов сразу передаём управление дальше
			case TypeService:
				logrus.Infof("type service")

				next.ServeHTTP(w, r)
				return
			// для запросов, где не смогли определить тип запроса отдаём ошибку
			// (на всякий случай, чтобы не пропустить лишнего, если вдруг в будущем будут другие запросы)
			default:
				logrus.Infof("type other")

				http.Error(w, "unknown request type", http.StatusBadRequest)
				return
			}
		})
	}
}
