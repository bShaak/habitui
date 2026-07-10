package view

import (
	"fmt"
	"time"

	types "github.com/bShaak/habitui/internal/models"
)

type HabitStats struct {
	Habit            types.Habit
	TotalCompletions int
	GoalDaysMet      int
	ScheduledDays    int
	CompletionRate   float64
	CurrentStreak    int
	LongestStreak    int
}

type StatsPeriod struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func getStatsPeriods() []StatsPeriod {
	now := time.Now()
	todayStart := startOfDay(now)
	todayEnd := endOfDay(now)

	return []StatsPeriod{
		{Name: "Last 7 Days", StartDate: todayStart.AddDate(0, 0, -6), EndDate: todayEnd},
		{Name: "Last 30 Days", StartDate: todayStart.AddDate(0, 0, -29), EndDate: todayEnd},
		{Name: "Last Year", StartDate: todayStart.AddDate(0, 0, -364), EndDate: todayEnd},
	}
}

func countScheduledDaysInRange(habit types.Habit, startDate, endDate time.Time) int {
	count := 0
	current := startOfDay(startDate)
	end := startOfDay(endDate)
	for !current.After(end) {
		if isScheduledOnDay(habit.Frequency, getDayName(current)) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}
	return count
}

func completionsByDay(completions []types.Completion, habitID int64) map[string]int {
	byDay := make(map[string]int)
	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		byDay[startOfDay(completedAt).Format("2006-01-02")]++
	}
	return byDay
}

func getCompletionsForHabitInRange(completions []types.Completion, habitID int64, startDate, endDate time.Time) int {
	count := 0
	start := startOfDay(startDate)
	end := endOfDay(endDate)
	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		if (completedAt.Equal(start) || completedAt.After(start)) &&
			(completedAt.Equal(end) || completedAt.Before(end)) {
			count++
		}
	}
	return count
}

func countGoalDaysMetInRange(habit types.Habit, completions []types.Completion, startDate, endDate time.Time) int {
	byDay := completionsByDay(completions, habit.ID)
	goal := habit.Goal
	if goal < 1 {
		goal = 1
	}
	met := 0
	current := startOfDay(startDate)
	end := startOfDay(endDate)
	for !current.After(end) {
		if isScheduledOnDay(habit.Frequency, getDayName(current)) {
			if byDay[current.Format("2006-01-02")] >= goal {
				met++
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return met
}

// getHabitStreak returns current and longest streaks.
// A scheduled day counts only when completions that day >= goal.
// Unscheduled days neither count nor break the streak.
func getHabitStreak(habit types.Habit, completions []types.Completion, today time.Time) (int, int) {
	byDay := completionsByDay(completions, habit.ID)
	goal := habit.Goal
	if goal < 1 {
		goal = 1
	}

	dayMet := func(d time.Time) bool {
		return byDay[startOfDay(d).Format("2006-01-02")] >= goal
	}

	// Current streak: walk backward from today.
	currentStreak := 0
	checkDate := startOfDay(today)
	for i := 0; i < 365*5; i++ {
		scheduled := isScheduledOnDay(habit.Frequency, getDayName(checkDate))
		if !scheduled {
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		if dayMet(checkDate) {
			currentStreak++
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		// Today incomplete: streak can still continue from yesterday.
		if i == 0 {
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		break
	}

	// Longest streak: scan from habit start (or 2 years back) through today.
	start := startOfDay(today).AddDate(-2, 0, 0)
	if habit.StartDate != "" {
		if parsed, err := time.Parse(time.RFC3339, habit.StartDate); err == nil {
			start = startOfDay(parsed)
		}
	}
	longestStreak := 0
	tempStreak := 0
	for d := start; !d.After(startOfDay(today)); d = d.AddDate(0, 0, 1) {
		if !isScheduledOnDay(habit.Frequency, getDayName(d)) {
			continue
		}
		if dayMet(d) {
			tempStreak++
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
		} else {
			tempStreak = 0
		}
	}

	return currentStreak, longestStreak
}

func calculateStatsForHabit(habit types.Habit, completions []types.Completion, period StatsPeriod) HabitStats {
	scheduledDays := countScheduledDaysInRange(habit, period.StartDate, period.EndDate)
	goalDaysMet := countGoalDaysMetInRange(habit, completions, period.StartDate, period.EndDate)
	totalCompletions := getCompletionsForHabitInRange(completions, habit.ID, period.StartDate, period.EndDate)

	var completionRate float64
	if scheduledDays > 0 {
		completionRate = float64(goalDaysMet) / float64(scheduledDays) * 100
		if completionRate > 100 {
			completionRate = 100
		}
	}

	currentStreak, longestStreak := getHabitStreak(habit, completions, time.Now())

	return HabitStats{
		Habit:            habit,
		TotalCompletions: totalCompletions,
		GoalDaysMet:      goalDaysMet,
		ScheduledDays:    scheduledDays,
		CompletionRate:   completionRate,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
	}
}

func formatHabitLabel(habit types.Habit) string {
	if habit.Icon != "" {
		return fmt.Sprintf("%s %s", habit.Icon, habit.Name)
	}
	return habit.Name
}
