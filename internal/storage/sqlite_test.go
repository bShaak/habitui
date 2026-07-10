package storage_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
)

func openTestStore(t *testing.T) *storage.SQLiteStore {
	t.Helper()
	store, err := storage.OpenSQLiteAt(filepath.Join(t.TempDir(), "habit.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestCreateCompletionSetsID(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	habit, err := store.CreateHabit(ctx, &models.Habit{
		Name:      "Run",
		StartDate: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	c, err := store.CreateCompletion(ctx, &models.Completion{
		HabitID:     habit.ID,
		CompletedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create completion: %v", err)
	}
	if c.ID == 0 {
		t.Fatal("expected completion ID to be set")
	}
}

func TestDeleteHabitRemovesCompletions(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	habit, err := store.CreateHabit(ctx, &models.Habit{
		Name:      "Read",
		StartDate: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	if _, err := store.CreateCompletion(ctx, &models.Completion{
		HabitID:     habit.ID,
		CompletedAt: time.Now().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("create completion: %v", err)
	}

	if err := store.DeleteHabit(ctx, habit.ID); err != nil {
		t.Fatalf("delete habit: %v", err)
	}

	completions, err := store.GetCompletionsByHabitId(ctx, habit.ID)
	if err != nil {
		t.Fatalf("list completions: %v", err)
	}
	if len(completions) != 0 {
		t.Fatalf("expected 0 completions after delete, got %d", len(completions))
	}
}

func TestGetCompletionsByDateHandlesLocalAndUTC(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()
	loc := time.Local

	habit, err := store.CreateHabit(ctx, &models.Habit{
		Name:      "Water",
		StartDate: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}

	day := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	localTS := time.Date(2026, 7, 10, 15, 30, 0, 0, loc).Format(time.RFC3339)
	utcTS := time.Date(2026, 7, 10, 15, 30, 0, 0, loc).UTC().Format(time.RFC3339)

	if _, err := store.CreateCompletion(ctx, &models.Completion{HabitID: habit.ID, CompletedAt: localTS}); err != nil {
		t.Fatalf("create local completion: %v", err)
	}
	if _, err := store.CreateCompletion(ctx, &models.Completion{HabitID: habit.ID, CompletedAt: utcTS}); err != nil {
		t.Fatalf("create utc completion: %v", err)
	}

	got, err := store.GetCompletionsByDate(ctx, day)
	if err != nil {
		t.Fatalf("get by date: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 completions for local day, got %d", len(got))
	}

	byHabitDay, err := store.GetCompletionsByHabitIdAndDate(ctx, habit.ID, day)
	if err != nil {
		t.Fatalf("get by habit+date: %v", err)
	}
	if len(byHabitDay) != 2 {
		t.Fatalf("expected 2 completions for habit+date, got %d", len(byHabitDay))
	}
}

func TestCreateHabitDefaults(t *testing.T) {
	store := openTestStore(t)
	habit, err := store.CreateHabit(context.Background(), &models.Habit{
		Name:      "Defaulted",
		StartDate: time.Now().Format(time.RFC3339),
		Goal:      0,
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	if habit.Goal != 1 {
		t.Fatalf("expected goal 1, got %d", habit.Goal)
	}
	if habit.Frequency != "daily" {
		t.Fatalf("expected frequency daily, got %q", habit.Frequency)
	}
}
