package postgres

import (
	"context"
	"fmt"

	"github.com/Cwby333/user-microservice/internal/models"
)

func (pg Postgres) Create(ctx context.Context, task models.DefferedTask) error {
	const op = "./internal/adapter/repository/postgres/defferedTasks.go.Create"
	const query = `INSERT INTO deffered_tasks(topic, data, created_at) VALUES($1, $2, $3)`

	_, err := pg.Pool.Exec(ctx, query, task.Topic, task.Data, task.CreatedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
