package view

import (
	"fmt"
	"time"

	"github.com/bShaak/habitui/internal/models"
)

type habitStats struct {
	Habit            models.Habit
	TotalCompletions int
	GoalDaysMet      int
	ScheduledDays    int
	CompletionRate   float64
	CurrentStreak    int
	LongestStreak    int
}

type statsPeriod struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

func getStatsPeriods() []statsPeriod {
	now := time.Now()
	todayStart := startOfDay(now)
	todayEnd := endOfDay(now)

	return []statsPeriod{
		{Name: "Last 7 Days", StartDate: todayStart.AddDate(0, 0, -6), EndDate: todayEnd},
		{Name: "Last 30 Days", StartDate: todayStart.AddDate(0, 0, -29), EndDate: todayEnd},
		{Name: "Last Year", StartDate: todayStart.AddDate(0, 0, -364), EndDate: todayEnd},
	}
}

func countScheduledDaysInRange(habit models.Habit, startDate, endDate time.Time) int {
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

func completionsByDay(completions []models.Completion, habitID int64) map[string]int {
	byDay := make(map[string]int)
	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		// Normalize to local calendar day so keys match startOfDay(time.Now()).
		localDay := startOfDay(completedAt.In(time.Local)).Format("2006-01-02")
		byDay[localDay]++
	}
	return byDay
}

func getCompletionsForHabitInRange(completions []models.Completion, habitID int64, startDate, endDate time.Time) int {
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

func countGoalDaysMetInRange(habit models.Habit, completions []models.Completion, startDate, endDate time.Time) int {
	byDay := completionsByDay(completions, habit.ID)
	goal := effectiveGoal(habit.Goal)
	met := 0
	current := startOfDay(startDate)
	end := startOfDay(endDate)
	for !current.After(end) {
		dayKey := current.Format("2006-01-02")
		if byDay[dayKey] >= goal {
			// Count any day the goal was met, including off-schedule check-ins.
			met++
		}
		current = current.AddDate(0, 0, 1)
	}
	return met
}

// getHabitStreak returns current and longest streaks.
// A day counts when completions that day >= goal.
// Unscheduled days with no completions neither count nor break the streak.
// Unscheduled days that were completed do count (so off-day check-ins aren't ignored).
func getHabitStreak(habit models.Habit, completions []models.Completion, today time.Time) (int, int) {
	byDay := completionsByDay(completions, habit.ID)
	goal := effectiveGoal(habit.Goal)

	dayMet := func(d time.Time) bool {
		return byDay[startOfDay(d).In(time.Local).Format("2006-01-02")] >= goal
	}

	// Current streak: walk backward from today.
	currentStreak := 0
	checkDate := startOfDay(today)
	graceForToday := true
	maxLookbackDays := 365 * streakLookbackYears
	for i := 0; i < maxLookbackDays; i++ {
		scheduled := isScheduledOnDay(habit.Frequency, getDayName(checkDate))
		met := dayMet(checkDate)

		if !scheduled && !met {
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		if met {
			currentStreak++
			graceForToday = false
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		// Scheduled (or otherwise relevant) day missed.
		// Allow one grace skip so an incomplete today doesn't zero the streak.
		if graceForToday {
			graceForToday = false
			checkDate = checkDate.AddDate(0, 0, -1)
			continue
		}
		break
	}

	// Longest streak: scan from habit start (or lookback window) through today.
	start := startOfDay(today).AddDate(-streakLookbackYears, 0, 0)
	if habit.StartDate != "" {
		if parsed, err := time.Parse(time.RFC3339, habit.StartDate); err == nil {
			parsedStart := startOfDay(parsed.In(time.Local))
			if parsedStart.After(start) {
				start = parsedStart
			}
		}
	}
	longestStreak := 0
	tempStreak := 0
	for d := start; !d.After(startOfDay(today)); d = d.AddDate(0, 0, 1) {
		scheduled := isScheduledOnDay(habit.Frequency, getDayName(d))
		met := dayMet(d)
		if !scheduled && !met {
			continue
		}
		if met {
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

func calculateStatsForHabit(habit models.Habit, completions []models.Completion, period statsPeriod) habitStats {
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

	return habitStats{
		Habit:            habit,
		TotalCompletions: totalCompletions,
		GoalDaysMet:      goalDaysMet,
		ScheduledDays:    scheduledDays,
		CompletionRate:   completionRate,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
	}
}

func formatRate(rate float64) string {
	return fmt.Sprintf("%.0f%%", rate)
}
