package userservice

import (
	"context"
	"fmt"
	"time"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

func (s Service) createTokens(ctx context.Context, user models.User) (access models.JWTAccess, refresh models.JWTRefresh, err error) {
	const op = "./internal/service/userService/tokens.go.createTokens"

	accessTokenID := uuid.NewString()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JWTAccess{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpired)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        accessTokenID,
		},
		Type: "access",
		Role: user.Role,
		TokenID: accessTokenID,
	})

	accessSign, err := accessToken.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}
	access = accessToken.Claims.(models.JWTAccess)
	access.Sign = accessSign
	access.Type = "access"

	refreshTokenID := uuid.NewString()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JWTRefresh{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpired)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        refreshTokenID,
		},
		Type: "refresh",
		Role: user.Role,
		TokenID: refreshTokenID,
	})
	refreshSign, err := refreshToken.SignedString([]byte(s.config.SecretKey))

	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	refresh = refreshToken.Claims.(models.JWTRefresh)
	refresh.Sign = refreshSign
	refresh.Type = "refresh"

	return access, refresh, nil
}

func (s Service) RefreshTokens(ctx context.Context, tokenID string, refreshVersionCredentials int, expTime time.Time, user models.User) (access models.JWTAccess, refresh models.JWTRefresh, err error) {
	const op = "./internal/service/userService.RefreshTokens.go"

	err = s.invalidator.CheckTokenInBlackList(ctx, tokenID)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	user, err = s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	if user.VersionCredentials != refreshVersionCredentials {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, allerrors.ErrDifferentVersionCredentials)
	}

	err = s.invalidator.InvalidRefresh(ctx, tokenID, expTime)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	access, refresh, err = s.createTokens(ctx, user)
	if err != nil {
		return models.JWTAccess{}, models.JWTRefresh{}, fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil
}