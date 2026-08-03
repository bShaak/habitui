package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bShaak/habitui/internal/habits"
	"github.com/bShaak/habitui/internal/models"
)

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var (
		jsonOut bool
		dbPath  string
		dateStr string
	)
	fs.BoolVar(&jsonOut, "json", false, "emit JSON")
	fs.StringVar(&dbPath, "db", "", "SQLite database path (default: ~/.habitui/habit.db or HABITUI_DB)")
	fs.StringVar(&dateStr, "date", "", "calendar day to evaluate (YYYY-MM-DD, default: today)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	date, err := parseDate(dateStr)
	if err != nil {
		return err
	}

	store, err := openStore(resolveDBPath(dbPath))
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	habitList, err := store.ListHabits(ctx)
	if err != nil {
		return err
	}
	completions, err := store.GetCompletionsByDate(ctx, date)
	if err != nil {
		return err
	}

	resp := listResponse{
		Date:   date.Format("2006-01-02"),
		Habits: make([]habitSummary, 0, len(habitList)),
	}
	for _, h := range habitList {
		count := habits.CompletionCountForDate(completions, h.ID, date)
		resp.Habits = append(resp.Habits, habitSummary{
			ID:              h.ID,
			Name:            h.Name,
			Description:     h.Description,
			Frequency:       h.Frequency,
			Goal:            habits.EffectiveGoal(h.Goal),
			Color:           h.Color,
			Icon:            h.Icon,
			Due:             habits.IsDueOnDate(h, date),
			Complete:        count >= habits.EffectiveGoal(h.Goal),
			CompletionCount: count,
		})
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	}

	for _, h := range resp.Habits {
		status := "—"
		if h.Due {
			if h.Complete {
				status = "complete"
			} else {
				status = fmt.Sprintf("incomplete (%d/%d)", h.CompletionCount, h.Goal)
			}
		}
		fmt.Printf("%d\t%s\t%s\n", h.ID, h.Name, status)
	}
	return nil
}

func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --date %q (use YYYY-MM-DD): %w", dateStr, err)
	}
	return t, nil
}

func findHabit(habitList []models.Habit, id int64) (*models.Habit, error) {
	for i := range habitList {
		if habitList[i].ID == id {
			return &habitList[i], nil
		}
	}
	return nil, fmt.Errorf("habit %d not found", id)
}
