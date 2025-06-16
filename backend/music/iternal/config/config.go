package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
}

func Get(key string) string {
	return os.Getenv(key)
}
