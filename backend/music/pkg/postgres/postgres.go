package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	Username string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Database string `env:"DB_NAME"`
}

func runMigrations(connString string) {

	pgxConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Не удалось распарсить строку подключения: %v", err)
	}

	sqlDB := stdlib.OpenDB(*pgxConfig)
	defer sqlDB.Close()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Не удалось получить рабочую директорию: %v", err)
	}
	cwd = filepath.Dir(cwd)
	migrationsPath := "file://" + filepath.Join(cwd, "db", "migrations")

	driver, err := pgxmigrate.WithInstance(sqlDB, &pgxmigrate.Config{})
	if err != nil {
		log.Fatalf("Не удалось создать драйвер для миграций: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Не удалось создать экземпляр мигратора: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}

	log.Println("Миграции успешно применены!")
}

func New(config Config) (*pgx.Conn, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	runMigrations(connString)

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return conn, nil
}
