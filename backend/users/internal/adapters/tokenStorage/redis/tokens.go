package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/redis/go-redis/v9"
)

const (
	refreshStorage = "refresh:"
)

func (r Redis) InvalidRefresh(ctx context.Context, tokenID string, unixTime time.Time) error {
	const op = "./internal/adapters/tokenStorage/redis/tokens.go.InvalidRefresh"

	cmd := r.client.HSet(ctx, refreshStorage, tokenID, 1)
	if cmd.Err() != nil {
		return fmt.Errorf("%s: %w", op, cmd.Err())
	}

	cmd2 := r.client.HExpireAt(ctx, refreshStorage, unixTime, tokenID)
	if cmd2.Err() != nil {
		return fmt.Errorf("%s: %w", op, cmd.Err())
	}

	return nil
}

func (r Redis) CheckTokenInBlackList(ctx context.Context, tokenID string) error {
	const op = "./internal/adapters/tokenStorage/redis/tokens.go.CheckTokenInBlackList"

	cmd := r.client.HGet(ctx, refreshStorage, tokenID)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return nil
		}

		return fmt.Errorf("%s: %w", op, cmd.Err())
	}

	return allerrors.ErrTokenInBlackList
}
