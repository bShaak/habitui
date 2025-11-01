package models
type Habit struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Frequency   string    `json:"frequency"` // e.g. "daily" or "weekly"; default daily
	Goal        int       `json:"goal"` // e.g. 100; default 1
	StartDate   string 	  `json:"start_date"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	ArchivedAt  string    `json:"archived_at"`
}

type Completion struct {
	ID          int64     `json:"id"`
	HabitID     int64     `json:"habit_id"`
	CompletedAt string    `json:"completed_at"`
}
