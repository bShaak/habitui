package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/cli"
	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
)

func openTestStore(t *testing.T) (*storage.SQLiteStore, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "habit.db")
	store, err := storage.OpenSQLiteAt(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, dbPath
}

func seedHabit(t *testing.T, store *storage.SQLiteStore, name, frequency string, goal int) *models.Habit {
	t.Helper()
	habit, err := store.CreateHabit(context.Background(), &models.Habit{
		Name:      name,
		Frequency: frequency,
		Goal:      goal,
		StartDate: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create habit: %v", err)
	}
	return habit
}

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })
	done := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.Bytes()
	}()
	fn()
	_ = w.Close()
	return <-done
}

func TestListJSON(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 12, 0, 0, 0, loc) // Monday
	daily := seedHabit(t, store, "Run", "daily", 1)
	weekdayOnly := seedHabit(t, store, "Code", "tuesday,thursday", 1)

	if _, err := store.CreateCompletion(context.Background(), &models.Completion{
		HabitID:     daily.ID,
		CompletedAt: time.Date(2026, 8, 3, 9, 0, 0, 0, loc).Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("create completion: %v", err)
	}

	out := captureStdout(t, func() {
		if err := cli.Run([]string{
			"list", "--json", "--db", dbPath, "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("list: %v", err)
		}
	})

	var resp struct {
		Date   string `json:"date"`
		Habits []struct {
			ID              int64  `json:"id"`
			Name            string `json:"name"`
			Due             bool   `json:"due"`
			Complete        bool   `json:"complete"`
			CompletionCount int    `json:"completion_count"`
			Goal            int    `json:"goal"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if resp.Date != "2026-08-03" {
		t.Fatalf("date = %q", resp.Date)
	}
	if len(resp.Habits) != 2 {
		t.Fatalf("expected 2 habits, got %d", len(resp.Habits))
	}

	byID := map[int64]struct {
		Due             bool
		Complete        bool
		CompletionCount int
	}{}
	for _, h := range resp.Habits {
		byID[h.ID] = struct {
			Due             bool
			Complete        bool
			CompletionCount int
		}{h.Due, h.Complete, h.CompletionCount}
	}

	run := byID[daily.ID]
	if !run.Due || !run.Complete || run.CompletionCount != 1 {
		t.Fatalf("daily habit: due=%v complete=%v count=%d", run.Due, run.Complete, run.CompletionCount)
	}

	code := byID[weekdayOnly.ID]
	if code.Due || code.Complete || code.CompletionCount != 0 {
		t.Fatalf("weekday habit on monday: due=%v complete=%v count=%d", code.Due, code.Complete, code.CompletionCount)
	}
}

func TestCompleteJSON(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	habit := seedHabit(t, store, "Meditate", "daily", 2)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete: %v", err)
		}
	})

	var resp struct {
		HabitID         int64 `json:"habit_id"`
		Date            string
		AlreadyComplete bool `json:"already_complete"`
		Completion      *struct {
			ID int64 `json:"id"`
		}
		CompletionCount int `json:"completion_count"`
		Goal            int `json:"goal"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if resp.HabitID != habit.ID || resp.Date != "2026-08-03" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.AlreadyComplete || resp.Completion == nil || resp.Completion.ID == 0 {
		t.Fatalf("expected new completion, got %+v", resp)
	}
	if resp.CompletionCount != 1 || resp.Goal != 2 {
		t.Fatalf("count/goal = %d/%d", resp.CompletionCount, resp.Goal)
	}

	completions, err := store.GetCompletionsByHabitIDAndDate(context.Background(), habit.ID, day)
	if err != nil {
		t.Fatalf("get completions: %v", err)
	}
	if len(completions) != 1 {
		t.Fatalf("expected 1 completion in store, got %d", len(completions))
	}

	out = captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete again: %v", err)
		}
	})
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if resp.CompletionCount != 2 || !resp.AlreadyComplete {
		t.Fatalf("expected complete after second call, got %+v", resp)
	}
}

func TestCompleteRequiresID(t *testing.T) {
	_, dbPath := openTestStore(t)
	err := cli.Run([]string{"complete", "--db", dbPath})
	if err == nil {
		t.Fatal("expected error without --id")
	}
}

func TestListUsesHABITUI_DB(t *testing.T) {
	store, dbPath := openTestStore(t)
	seedHabit(t, store, "Walk", "daily", 1)
	t.Setenv("HABITUI_DB", dbPath)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{"list", "--json"}); err != nil {
			t.Fatalf("list: %v", err)
		}
	})

	var resp struct {
		Habits []struct {
			Name string `json:"name"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if len(resp.Habits) != 1 || resp.Habits[0].Name != "Walk" {
		t.Fatalf("unexpected habits: %+v", resp.Habits)
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
