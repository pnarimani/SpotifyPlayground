package signup

import (
	"auth/internals/config"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SignupResponse struct {
	UserID string `json:"user_id"`
	Jwt    string `json:"jwt"`
}

func Signup() (SignupResponse, error) {
	userID := uuid.NewString()

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})

	tstr, err := t.SignedString([]byte(config.ReadFromEnv().JwtSigningKey))
	if err != nil {
		return SignupResponse{}, fmt.Errorf("failed to sign jwt token, err: %w", err)
	}

	return SignupResponse{
		UserID: userID,
		Jwt:    tstr,
	}, nil
}
