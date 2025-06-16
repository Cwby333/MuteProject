package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Task struct {
	Action  string `json:"action"`
	UserID  string `json:"user_id"`
	TrackID string `json:"track_id"`
}

func Start(ctx context.Context, db *pgx.Conn, brokers, groupID, topic string) {
	c := NewConsumer(brokers, groupID)
	defer c.Close()

	if err := c.Subscribe(topic, nil); err != nil {
		log.Fatalf("Не удалось подписаться на %s: %v", topic, err)
	}
	log.Printf("✅ Kafka consumer subscribed to %s", topic)

	for {
		msg, err := c.ReadMessage(500 * time.Millisecond)
		if err != nil {
			if kafkaErr, ok := err.(ckafka.Error); ok && kafkaErr.Code() == ckafka.ErrTimedOut {
				continue
			}
			log.Printf("Kafka error: %v", err)
			continue
		}

		var t Task
		if err := json.Unmarshal(msg.Value, &t); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		if t.Action != "like" && t.Action != "dislike" {
			continue
		}

		tx, err := db.Begin(ctx)
		if err != nil {
			log.Printf("Begin tx error: %v", err)
			continue
		}

		switch t.Action {
		case "like":
			_, err = tx.Exec(ctx,
				`INSERT INTO liked_music(id, track_id, user_id)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (track_id, user_id) DO NOTHING`,
				uuid.New(), t.TrackID, t.UserID,
			)
		case "dislike":
			_, err = tx.Exec(ctx,
				`DELETE FROM liked_music
				 WHERE track_id = $1 AND user_id = $2`,
				t.TrackID, t.UserID,
			)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("DB exec error (%s): %v", t.Action, err)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			log.Printf("Tx commit error: %v", err)
			continue
		}

		if t.Action == "like" {
			log.Printf("🎉 Like processed: user=%s track=%s", t.UserID, t.TrackID)
		} else {
			log.Printf("💔 Dislike processed: user=%s track=%s", t.UserID, t.TrackID)
		}
	}
}
