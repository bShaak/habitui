package models

type Habit struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Frequency   string `json:"frequency"` // e.g. "daily" or "monday,wednesday"; default daily
	Goal        int    `json:"goal"`      // times per day; default 1
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	StartDate   string `json:"start_date"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Completion struct {
	ID          int64  `json:"id"`
	HabitID     int64  `json:"habit_id"`
	CompletedAt string `json:"completed_at"`
}
