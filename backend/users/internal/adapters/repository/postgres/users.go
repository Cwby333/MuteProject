package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	allerrors "github.com/Cwby333/user-microservice/internal/allErrors"
	"github.com/Cwby333/user-microservice/internal/models"
	"github.com/jackc/pgx/v5"
)

func (pg Postgres) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	const op = "./internal/adapters/postgres/users.go.CreateUser"
	const query = `INSERT INTO users(username, password, email, role) VALUES($1, $2, $3, $4)`
	const query2 = `SELECT id FROM users WHERE username = $1`

	userDTO := ToUserDTO(user)

	_, err := pg.Pool.Exec(ctx, query,
		userDTO.Username,
		userDTO.Password,
		userDTO.Email,
		userDTO.Role,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_key") {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrUsernameExists)
		}

		if strings.Contains(err.Error(), "users_email_key") {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrEmailExists)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := pg.Pool.Query(ctx, query2, user.Username)

	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return models.User{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	user.ID = id

	return user, nil
}

func (pg Postgres) GetUserByID(ctx context.Context, ID string) (models.User, error) {
	const op = "./internal/adapters/postgres/users.go.GetUserByUsername"
	const query = `SELECT id, username, password, email, role, version_credentials FROM users WHERE id = $1`

	rows, err := pg.Pool.Query(ctx, query, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrUserNotExists)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	DTO, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[UserDTO])
	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrUserNotExists)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user := DTOToUser(DTO)

	return user, nil
}

func (pg Postgres) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	const op = "./internal/adapters/postgres/users.go.GetUserByUsername"
	const query = `SELECT id, username, password, email, role, version_credentials FROM users WHERE username = $1`

	rows, err := pg.Pool.Query(ctx, query, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrUserNotExists)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	DTO, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[UserDTO])
	if err != nil {
		if err.Error() == "no rows in result set" {
			return models.User{}, fmt.Errorf("%s: %w", op, allerrors.ErrUserNotExists)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user := DTOToUser(DTO)

	return user, nil
}

func (pg Postgres) GetAllUsers(ctx context.Context) ([]models.User, error) {
	const op = "./internal/adapters/postgres/users.go.GetUserByUsername"
	const query = `SELECT id, username, password, email, role, version_credentials FROM users`

	rows, err := pg.Pool.Query(ctx, query)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sliceDTO, err := pgx.CollectRows(rows, pgx.RowToStructByName[UserDTO])
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	sliceUsers := make([]models.User, 0, len(sliceDTO))

	for i := range sliceDTO {
		sliceUsers = append(sliceUsers, DTOToUser(sliceDTO[i]))
	}

	return sliceUsers, nil
}

func (pg Postgres) DeleteUserByID(ctx context.Context, ID string) error {
	const op = "./internal/adapters/postgres/users.go.DeleteUserByID"
	const query = `DELETE FROM users WHERE id = $1`

	rows, err := pg.Pool.Query(ctx, query, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("%s: %w", op, err)
	}
	rows.Close()

	return nil
}

func (pg Postgres) UpdateUserByID(ctx context.Context, ID string, newUserInfo models.User) (models.User, error) {
	const op = "./internal/adapters/postgres/users.go.UpdateUserByID"
	const query = `UPDATE users SET username = $1, role = $2, password = $3, email = $4, version_credentials = $5 WHERE id = $5`

	rows, err := pg.Pool.Query(ctx, query,
		newUserInfo.Username,
		newUserInfo.Role,
		newUserInfo.Password,
		newUserInfo.Email,
		newUserInfo.VersionCredentials,
		ID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return newUserInfo, nil
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	rows.Close()

	return newUserInfo, nil
}
