package view

import (
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/habits"
	"github.com/bShaak/habitui/internal/models"
)

var allWeekdays = []string{
	"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func inDayBounds(t, day time.Time) bool {
	start, end := startOfDay(day), endOfDay(day)
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}

// normalizeFrequency stores empty or all-days schedules as "daily".
func normalizeFrequency(days []string) string {
	cleaned := make([]string, 0, len(days))
	seen := make(map[string]bool, len(days))
	for _, d := range days {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || d == "daily" || seen[d] {
			continue
		}
		seen[d] = true
		cleaned = append(cleaned, d)
	}
	if len(cleaned) == 0 || len(cleaned) == 7 {
		return "daily"
	}
	return strings.Join(cleaned, ",")
}

// frequencyDaysForForm expands "daily" into all weekdays so the multi-select shows a full schedule.
func frequencyDaysForForm(frequency string) []string {
	freq := strings.ToLower(strings.TrimSpace(frequency))
	if freq == "" || freq == "daily" {
		out := make([]string, len(allWeekdays))
		copy(out, allWeekdays)
		return out
	}
	parts := strings.Split(freq, ",")
	cleaned := make([]string, 0, len(parts))
	valid := make(map[string]bool, len(allWeekdays))
	for _, d := range allWeekdays {
		valid[d] = true
	}
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if !valid[p] || seen[p] {
			continue
		}
		seen[p] = true
		cleaned = append(cleaned, p)
	}
	if len(cleaned) == 0 {
		out := make([]string, len(allWeekdays))
		copy(out, allWeekdays)
		return out
	}
	return cleaned
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

// parseFrequency builds a set of day tokens from a frequency string (view-specific calendar UI).
func parseFrequency(frequency string) map[string]bool {
	days := make(map[string]bool)
	for _, d := range strings.Split(strings.ToLower(frequency), ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			days[d] = true
		}
	}
	return days
}

// isCompleted reports whether today's pre-filtered completion list meets the habit goal.
// It uses todayCompletionCount (no timestamp re-parse) because m.completions is already today-scoped.
func isCompleted(completions []models.Completion, h models.Habit) bool {
	return todayCompletionCount(completions, h.ID) >= habits.EffectiveGoal(h.Goal)
}

func todayCompletionCount(completions []models.Completion, habitID int64) int {
	count := 0
	for _, c := range completions {
		if c.HabitID == habitID {
			count++
		}
	}
	return count
}

func formatHabitLabel(habit models.Habit) string {
	if habit.Icon != "" {
		return habit.Icon + " " + habit.Name
	}
	return habit.Name
}

func getMonday(t time.Time) time.Time {
	weekday := t.Weekday()
	daysSinceMonday := (int(weekday) + 6) % 7
	return startOfDay(t.AddDate(0, 0, -daysSinceMonday))
}
