package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Server Server `yaml:"server" env-required:"true"`
	DB     DB     `yaml:"postgres" env-required:"true"`
	JWT    JWT    `yaml:"jwt" env-required:"true"`
	Redis  Redis  `yaml:"redis" env-required:"true"`
}

type Server struct {
	Address         string        `yaml:"address" env-required:"true"`
	IdleTimeout     time.Duration `yaml:"idle-timeout" env-required:"true"`
	ShutdownTimeout time.Duration `yaml:"shutdown-timeout" env-required:"true"`
}

// Postgres(pgxpool)
type DB struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     uint16 `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       string `yaml:"db" env-required:"true"`
	MaxConns int    `yaml:"max-conns" env-required:"true"`
	MinConns int    `yaml:"min-conns" env-required:"true"`
}

type Redis struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       int    `yaml:"db"`
}

type JWT struct {
	SecretKey string `yaml:"secret-key" env-required:"true"`
	Issuer    string `yaml:"issuer" env-required:"true"`

	JWTAccess struct {
		Expired time.Duration `yaml:"expired" env-required:"true"`
	} `yaml:"jwt-access" env-required:"true"`

	JWTRefresh struct {
		Expired time.Duration `yaml:"expired" env-required:"true"`
	} `yaml:"jwt-refresh" env-required:"true"`
}

func MustLoad() Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = cleanenv.ReadConfig(os.Getenv("CONFIG_PATH"), &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
