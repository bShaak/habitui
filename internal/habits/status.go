package habits

import (
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/models"
)

// EffectiveGoal returns at least 1.
func EffectiveGoal(goal int) int {
	if goal < 1 {
		return 1
	}
	return goal
}

// DayName returns the lowercase English weekday name for t.
func DayName(t time.Time) string {
	return strings.ToLower(t.Weekday().String())
}

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

// IsScheduledOnDay reports whether frequency includes dayName ("daily" or empty matches all days).
func IsScheduledOnDay(frequency string, dayName string) bool {
	if frequency == "" || strings.ToLower(frequency) == "daily" {
		return true
	}
	return parseFrequency(frequency)[dayName]
}

// IsDueOnDate reports whether the habit is scheduled on the given calendar day.
func IsDueOnDate(h models.Habit, date time.Time) bool {
	return IsScheduledOnDay(h.Frequency, DayName(date))
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// InDayBounds reports whether timestamp t falls within day's calendar day.
func InDayBounds(t, day time.Time) bool {
	start, end := startOfDay(day), endOfDay(day)
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}

// CompletionCountForDate counts completions for habitID that fall on date's calendar day.
func CompletionCountForDate(completions []models.Completion, habitID int64, date time.Time) int {
	count := 0
	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		if InDayBounds(completedAt, date) {
			count++
		}
	}
	return count
}

// IsCompleteForDate reports whether completions on date meet the habit goal.
func IsCompleteForDate(completions []models.Completion, h models.Habit, date time.Time) bool {
	return CompletionCountForDate(completions, h.ID, date) >= EffectiveGoal(h.Goal)
}
