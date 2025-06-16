package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

type PostgresTest struct {
	pg   Postgres
	isUp bool
	mustSkipTest bool
}

type Config struct {
	Host     string
	Port     uint16
	User     string
	Password string
	DB       string
	MaxConns int
	MinConns int
}

func New(ctx context.Context, cfg Config) (Postgres, error) {
	const op = "./internal/adapters/postgres/init.go"

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d&pool_min_conns=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		cfg.MaxConns,
		cfg.MinConns,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	pgxpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	err = pgxpool.Ping(ctx)
	if err != nil {
		return Postgres{}, fmt.Errorf("%s: %w", op, err)
	}

	return Postgres{
		Pool: pgxpool,
	}, nil
}

func (p Postgres) Close() {
	p.Pool.Close()
}
