package internal

import (
	"context"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("AAAAAAAAAAAAds")

type Claims struct {
	Login string
	Role  string
	jwt.RegisteredClaims
}

type ctxKey struct{}

// Создаем JWT
func GenerateJWT(login, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Login: login,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "example.com/myapp",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 5)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Читаем метаданные
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// Обновляем контекст с учетом метаданных
func SaveClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// Получаем метаданные из контекста
func GetClaims(ctx context.Context) *Claims {
	v := ctx.Value(ctxKey{})
	claims, ok := v.(*Claims)
	if !ok {
		return nil
	}
	return claims
}
