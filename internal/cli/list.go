package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
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

	svc, err := openService(dbPath)
	if err != nil {
		return err
	}
	defer svc.Close()

	resp, err := svc.DaySummary(context.Background(), date)
	if err != nil {
		return err
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
