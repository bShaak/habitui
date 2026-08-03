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

func runComplete(args []string) error {
	fs := flag.NewFlagSet("complete", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var (
		jsonOut bool
		dbPath  string
		dateStr string
		habitID int64
	)
	fs.BoolVar(&jsonOut, "json", false, "emit JSON")
	fs.StringVar(&dbPath, "db", "", "SQLite database path (default: ~/.habitui/habit.db or HABITUI_DB)")
	fs.StringVar(&dateStr, "date", "", "calendar day to complete (YYYY-MM-DD, default: today)")
	fs.Int64Var(&habitID, "id", 0, "habit ID (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if habitID == 0 {
		return fmt.Errorf("--id is required")
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
	habit, err := findHabit(habitList, habitID)
	if err != nil {
		return err
	}

	goal := habits.EffectiveGoal(habit.Goal)
	existing, err := store.GetCompletionsByHabitIDAndDate(ctx, habitID, date)
	if err != nil {
		return err
	}
	count := habits.CompletionCountForDate(existing, habitID, date)

	resp := completeResponse{
		HabitID:         habitID,
		Date:            date.Format("2006-01-02"),
		CompletionCount: count,
		Goal:            goal,
	}

	if count >= goal {
		resp.AlreadyComplete = true
		return writeCompleteOutput(jsonOut, resp)
	}

	now := time.Now()
	completedAt := time.Date(
		date.Year(), date.Month(), date.Day(),
		now.Hour(), now.Minute(), now.Second(), 0, now.Location(),
	)
	c, err := store.CreateCompletion(ctx, &models.Completion{
		HabitID:     habitID,
		CompletedAt: completedAt.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}
	resp.Completion = c
	resp.CompletionCount = count + 1
	resp.AlreadyComplete = resp.CompletionCount >= goal

	return writeCompleteOutput(jsonOut, resp)
}

func writeCompleteOutput(jsonOut bool, resp completeResponse) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp)
	}
	if resp.AlreadyComplete && resp.Completion == nil {
		fmt.Printf("habit %d already complete on %s (%d/%d)\n", resp.HabitID, resp.Date, resp.CompletionCount, resp.Goal)
		return nil
	}
	fmt.Printf("habit %d completed on %s (%d/%d)\n", resp.HabitID, resp.Date, resp.CompletionCount, resp.Goal)
	return nil
}
