package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"

	gojson "github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
)

const (
	usersStorage = "users:"
)

type DTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (r Redis) Set(ctx context.Context, userID string, user models.User) error {
	const op = "./internal/adapters/tokenStorage/redis/users.go.Set"

	dto := DTO{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
		Email:    user.Email,
		Role:     user.Role,
	}
	data, err := gojson.Marshal(dto)
	if err != nil {
		slog.Info("user cache, gojson marshal", slog.String("error", err.Error()))

		return fmt.Errorf("%s: %w", op, err)
	}

	cmd := r.client.HSet(ctx, usersStorage, userID, data)
	if cmd.Err() != nil {
		return fmt.Errorf("%s: %w", op, cmd.Err())
	}

	return nil
}

func (r Redis) Get(ctx context.Context, userID string) (models.User, error) {
	const op = "./internal/adapters/tokenStorage/redis/users.go.Get"

	cmd := r.client.HGet(ctx, usersStorage, userID)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrNotFoundInCache)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, cmd.Err())
	}

	res, err := cmd.Result()
	if err != nil {
		slog.Info("cmd result", slog.String("error", err.Error()))
	}

	var dto DTO
	err = gojson.Unmarshal([]byte(res), &dto)
	if err != nil {
		slog.Info("gojson unmarshal", slog.String("error", err.Error()))
	}

	user := models.User{
		ID:       dto.ID,
		Username: dto.Username,
		Password: dto.Password,
		Email:    dto.Email,
		Role:     dto.Role,
	}

	return user, nil
}

func (r Redis) Delete(ctx context.Context, userID string) error {
	const op = "./internal/adapters/tokenStorage/redis/users.go.Delete"

	cmd := r.client.HDel(ctx, usersStorage, userID)
	if cmd.Err() != nil {
		return fmt.Errorf("%s: %w", op, cmd.Err())
	}

	return nil
}
