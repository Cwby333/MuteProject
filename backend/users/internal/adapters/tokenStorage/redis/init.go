package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       int
}

func New(ctx context.Context, cfg Config) (Redis, error) {
	const op = "./internal/adapters/tokenStorage/redis/init.go.New"

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	tag := client.Ping(ctx)
	if tag.Err() != nil {
		return Redis{}, fmt.Errorf("%s: %w", op, tag.Err())
	}

	return Redis{
		client: client,
	}, nil
}
