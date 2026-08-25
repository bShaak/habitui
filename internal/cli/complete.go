package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/bShaak/habitui/internal/api"
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

	svc, err := openService(dbPath)
	if err != nil {
		return err
	}
	defer svc.Close()

	resp, err := svc.CompleteHabit(context.Background(), habitID, date)
	if err != nil {
		return err
	}
	return writeCompleteOutput(jsonOut, resp)
}

func writeCompleteOutput(jsonOut bool, resp api.CompleteResult) error {
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
