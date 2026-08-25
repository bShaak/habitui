// Package api is the application facade over the habits business logic and
// storage layer. All clients (CLI, TUI, HTTP server) should depend on this
// package rather than on internal/habits or internal/storage directly.
package api

import (
	"context"
	"os"
	"time"

	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
)

// HabitSummary is a habit enriched with due/complete state for a specific day.
type HabitSummary struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Frequency       string `json:"frequency"`
	Goal            int    `json:"goal"`
	Color           string `json:"color,omitempty"`
	Icon            string `json:"icon,omitempty"`
	Due             bool   `json:"due"`
	Complete        bool   `json:"complete"`
	CompletionCount int    `json:"completion_count"`
}

// DaySummary describes all habits evaluated against a single calendar day.
type DaySummary struct {
	Date   string         `json:"date"`
	Habits []HabitSummary `json:"habits"`
}

// CompleteResult reports the outcome of recording a completion for a habit.
type CompleteResult struct {
	HabitID         int64              `json:"habit_id"`
	Date            string             `json:"date"`
	AlreadyComplete bool               `json:"already_complete"`
	Completion      *models.Completion `json:"completion,omitempty"`
	CompletionCount int                `json:"completion_count"`
	Goal            int                `json:"goal"`
}

// ToggleResult reports the outcome of toggling a habit's completions for a
// day: if the goal was already met, all of that day's completions are removed;
// otherwise one completion is added.
type ToggleResult struct {
	HabitID         int64              `json:"habit_id"`
	Date            string             `json:"date"`
	Complete        bool               `json:"complete"`
	CompletionCount int                `json:"completion_count"`
	Goal            int                `json:"goal"`
	Added           *models.Completion `json:"added,omitempty"`
	RemovedIDs      []int64            `json:"removed_ids,omitempty"`
}

// Service is the public API for controlling habits. It hides both the
// business-logic package and the persistence layer behind these methods.
type Service interface {
	ListHabits(ctx context.Context) ([]models.Habit, error)
	GetHabit(ctx context.Context, id int64) (*models.Habit, error)
	CreateHabit(ctx context.Context, h *models.Habit) (*models.Habit, error)
	UpdateHabit(ctx context.Context, h *models.Habit) error
	DeleteHabit(ctx context.Context, id int64) error

	DaySummary(ctx context.Context, date time.Time) (DaySummary, error)
	CompletionsByDate(ctx context.Context, date time.Time) ([]models.Completion, error)
	CompletionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Completion, error)
	ListCompletions(ctx context.Context) ([]models.Completion, error)

	CompleteHabit(ctx context.Context, habitID int64, date time.Time) (CompleteResult, error)
	ToggleCompletionForDate(ctx context.Context, habitID int64, date time.Time) (ToggleResult, error)

	Close() error
}

// New returns a Service backed by the given store.
func New(store storage.Store) Service {
	return &service{store: store}
}

// Open opens the SQLite database at dbPath and returns a Service over it.
// If dbPath is empty, the HABITUI_DB environment variable is used; falling
// back to ~/.habitui/habit.db when unset.
func Open(dbPath string) (Service, error) {
	if dbPath == "" {
		dbPath = os.Getenv("HABITUI_DB")
	}
	var (
		store *storage.SQLiteStore
		err   error
	)
	if dbPath == "" {
		store, err = storage.OpenSQLite()
	} else {
		store, err = storage.OpenSQLiteAt(dbPath)
	}
	if err != nil {
		return nil, err
	}
	return New(store), nil
}
