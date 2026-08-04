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
	if !habits.IsScheduledOnDay("", "sunday") {
		t.Fatal("empty frequency should schedule all days")
	}
	if habits.IsScheduledOnDay("monday,wednesday", "tuesday") {
		t.Fatal("tuesday should not be scheduled")
	}
	if !habits.IsScheduledOnDay("monday, wednesday", "wednesday") {
		t.Fatal("whitespace after comma should still schedule wednesday")
	}
	if !habits.IsScheduledOnDay(" monday , friday ", "monday") {
		t.Fatal("padded day names should schedule monday")
	}
}

func TestEffectiveGoal(t *testing.T) {
	if habits.EffectiveGoal(0) != 1 {
		t.Fatal("goal 0 should become 1")
	}
	if habits.EffectiveGoal(-3) != 1 {
		t.Fatal("negative goal should become 1")
	}
	if habits.EffectiveGoal(3) != 3 {
		t.Fatal("positive goal should be unchanged")
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

func TestCompletionCountForDateUTCAndLocal(t *testing.T) {
	loc := time.Local
	date := time.Date(2026, 7, 10, 0, 0, 0, 0, loc)

	// Same local instant encoded as local offset and as UTC.
	localTS := time.Date(2026, 7, 10, 15, 30, 0, 0, loc).Format(time.RFC3339)
	utcTS := time.Date(2026, 7, 10, 15, 30, 0, 0, loc).UTC().Format(time.RFC3339)

	// Just before/after local midnight in UTC form — should not count for this local day.
	before := time.Date(2026, 7, 10, 0, 0, 0, 0, loc).Add(-time.Second).UTC().Format(time.RFC3339)
	after := time.Date(2026, 7, 10, 23, 59, 59, 999999999, loc).Add(time.Second).UTC().Format(time.RFC3339)

	completions := []models.Completion{
		{HabitID: 1, CompletedAt: localTS},
		{HabitID: 1, CompletedAt: utcTS},
		{HabitID: 1, CompletedAt: before},
		{HabitID: 1, CompletedAt: after},
	}
	if got := habits.CompletionCountForDate(completions, 1, date); got != 2 {
		t.Fatalf("count = %d, want 2 (local+utc same instant; exclude outside day)", got)
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
