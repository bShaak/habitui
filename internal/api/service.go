package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bShaak/habitui/internal/habits"
	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
)

type service struct {
	store storage.Store
}

func (s *service) ListHabits(ctx context.Context) ([]models.Habit, error) {
	return s.store.ListHabits(ctx)
}

func (s *service) GetHabit(ctx context.Context, id int64) (*models.Habit, error) {
	habitList, err := s.store.ListHabits(ctx)
	if err != nil {
		return nil, err
	}
	for i := range habitList {
		if habitList[i].ID == id {
			return &habitList[i], nil
		}
	}
	return nil, fmt.Errorf("habit %d not found", id)
}

func (s *service) CreateHabit(ctx context.Context, h *models.Habit) (*models.Habit, error) {
	return s.store.CreateHabit(ctx, h)
}

func (s *service) UpdateHabit(ctx context.Context, h *models.Habit) error {
	return s.store.UpdateHabit(ctx, h)
}

func (s *service) DeleteHabit(ctx context.Context, id int64) error {
	return s.store.DeleteHabit(ctx, id)
}

func (s *service) DaySummary(ctx context.Context, date time.Time) (DaySummary, error) {
	habitList, err := s.store.ListHabits(ctx)
	if err != nil {
		return DaySummary{}, err
	}
	completions, err := s.store.GetCompletionsByDate(ctx, date)
	if err != nil {
		return DaySummary{}, err
	}

	resp := DaySummary{
		Date:   date.Format("2006-01-02"),
		Habits: make([]HabitSummary, 0, len(habitList)),
	}
	for _, h := range habitList {
		goal := habits.EffectiveGoal(h.Goal)
		count := habits.CompletionCountForDate(completions, h.ID, date)
		resp.Habits = append(resp.Habits, HabitSummary{
			ID:              h.ID,
			Name:            h.Name,
			Description:     h.Description,
			Frequency:       h.Frequency,
			Goal:            goal,
			Color:           h.Color,
			Icon:            h.Icon,
			Due:             habits.IsDueOnDate(h, date),
			Complete:        count >= goal,
			CompletionCount: count,
		})
	}
	return resp, nil
}

func (s *service) CompletionsByDate(ctx context.Context, date time.Time) ([]models.Completion, error) {
	return s.store.GetCompletionsByDate(ctx, date)
}

func (s *service) CompletionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]models.Completion, error) {
	return s.store.GetCompletionsByDateRange(ctx, startDate, endDate)
}

func (s *service) ListCompletions(ctx context.Context) ([]models.Completion, error) {
	return s.store.ListCompletions(ctx)
}

// CompleteHabit records one completion for habitID on date unless the daily
// goal is already met.
func (s *service) CompleteHabit(ctx context.Context, habitID int64, date time.Time) (CompleteResult, error) {
	habit, err := s.GetHabit(ctx, habitID)
	if err != nil {
		return CompleteResult{}, err
	}

	goal := habits.EffectiveGoal(habit.Goal)
	existing, err := s.store.GetCompletionsByHabitIDAndDate(ctx, habitID, date)
	if err != nil {
		return CompleteResult{}, err
	}
	count := habits.CompletionCountForDate(existing, habitID, date)

	resp := CompleteResult{
		HabitID:         habitID,
		Date:            date.Format("2006-01-02"),
		CompletionCount: count,
		Goal:            goal,
	}
	if count >= goal {
		resp.AlreadyComplete = true
		return resp, nil
	}

	c, err := s.store.CreateCompletion(ctx, &models.Completion{
		HabitID:     habitID,
		CompletedAt: timestampOnDay(date),
	})
	if err != nil {
		return CompleteResult{}, err
	}
	resp.Completion = c
	resp.CompletionCount = count + 1
	return resp, nil
}

// ToggleCompletionForDate adds a completion for the day if the goal is not yet
// met; otherwise it removes all of that day's completions for the habit.
func (s *service) ToggleCompletionForDate(ctx context.Context, habitID int64, date time.Time) (ToggleResult, error) {
	habit, err := s.GetHabit(ctx, habitID)
	if err != nil {
		return ToggleResult{}, err
	}

	goal := habits.EffectiveGoal(habit.Goal)
	existing, err := s.store.GetCompletionsByHabitIDAndDate(ctx, habitID, date)
	if err != nil {
		return ToggleResult{}, err
	}
	count := habits.CompletionCountForDate(existing, habitID, date)

	resp := ToggleResult{
		HabitID:         habitID,
		Date:            date.Format("2006-01-02"),
		CompletionCount: count,
		Goal:            goal,
	}

	if count < goal {
		c, err := s.store.CreateCompletion(ctx, &models.Completion{
			HabitID:     habitID,
			CompletedAt: timestampOnDay(date),
		})
		if err != nil {
			return ToggleResult{}, err
		}
		resp.Added = c
		resp.CompletionCount = count + 1
	} else {
		all, err := s.store.GetCompletionsByHabitID(ctx, habitID)
		if err != nil {
			return ToggleResult{}, err
		}
		var removed []int64
		for _, c := range all {
			completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
			if err != nil || !habits.InDayBounds(completedAt, date) {
				continue
			}
			if err := s.store.DeleteCompletion(ctx, c.ID); err != nil {
				return ToggleResult{}, err
			}
			removed = append(removed, c.ID)
		}
		resp.RemovedIDs = removed
		resp.CompletionCount = 0
	}

	resp.Complete = resp.CompletionCount >= goal
	return resp, nil
}

func (s *service) Close() error { return s.store.Close() }

// timestampOnDay returns an RFC3339 timestamp combining date's calendar day
// with the current wall-clock time.
func timestampOnDay(date time.Time) string {
	now := time.Now()
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		now.Hour(), now.Minute(), now.Second(), 0, now.Location(),
	).Format(time.RFC3339)
}
