package habits_test

import (
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/habits"
	"github.com/bShaak/habitui/internal/models"
)

func TestIsScheduledOnDay(t *testing.T) {
	if !habits.IsScheduledOnDay("daily", "monday") {
		t.Fatal("daily should schedule monday")
	}
	if habits.IsScheduledOnDay("monday,wednesday", "tuesday") {
		t.Fatal("tuesday should not be scheduled")
	}
}

func TestCompletionCountForDate(t *testing.T) {
	loc := time.Local
	date := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	completions := []models.Completion{
		{HabitID: 1, CompletedAt: time.Date(2026, 8, 3, 9, 0, 0, 0, loc).Format(time.RFC3339)},
		{HabitID: 1, CompletedAt: time.Date(2026, 8, 2, 9, 0, 0, 0, loc).Format(time.RFC3339)},
		{HabitID: 2, CompletedAt: time.Date(2026, 8, 3, 10, 0, 0, 0, loc).Format(time.RFC3339)},
	}
	if got := habits.CompletionCountForDate(completions, 1, date); got != 1 {
		t.Fatalf("count = %d, want 1", got)
	}
}

func TestIsCompleteForDate(t *testing.T) {
	loc := time.Local
	date := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	h := models.Habit{ID: 1, Goal: 2}
	completions := []models.Completion{
		{HabitID: 1, CompletedAt: time.Date(2026, 8, 3, 9, 0, 0, 0, loc).Format(time.RFC3339)},
	}
	if habits.IsCompleteForDate(completions, h, date) {
		t.Fatal("expected incomplete with 1/2")
	}
	completions = append(completions, models.Completion{
		HabitID: 1, CompletedAt: time.Date(2026, 8, 3, 10, 0, 0, 0, loc).Format(time.RFC3339),
	})
	if !habits.IsCompleteForDate(completions, h, date) {
		t.Fatal("expected complete with 2/2")
	}
}
