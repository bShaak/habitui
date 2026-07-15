package storage

import (
	"context"
	"time"

	"github.com/bShaak/habitui/internal/models"
)

type Store interface {
	CreateHabit(ctx context.Context, h *models.Habit) (*models.Habit, error)
	UpdateHabit(ctx context.Context, h *models.Habit) error
	DeleteHabit(ctx context.Context, id int64) error
	ListHabits(ctx context.Context) ([]models.Habit, error)

	CreateCompletion(ctx context.Context, c *models.Completion) (*models.Completion, error)
	DeleteCompletion(ctx context.Context, id int64) error
	ListCompletions(ctx context.Context) ([]models.Completion, error)
	GetCompletionsByHabitID(ctx context.Context, habitID int64) ([]models.Completion, error)
	GetCompletionsByHabitIDAndDate(ctx context.Context, habitID int64, date time.Time) ([]models.Completion, error)
	GetCompletionsByDate(ctx context.Context, date time.Time) ([]models.Completion, error)
	GetCompletionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Completion, error)

	Close() error
}
