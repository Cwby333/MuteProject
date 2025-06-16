package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type SongAction struct {
	Action  string `json:"action"`
	UserID  string `json:"user_id"`
	TrackID string `json:"track_id"`
}

func main() {
	// 0) Попытка загрузить .env
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠ warning: не удалось загрузить %s: %v", envPath, err)
	} else {
		log.Printf("✅ загружен файл %s", envPath)
	}

	// 1) Подключение к Postgres (rest опущен, без изменений)...
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = os.Getenv("POSTGRES_USER")
	}
	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = os.Getenv("POSTGRES_PASSWORD")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = os.Getenv("POSTGRES_DB")
	}
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)

	var db *pgx.Conn
	var err error
	for attempt := 1; attempt <= 5; attempt++ {
		db, err = pgx.Connect(context.Background(), dbURL)
		if err == nil {
			break
		}
		log.Printf("⚠ попытка %d: не удалось подключиться: %v", attempt, err)
		time.Sleep(time.Duration(attempt*2) * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}
	defer db.Close(context.Background())
	log.Println("✅ Connected to Postgres:", dbURL)

	// 2) Producer Kafka
	kafkaBrokers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBrokers})
	if err != nil {
		log.Fatalf("❌ Не удалось создать продьюсер: %v", err)
	}
	defer p.Close()
	log.Println("✅ Kafka producer, brokers:", kafkaBrokers)

	// События доставки
	go func() {
		for e := range p.Events() {
			if m, ok := e.(*kafka.Message); ok {
				if m.TopicPartition.Error != nil {
					log.Printf("❌ Delivery failed: %v", m.TopicPartition)
				} else {
					log.Printf("✅ Delivered to %v", m.TopicPartition)
				}
			}
		}
	}()

	log.Println("🔄 Бесконечный цикл обработки deferred_tasks")
	for {
		tx, err := db.Begin(context.Background())
		if err != nil {
			log.Fatalf("Begin tx error: %v", err)
		}

		rows, err := tx.Query(context.Background(),
			`SELECT id, topic, data FROM deffered_tasks ORDER BY created_at ASC FOR UPDATE SKIP LOCKED LIMIT 20`)
		if err != nil {
			tx.Rollback(context.Background())
			log.Fatalf("Select tasks error: %v", err)
		}

		type taskRow struct {
			id, topic, data string
		}
		var tasks []taskRow
		for rows.Next() {
			var tr taskRow
			if err := rows.Scan(&tr.id, &tr.topic, &tr.data); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			tasks = append(tasks, tr)
		}
		rows.Close()

		if len(tasks) == 0 {
			tx.Rollback(context.Background())
			time.Sleep(5 * time.Second)
			continue
		}

		for _, tr := range tasks {
			// —————— НОВОЕ: логируем весь JSON и распаршенный объект ——————
			log.Printf("🔔 Обнаружена задача: id=%s, topic=%s, raw data=%s", tr.id, tr.topic, tr.data)
			var action SongAction
			if err := json.Unmarshal([]byte(tr.data), &action); err != nil {
				log.Printf("⚠ не удалось распарсить data для задачи %s: %v", tr.id, err)
			} else {
				log.Printf("└─ parsed SongAction: action=%s, user_id=%s, track_id=%s",
					action.Action, action.UserID, action.TrackID)
			}
			// ——————————————————————————————————————————————————————————

			// Produce в Kafka с retry
			var prodErr error
			for attempt := 1; attempt <= 3; attempt++ {
				prodErr = p.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &tr.topic, Partition: kafka.PartitionAny},
					Key:            []byte(tr.id),
					Value:          []byte(tr.data),
				}, nil)
				if prodErr == nil {
					break
				}
				log.Printf("⚠ Produce attempt %d for %s failed: %v", attempt, tr.id, prodErr)
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			}
			if prodErr != nil {
				log.Fatalf("❌ Failed to produce %s: %v", tr.id, prodErr)
			}

			// Удаляем задачу
			if _, err := tx.Exec(context.Background(), `DELETE FROM deffered_tasks WHERE id = $1`, tr.id); err != nil {
				tx.Rollback(context.Background())
				log.Fatalf("Failed to delete task %s: %v", tr.id, err)
			}
			log.Printf("✅ Задача %s успешно отправлена и удалена", tr.id)
		}

		if err := tx.Commit(context.Background()); err != nil {
			log.Fatalf("Tx commit error: %v", err)
		}
	}
}
