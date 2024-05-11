package frontend

import (
	"context"
	"fmt"
	"net/http"
)

// UserIDContextField
// В какое поле контекста будет помещен определенный user ID (0 если не определен)
const UserIDContextField = "user_id"

// AuthMiddleware
// Мидлвар для проверки авторизации, определения user ID и записи его в заголовок
func (f *HTTPFrontend) AuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Получаем тип запроса из контекста (добавляется в предыдущих middleware)
		requestType := r.Context().Value(RequestTypeContextField).(RequestType)

		// Не проверяем аутентификацию для запросов не на получение логов или метрик
		if requestType != TypeMetrics && requestType != TypeLogs {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := f.config.Auth.Header
		allowUnauthorized := f.config.Auth.AllowUnauthorized

		// Пробуем получить user ID
		userID, err := f.auth.GetUserID(r.Header.Get(authHeader))
		if err != nil {
			f.log.Warnf("Error getting user ID: %v", err)
		}

		// Если запрещены запросы от неавторизованных пользователей и не смогли получить user UD, то возвращаем ошибку
		if !allowUnauthorized && userID == 0 {
			f.log.Warnf("Unauthorized requests is forbidden")
			http.Error(w, "Unauthorized requests is forbidden", http.StatusForbidden)
			return
		}

		// Добавляем user ID в context
		ctx := context.WithValue(r.Context(), UserIDContextField, userID)
		// Передаём управление дальше с новым контекстом
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func GetUserID(r *http.Request) (int64, error) {
	// Пробуем получить user ID из контекста (добавляется в middleware)
	userID := r.Context().Value(UserIDContextField).(int64)
	if userID == 0 {
		return 0, fmt.Errorf("can not find user id in context")
	}
	return userID, nil
}
