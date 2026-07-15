package view

import (
	"testing"
	"time"
	"unicode/utf8"

	"github.com/bShaak/habitui/internal/models"
)

func TestNormalizeFrequency(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{name: "empty", in: nil, want: "daily"},
		{name: "all days", in: allWeekdays, want: "daily"},
		{name: "weekdays", in: []string{"monday", "wednesday"}, want: "monday,wednesday"},
		{name: "daily token", in: []string{"Daily"}, want: "daily"},
		{name: "duplicates", in: []string{"monday", " Monday ", "monday"}, want: "monday"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeFrequency(tt.in); got != tt.want {
				t.Fatalf("normalizeFrequency(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFrequencyDaysForForm(t *testing.T) {
	days := frequencyDaysForForm("daily")
	if len(days) != 7 {
		t.Fatalf("expected 7 days for daily, got %d", len(days))
	}
	days = frequencyDaysForForm("Daily")
	if len(days) != 7 {
		t.Fatalf("expected 7 days for Daily, got %d", len(days))
	}
	days = frequencyDaysForForm("monday,friday")
	if len(days) != 2 || days[0] != "monday" || days[1] != "friday" {
		t.Fatalf("unexpected days: %v", days)
	}
}

func TestIsScheduledOnDay(t *testing.T) {
	if !isScheduledOnDay("daily", "monday") {
		t.Fatal("daily should schedule monday")
	}
	if !isScheduledOnDay("", "sunday") {
		t.Fatal("empty frequency should schedule all days")
	}
	if isScheduledOnDay("monday,wednesday", "tuesday") {
		t.Fatal("tuesday should not be scheduled")
	}
	if !isScheduledOnDay("monday,wednesday", "wednesday") {
		t.Fatal("wednesday should be scheduled")
	}
}

func TestEffectiveGoal(t *testing.T) {
	if effectiveGoal(0) != 1 {
		t.Fatal("expected goal clamp to 1")
	}
	if effectiveGoal(3) != 3 {
		t.Fatal("expected goal 3 unchanged")
	}
}

func TestTruncateRunes(t *testing.T) {
	name := "❤️ Super Long Habit Name Here"
	got := truncateRunes(name, 10)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate produced invalid UTF-8: %q", got)
	}
	if utf8.RuneCountInString(got) > 10 {
		t.Fatalf("truncate too long: %q (%d runes)", got, utf8.RuneCountInString(got))
	}
}

func TestGetHabitStreak(t *testing.T) {
	loc := time.Local
	today := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	habit := models.Habit{
		ID:        1,
		Frequency: "daily",
		Goal:      1,
		StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, loc).Format(time.RFC3339),
	}

	completions := []models.Completion{
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 8, 9, 0, 0, 0, loc).Format(time.RFC3339)},
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 9, 9, 0, 0, 0, loc).Format(time.RFC3339)},
		// today incomplete — current streak should still be 2
	}

	current, longest := getHabitStreak(habit, completions, today)
	if current != 2 {
		t.Fatalf("current streak = %d, want 2", current)
	}
	if longest != 2 {
		t.Fatalf("longest streak = %d, want 2", longest)
	}
}

func TestGetHabitStreakCountsOffScheduleCompletion(t *testing.T) {
	loc := time.Local
	// Friday — Code habit is not scheduled on Fridays.
	today := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	habit := models.Habit{
		ID:        1,
		Frequency: "monday,tuesday,wednesday,thursday,sunday",
		Goal:      1,
		StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, loc).Format(time.RFC3339),
	}
	completions := []models.Completion{
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 10, 12, 51, 0, 0, loc).Format(time.RFC3339)},
	}

	current, longest := getHabitStreak(habit, completions, today)
	if current != 1 {
		t.Fatalf("current streak = %d, want 1 (off-schedule completion should count)", current)
	}
	if longest != 1 {
		t.Fatalf("longest streak = %d, want 1", longest)
	}

	period := statsPeriod{
		Name:      "Last 7 Days",
		StartDate: startOfDay(today).AddDate(0, 0, -6),
		EndDate:   endOfDay(today),
	}
	stats := calculateStatsForHabit(habit, completions, period)
	if stats.GoalDaysMet != 1 {
		t.Fatalf("GoalDaysMet = %d, want 1", stats.GoalDaysMet)
	}
}

func TestGetHabitStreakSkipsUnscheduledDays(t *testing.T) {
	loc := time.Local
	// Friday Jul 10, 2026
	today := time.Date(2026, 7, 10, 12, 0, 0, 0, loc)
	habit := models.Habit{
		ID:        1,
		Frequency: "monday,wednesday,friday",
		Goal:      1,
		StartDate: time.Date(2026, 7, 1, 0, 0, 0, 0, loc).Format(time.RFC3339),
	}
	completions := []models.Completion{
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 6, 9, 0, 0, 0, loc).Format(time.RFC3339)},  // Mon
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 8, 9, 0, 0, 0, loc).Format(time.RFC3339)},  // Wed
		{HabitID: 1, CompletedAt: time.Date(2026, 7, 10, 9, 0, 0, 0, loc).Format(time.RFC3339)}, // Fri
	}

	current, longest := getHabitStreak(habit, completions, today)
	if current != 3 {
		t.Fatalf("current streak = %d, want 3", current)
	}
	if longest != 3 {
		t.Fatalf("longest streak = %d, want 3", longest)
	}
}

func TestGetMondayNormalizesToMidnight(t *testing.T) {
	loc := time.Local
	wednesday := time.Date(2026, 7, 8, 15, 45, 30, 0, loc)
	monday := getMonday(wednesday)
	if monday.Weekday() != time.Monday {
		t.Fatalf("expected Monday, got %s", monday.Weekday())
	}
	if monday.Hour() != 0 || monday.Minute() != 0 || monday.Second() != 0 {
		t.Fatalf("expected midnight, got %v", monday)
	}
}

func TestNeedsDayRefresh(t *testing.T) {
	loc := time.Local
	yesterday := time.Date(2026, 7, 14, 23, 0, 0, 0, loc)
	todayMorning := time.Date(2026, 7, 15, 8, 30, 0, 0, loc)
	todayEvening := time.Date(2026, 7, 15, 22, 0, 0, 0, loc)

	if !needsDayRefresh(time.Time{}, todayMorning) {
		t.Fatal("zero viewDay should need refresh")
	}
	if !needsDayRefresh(yesterday, todayMorning) {
		t.Fatal("overnight rollover should need refresh")
	}
	if needsDayRefresh(todayMorning, todayEvening) {
		t.Fatal("same calendar day should not need refresh")
	}
}
