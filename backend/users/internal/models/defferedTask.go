package models

import "time"

// Task for sender microrservice

type DefferedTask struct {
	ID        string
	Topic     string
	Data      []byte
	CreatedAt time.Time
}
