package cli_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bShaak/habitui/internal/cli"
	"github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
	_ "modernc.org/sqlite"
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
	if resp.CompletionCount != 2 || resp.AlreadyComplete || resp.Completion == nil {
		t.Fatalf("expected write that met goal, got %+v", resp)
	}
}

func TestCompleteAlreadyCompleteNoop(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	habit := seedHabit(t, store, "Stretch", "daily", 1)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete: %v", err)
		}
	})

	var first struct {
		AlreadyComplete bool `json:"already_complete"`
		Completion      *struct {
			ID int64 `json:"id"`
		}
		CompletionCount int `json:"completion_count"`
	}
	if err := json.Unmarshal(out, &first); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if first.AlreadyComplete || first.Completion == nil || first.CompletionCount != 1 {
		t.Fatalf("expected first write to goal, got %+v", first)
	}

	out = captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete again: %v", err)
		}
	})

	var resp struct {
		HabitID         int64 `json:"habit_id"`
		AlreadyComplete bool  `json:"already_complete"`
		Completion      *struct {
			ID int64 `json:"id"`
		}
		CompletionCount int `json:"completion_count"`
		Goal            int `json:"goal"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if !resp.AlreadyComplete {
		t.Fatalf("expected already_complete=true, got %+v", resp)
	}
	if resp.Completion != nil {
		t.Fatalf("expected completion omitted/null, got %+v", resp.Completion)
	}
	if resp.CompletionCount != 1 || resp.Goal != 1 {
		t.Fatalf("count/goal = %d/%d, want 1/1", resp.CompletionCount, resp.Goal)
	}

	completions, err := store.GetCompletionsByHabitIDAndDate(context.Background(), habit.ID, day)
	if err != nil {
		t.Fatalf("get completions: %v", err)
	}
	if len(completions) != 1 {
		t.Fatalf("expected 1 completion in store (noop), got %d", len(completions))
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

func TestCompleteUnknownID(t *testing.T) {
	store, dbPath := openTestStore(t)
	err := cli.Run([]string{"complete", "--json", "--db", dbPath, "--id", "99999", "--date", "2026-08-03"})
	if err == nil {
		t.Fatal("expected error for unknown habit id")
	}
	if !strings.Contains(err.Error(), "99999") {
		t.Fatalf("error should mention id, got %v", err)
	}
	completions, listErr := store.ListCompletions(context.Background())
	if listErr != nil {
		t.Fatalf("list completions: %v", listErr)
	}
	if len(completions) != 0 {
		t.Fatalf("expected no writes, got %d completions", len(completions))
	}
}

func TestInvalidDate(t *testing.T) {
	_, dbPath := openTestStore(t)
	for _, args := range [][]string{
		{"list", "--json", "--db", dbPath, "--date", "not-a-date"},
		{"list", "--json", "--db", dbPath, "--date", "08/03/2026"},
		{"complete", "--json", "--db", dbPath, "--id", "1", "--date", "not-a-date"},
		{"complete", "--json", "--db", dbPath, "--id", "1", "--date", "08/03/2026"},
	} {
		err := cli.Run(args)
		if err == nil {
			t.Fatalf("expected invalid --date error for %v", args)
		}
		if !strings.Contains(err.Error(), "--date") {
			t.Fatalf("error should mention --date for %v, got %v", args, err)
		}
	}
}

func TestCompleteStampsDate(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 7, 15, 0, 0, 0, 0, loc)
	habit := seedHabit(t, store, "Journal", "daily", 1)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", "2026-07-15",
		}); err != nil {
			t.Fatalf("complete: %v", err)
		}
	})

	var resp struct {
		Date       string `json:"date"`
		Completion *struct {
			CompletedAt string
		} `json:"completion"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if resp.Date != "2026-07-15" {
		t.Fatalf("date = %q", resp.Date)
	}
	if resp.Completion == nil {
		t.Fatal("expected completion")
	}
	completedAt, err := time.Parse(time.RFC3339, resp.Completion.CompletedAt)
	if err != nil {
		t.Fatalf("parse completed_at: %v", err)
	}
	y, m, d := completedAt.In(loc).Date()
	if y != 2026 || m != time.July || d != 15 {
		t.Fatalf("completed_at calendar day = %04d-%02d-%02d, want 2026-07-15", y, m, d)
	}

	got, err := store.GetCompletionsByHabitIDAndDate(context.Background(), habit.ID, day)
	if err != nil {
		t.Fatalf("get completions: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 completion on 2026-07-15, got %d", len(got))
	}
}

func TestDBFlagWinsOverEnv(t *testing.T) {
	envStore, envPath := openTestStore(t)
	flagStore, flagPath := openTestStore(t)
	seedHabit(t, envStore, "FromEnv", "daily", 1)
	seedHabit(t, flagStore, "FromFlag", "daily", 1)
	t.Setenv("HABITUI_DB", envPath)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{"list", "--json", "--db", flagPath}); err != nil {
			t.Fatalf("list: %v", err)
		}
	})

	var resp struct {
		Habits []struct {
			Name string `json:"name"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if len(resp.Habits) != 1 || resp.Habits[0].Name != "FromFlag" {
		t.Fatalf("expected --db to win, got %+v", resp.Habits)
	}
}

func TestListPartialProgress(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	habit := seedHabit(t, store, "Pushups", "daily", 3)

	if _, err := store.CreateCompletion(context.Background(), &models.Completion{
		HabitID:     habit.ID,
		CompletedAt: time.Date(2026, 8, 3, 8, 0, 0, 0, loc).Format(time.RFC3339),
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
		Habits []struct {
			ID              int64 `json:"id"`
			Due             bool  `json:"due"`
			Complete        bool  `json:"complete"`
			CompletionCount int   `json:"completion_count"`
			Goal            int   `json:"goal"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("decode json: %v\n%s", err, out)
	}
	if len(resp.Habits) != 1 {
		t.Fatalf("expected 1 habit, got %d", len(resp.Habits))
	}
	h := resp.Habits[0]
	if !h.Due || h.Complete || h.CompletionCount != 1 || h.Goal != 3 {
		t.Fatalf("partial progress: due=%v complete=%v count=%d goal=%d", h.Due, h.Complete, h.CompletionCount, h.Goal)
	}
}

func TestRunMissingAndUnknownCommand(t *testing.T) {
	if err := cli.Run(nil); err == nil {
		t.Fatal("expected error for missing command")
	}
	if err := cli.Run([]string{}); err == nil {
		t.Fatal("expected error for empty args")
	}
	err := cli.Run([]string{"nope"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("error should mention command, got %v", err)
	}
}

func TestGoalZeroTreatedAsOne(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 0, 0, 0, 0, loc)
	habit := seedHabit(t, store, "Water", "daily", 1)
	forceHabitGoal(t, dbPath, habit.ID, 0)

	out := captureStdout(t, func() {
		if err := cli.Run([]string{
			"list", "--json", "--db", dbPath, "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("list: %v", err)
		}
	})
	var listResp struct {
		Habits []struct {
			Goal     int  `json:"goal"`
			Complete bool `json:"complete"`
		} `json:"habits"`
	}
	if err := json.Unmarshal(out, &listResp); err != nil {
		t.Fatalf("decode list json: %v\n%s", err, out)
	}
	if len(listResp.Habits) != 1 || listResp.Habits[0].Goal != 1 {
		t.Fatalf("list goal = %+v, want goal 1", listResp.Habits)
	}

	out = captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--json", "--db", dbPath, "--id", itoa(habit.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete: %v", err)
		}
	})
	var completeResp struct {
		Goal            int  `json:"goal"`
		CompletionCount int  `json:"completion_count"`
		AlreadyComplete bool `json:"already_complete"`
	}
	if err := json.Unmarshal(out, &completeResp); err != nil {
		t.Fatalf("decode complete json: %v\n%s", err, out)
	}
	if completeResp.Goal != 1 || completeResp.CompletionCount != 1 || completeResp.AlreadyComplete {
		t.Fatalf("complete with goal 0: %+v", completeResp)
	}
}

func TestListAndCompleteTextOutput(t *testing.T) {
	store, dbPath := openTestStore(t)
	loc := time.Local
	day := time.Date(2026, 8, 3, 0, 0, 0, 0, loc) // Monday
	due := seedHabit(t, store, "Run", "daily", 2)
	notDue := seedHabit(t, store, "Code", "tuesday", 1)

	if _, err := store.CreateCompletion(context.Background(), &models.Completion{
		HabitID:     due.ID,
		CompletedAt: time.Date(2026, 8, 3, 9, 0, 0, 0, loc).Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("create completion: %v", err)
	}

	listOut := string(captureStdout(t, func() {
		if err := cli.Run([]string{
			"list", "--db", dbPath, "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("list: %v", err)
		}
	}))
	if !strings.Contains(listOut, "incomplete (1/2)") {
		t.Fatalf("list text missing partial status:\n%s", listOut)
	}
	if !strings.Contains(listOut, itoa(notDue.ID)+"\tCode\t—") {
		t.Fatalf("list text missing not-due row:\n%s", listOut)
	}

	completeOut := string(captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--db", dbPath, "--id", itoa(due.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete: %v", err)
		}
	}))
	if !strings.Contains(completeOut, "completed") || !strings.Contains(completeOut, "2/2") {
		t.Fatalf("complete text unexpected:\n%s", completeOut)
	}

	noopOut := string(captureStdout(t, func() {
		if err := cli.Run([]string{
			"complete", "--db", dbPath, "--id", itoa(due.ID), "--date", day.Format("2006-01-02"),
		}); err != nil {
			t.Fatalf("complete noop: %v", err)
		}
	}))
	if !strings.Contains(noopOut, "already complete") {
		t.Fatalf("noop text unexpected:\n%s", noopOut)
	}
}

func forceHabitGoal(t *testing.T, dbPath string, habitID int64, goal int) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(`UPDATE habits SET goal = ? WHERE id = ?`, goal, habitID); err != nil {
		t.Fatalf("force goal: %v", err)
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
