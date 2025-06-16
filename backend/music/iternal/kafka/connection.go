package kafka

import (
	"log"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func NewConsumer(brokers, groupID string) *ckafka.Consumer {
	c, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Не удалось создать Kafka consumer: %v", err)
	}
	return c
}
