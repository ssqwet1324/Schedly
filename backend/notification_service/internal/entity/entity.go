package entity

import "time"

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	EventTime time.Time `json:"event_time"`
	Email     string    `json:"email"`
}
