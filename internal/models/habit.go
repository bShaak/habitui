package models

import "time"

type Habit struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Frequency   string    `json:"frequency"` // e.g. "daily" or "weekly"; default daily
	StartDate   time.Time `json:"start_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ArchivedAt  time.Time `json:"archived_at"`
}

type HabitLog struct {
	ID        int64     `json:"id"`
	HabitID   int64     `json:"habit_id"`
	Timestamp time.Time `json:"timestamp"`
}
