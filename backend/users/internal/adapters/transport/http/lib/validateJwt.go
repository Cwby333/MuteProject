package lib

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(token string) (jwt.MapClaims, error) {
	const op = "./internal/adapters/transport/http/lib/validateJwt"
	secretKey := os.Getenv("JWT_SECRET_KEY")

	t, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}, jwt.WithIssuer(os.Getenv("JWT_ISSUER")),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return jwt.MapClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	if !t.Valid {
		return jwt.MapClaims{}, fmt.Errorf("%s: %w", op, err)
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, fmt.Errorf("%s: %w", op, errors.New("wrong claims type"))
	}

	return claims, nil
}
