package validation

import (
	"auth/internals/config"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func Validate(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		slog.Debug("validation failed", "reason", "no auth header")
		return "", false
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		slog.Debug("validation failed", "reason", "no prefix")
		return "", false
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if tokenString == "" {
		slog.Debug("validation failed", "reason", "no token")
		return "", false
	}

	secret := config.ReadFromEnv().JwtSigningKey
	if secret == "" {
		slog.Error("no secret provided")
		return "", false
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method)
		}

		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || token == nil || !token.Valid {
		slog.Debug("validation failed", "err", err, "token", token)
		return "", false
	}

	if claims.UserID == "" {
		slog.Debug("validation failed", "reason", "no user id")
		return "", false
	}

	return claims.UserID, true
}
