package main

import (
	"context"
	"log"
	config "music/iternal/config"
	kafkaconsumer "music/iternal/kafka"
	"music/iternal/musicserver"
	"music/iternal/storage"
	postgres "music/pkg/postgres"
)

func main() {
	config.LoadEnv()

	dbConfig := postgres.Config{
		Host:     config.Get("DB_HOST"),
		Port:     config.Get("DB_PORT"),
		Username: config.Get("DB_USER"),
		Password: config.Get("DB_PASSWORD"),
		Database: config.Get("DB_NAME"),
	}

	dbConn, err := postgres.New(dbConfig)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbConn.Close(context.Background())
	log.Println("Успешное подключение к базе данных!")

	s3Config := storage.Config{
		AccessKey: config.Get("S3_ACCESS_KEY"),
		SecretKey: config.Get("S3_SECRET_KEY"),
		Region:    config.Get("S3_REGION"),
		Endpoint:  config.Get("S3_ENDPOINT"),
		Bucket:    config.Get("S3_BUCKET"),
	}

	s3Client, err := storage.NewS3Client(s3Config)
	if err != nil {
		log.Fatalf("Не удалось подключиться к S3-хранилищу: %v", err)
	}
	log.Println("Успешное подключение к S3-хранилищу!")

	// Запуск Kafka-консьюмера в отдельной горутине
	ctx := context.Background()
	go kafkaconsumer.Start(
		ctx,
		dbConn,
		config.Get("KAFKA_BOOTSTRAP_SERVERS"),
		"music-consumer-group",
		"songs_actions",
	)
	log.Println("🔄 Kafka consumer запущен")

	address := ":8080"
	if err := musicserver.StartServer(address, dbConn, s3Client); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}

}
