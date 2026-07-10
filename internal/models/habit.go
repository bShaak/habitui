package models

type Habit struct {
	ID          int64
	Name        string
	Description string
	Frequency   string // e.g. "daily" or "monday,wednesday"; default daily
	Goal        int    // times per day; default 1
	Color       string // red, blue, green, yellow, orange, purple, pink
	Icon        string // optional emoji icon
	StartDate   string
	CreatedAt   string
	UpdatedAt   string
}

type Completion struct {
	ID          int64
	HabitID     int64
	CompletedAt string
}
