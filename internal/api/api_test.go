package api

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
)

func newTestService(t *testing.T) Service {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	svc, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q): %v", dbPath, err)
	}
	t.Cleanup(func() { _ = svc.Close() })
	return svc
}

func dayOf(y int, m time.Month, d int) time.Time {
	now := time.Now()
	return time.Date(y, m, d, now.Hour(), now.Minute(), now.Second(), 0, time.Local)
}

func today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
}

func createHabit(t *testing.T, svc Service, name string, goal int) *models.Habit {
	t.Helper()
	h, err := svc.CreateHabit(context.Background(), &models.Habit{Name: name, Goal: goal})
	if err != nil {
		t.Fatalf("CreateHabit(%q): %v", name, err)
	}
	return h
}

func TestDaySummaryEmptyAndPopulated(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	sum, err := svc.DaySummary(ctx, today())
	if err != nil {
		t.Fatalf("DaySummary: %v", err)
	}
	if len(sum.Habits) != 0 {
		t.Fatalf("expected no habits, got %d", len(sum.Habits))
	}

	createHabit(t, svc, "Read", 2)
	createHabit(t, svc, "Run", 1)

	sum, err = svc.DaySummary(ctx, today())
	if err != nil {
		t.Fatalf("DaySummary: %v", err)
	}
	if sum.Date != today().Format("2006-01-02") {
		t.Errorf("Date = %q, want %q", sum.Date, today().Format("2006-01-02"))
	}
	if len(sum.Habits) != 2 {
		t.Fatalf("expected 2 habits, got %d", len(sum.Habits))
	}
	for _, h := range sum.Habits {
		if !h.Due {
			t.Errorf("%s: expected due on its scheduled day", h.Name)
		}
		if h.Complete || h.CompletionCount != 0 {
			t.Errorf("%s: expected incomplete with 0 completions, got complete=%v count=%d", h.Name, h.Complete, h.CompletionCount)
		}
	}
}

func TestCompleteHabitRespectsGoal(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	h := createHabit(t, svc, "Meditate", 2)

	res1, err := svc.CompleteHabit(ctx, h.ID, today())
	if err != nil {
		t.Fatalf("CompleteHabit #1: %v", err)
	}
	if res1.AlreadyComplete || res1.CompletionCount != 1 || res1.Completion == nil {
		t.Errorf("res1 = %+v; want first completion recorded", res1)
	}

	res2, err := svc.CompleteHabit(ctx, h.ID, today())
	if err != nil {
		t.Fatalf("CompleteHabit #2: %v", err)
	}
	if res2.AlreadyComplete || res2.CompletionCount != 2 || res2.Completion == nil {
		t.Errorf("res2 = %+v; want second completion recorded", res2)
	}

	res3, err := svc.CompleteHabit(ctx, h.ID, today())
	if err != nil {
		t.Fatalf("CompleteHabit #3: %v", err)
	}
	if !res3.AlreadyComplete || res3.Completion != nil || res3.CompletionCount != 2 {
		t.Errorf("res3 = %+v; want already-complete no-op", res3)
	}

	sum, err := svc.DaySummary(ctx, today())
	if err != nil {
		t.Fatalf("DaySummary: %v", err)
	}
	if len(sum.Habits) != 1 || !sum.Habits[0].Complete {
		t.Fatalf("summary after completing = %+v; want complete habit", sum.Habits)
	}
}

func TestToggleCompletionForDateAddsThenRemoves(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	h := createHabit(t, svc, "Water plants", 1)

	on, err := svc.ToggleCompletionForDate(ctx, h.ID, today())
	if err != nil {
		t.Fatalf("toggle on: %v", err)
	}
	if on.Added == nil || len(on.RemovedIDs) != 0 || !on.Complete || on.CompletionCount != 1 {
		t.Errorf("on = %+v; want one added completion", on)
	}

	comps, err := svc.CompletionsByDate(ctx, today())
	if err != nil {
		t.Fatalf("CompletionsByDate: %v", err)
	}
	if len(comps) != 1 {
		t.Fatalf("expected 1 completion stored, got %d", len(comps))
	}

	off, err := svc.ToggleCompletionForDate(ctx, h.ID, today())
	if err != nil {
		t.Fatalf("toggle off: %v", err)
	}
	if off.Added != nil || len(off.RemovedIDs) != 1 || off.Complete || off.CompletionCount != 0 {
		t.Errorf("off = %+v; want removal of one completion", off)
	}

	comps, err = svc.CompletionsByDate(ctx, today())
	if err != nil {
		t.Fatalf("CompletionsByDate: %v", err)
	}
	if len(comps) != 0 {
		t.Fatalf("expected 0 completions stored, got %d", len(comps))
	}
}

func TestGetHabitNotFound(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.GetHabit(context.Background(), 999); err == nil {
		t.Fatal("expected error for missing habit")
	}
}

func TestUpdateAndDeleteHabit(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	h := createHabit(t, svc, "Journal", 1)

	h.Name = "Journal nightly"
	h.Goal = 3
	if err := svc.UpdateHabit(ctx, h); err != nil {
		t.Fatalf("UpdateHabit: %v", err)
	}

	got, err := svc.GetHabit(ctx, h.ID)
	if err != nil {
		t.Fatalf("GetHabit: %v", err)
	}
	if got.Name != "Journal nightly" || got.Goal != 3 {
		t.Errorf("got = %+v; want updated fields", got)
	}

	if _, err := svc.CompleteHabit(ctx, h.ID, today()); err != nil {
		t.Fatalf("CompleteHabit: %v", err)
	}
	if err := svc.DeleteHabit(ctx, h.ID); err != nil {
		t.Fatalf("DeleteHabit: %v", err)
	}
	if _, err := svc.GetHabit(ctx, h.ID); err == nil {
		t.Error("expected habit gone after delete")
	}
	comps, err := svc.ListCompletions(ctx)
	if err != nil {
		t.Fatalf("ListCompletions: %v", err)
	}
	if len(comps) != 0 {
		t.Errorf("expected completions cascaded away, got %d", len(comps))
	}
}

func TestCompletionsByDateRange(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	h := createHabit(t, svc, "Stretch", 1)

	yesterday := today().AddDate(0, 0, -1)
	if _, err := svc.CompleteHabit(ctx, h.ID, yesterday); err != nil {
		t.Fatalf("CompleteHabit yesterday: %v", err)
	}
	if _, err := svc.CompleteHabit(ctx, h.ID, today()); err != nil {
		t.Fatalf("CompleteHabit today: %v", err)
	}

	comps, err := svc.CompletionsByDateRange(ctx, today().AddDate(0, 0, -6), today())
	if err != nil {
		t.Fatalf("CompletionsByDateRange: %v", err)
	}
	if len(comps) != 2 {
		t.Errorf("range including both days: got %d completions, want 2", len(comps))
	}

	comps, err = svc.CompletionsByDateRange(ctx, today().AddDate(0, 0, 1), today().AddDate(0, 0, 7))
	if err != nil {
		t.Fatalf("CompletionsByDateRange future: %v", err)
	}
	if len(comps) != 0 {
		t.Errorf("future range: got %d completions, want 0", len(comps))
	}
}

func TestCreateHabitDefaultsApplied(t *testing.T) {
	svc := newTestService(t)
	h, err := svc.CreateHabit(context.Background(), &models.Habit{Name: "Bare"})
	if err != nil {
		t.Fatalf("CreateHabit: %v", err)
	}
	if h.Frequency != "daily" || h.Goal != 1 || h.Color == "" || h.StartDate == "" {
		t.Errorf("defaults not applied: %+v", h)
	}
}

func TestOpenInvalidPathFailsCleanly(t *testing.T) {
	svc, err := Open(filepath.Join(t.TempDir(), "sub", "dir", "ok.db"))
	if err != nil {
		t.Fatalf("Open should create parent dirs: %v", err)
	}
	defer svc.Close()

	var _ storage.Store // keep storage import used
}
