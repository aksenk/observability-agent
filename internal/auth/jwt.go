package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

// Verifier
// Интерфейс для работы с авторизацией клиентов
type Verifier interface {
	GetUserID(jwtToken string) (int64, error)
}

// JWTVerifier
// Реализация интерфейса Verifier для работы с JWT
type JWTVerifier struct {
	secret      []byte
	userIDField string
}

// NewJWTVerifier
// Создает новый JWTVerifier
func NewJWTVerifier(secret string) (*JWTVerifier, error) {
	if secret == "" {
		return nil, fmt.Errorf("secret is empty")
	}
	return &JWTVerifier{
		secret:      []byte(secret),
		userIDField: "gambler_id",
	}, nil
}

// GetUserID
// Получает ID пользователя из JWT токена
func (v *JWTVerifier) GetUserID(jwtToken string) (int64, error) {
	if jwtToken == "" {
		return 0, fmt.Errorf("token is empty")
	}

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return v.secret, nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		userID, ok := claims[v.userIDField]
		if ok {
			return int64(userID.(float64)), nil
		}
		return 0, fmt.Errorf("user id is not found")
	} else {
		return 0, fmt.Errorf("invalid token")
	}
}
