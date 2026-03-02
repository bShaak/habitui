package view

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HabitStats struct {
	Habit            types.Habit
	TotalCompletions int
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

func getStatsPeriods() []StatsPeriod {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	weekStart := todayStart.AddDate(0, 0, -6)
	monthStart := todayStart.AddDate(0, 0, -29)
	yearStart := todayEnd.AddDate(0, 0, -364)

	return []StatsPeriod{
		{Name: "Last 7 Days", StartDate: weekStart, EndDate: todayEnd},
		{Name: "Last 30 Days", StartDate: monthStart, EndDate: todayEnd},
		{Name: "Last Year", StartDate: yearStart, EndDate: todayEnd},
	}
}

func countScheduledDaysInRange(habit types.Habit, startDate, endDate time.Time) int {
	count := 0
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		dayName := getDayName(current)
		if isScheduledOnDay(habit.Frequency, dayName) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}
	return count
}

func getCompletionsForHabitInRange(completions []types.Completion, habitID int64, startDate, endDate time.Time) int {
	count := 0
	startOfRange := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endOfRange := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		if (completedAt.Equal(startOfRange) || completedAt.After(startOfRange)) &&
			(completedAt.Equal(endOfRange) || completedAt.Before(endOfRange)) {
			count++
		}
	}
	return count
}

func getHabitStreak(completions []types.Completion, habitID int64, today time.Time) (int, int) {
	type dayCompletion struct {
		date  time.Time
		count int
	}

	completionByDay := make(map[string]int)
	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		dayKey := time.Date(completedAt.Year(), completedAt.Month(), completedAt.Day(), 0, 0, 0, 0, completedAt.Location()).Format("2006-01-02")
		completionByDay[dayKey]++
	}

	currentStreak := 0
	longestStreak := 0
	tempStreak := 0

	checkDate := today
	for i := 0; i < 365; i++ {
		dayKey := checkDate.Format("2006-01-02")
		completionCount := completionByDay[dayKey]

		if completionCount > 0 {
			tempStreak++
			if i == 0 || currentStreak > 0 {
				currentStreak = tempStreak
			}
		} else {
			if currentStreak == 0 && i > 0 {
				break
			}
			if currentStreak > 0 {
				break
			}
			dayName := getDayName(checkDate)
			if isScheduledOnDay("", dayName) {
				break
			}
		}

		if tempStreak > longestStreak {
			longestStreak = tempStreak
		}

		if currentStreak == 0 && i > 0 {
			break
		}

		checkDate = checkDate.AddDate(0, 0, -1)
		if checkDate.Year() < 2020 {
			break
		}
	}

	allDates := make([]string, 0, len(completionByDay))
	for d := range completionByDay {
		allDates = append(allDates, d)
	}
	for i := 0; i < len(allDates)-1; i++ {
		for j := 0; j < len(allDates)-i-1; j++ {
			if allDates[j] > allDates[j+1] {
				allDates[j], allDates[j+1] = allDates[j+1], allDates[j]
			}
		}
	}

	tempStreak = 0
	longestStreak = 0
	for i, dayStr := range allDates {
		dayDate, _ := time.Parse("2006-01-02", dayStr)
		if i == 0 {
			tempStreak = 1
		} else {
			prevDayStr := allDates[i-1]
			prevDayDate, _ := time.Parse("2006-01-02", prevDayStr)
			diff := dayDate.Sub(prevDayDate).Hours() / 24
			if diff == 1 {
				tempStreak++
			} else {
				tempStreak = 1
			}
		}
		if tempStreak > longestStreak {
			longestStreak = tempStreak
		}
	}

	return currentStreak, longestStreak
}

func calculateStatsForHabit(habit types.Habit, completions []types.Completion, period StatsPeriod) HabitStats {
	scheduledDays := countScheduledDaysInRange(habit, period.StartDate, period.EndDate)
	totalCompletions := getCompletionsForHabitInRange(completions, habit.ID, period.StartDate, period.EndDate)

	var completionRate float64
	if scheduledDays > 0 {
		completionRate = float64(totalCompletions) / float64(scheduledDays) * 100
		if completionRate > 100 {
			completionRate = 100
		}
	}

	today := time.Now()
	currentStreak, longestStreak := getHabitStreak(completions, habit.ID, today)

	return HabitStats{
		Habit:            habit,
		TotalCompletions: totalCompletions,
		ScheduledDays:    scheduledDays,
		CompletionRate:   completionRate,
		CurrentStreak:    currentStreak,
		LongestStreak:    longestStreak,
	}
}

func GetStatsUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			defer m.store.Close()
			return m, tea.Quit
		case "left":
			if m.statsTab > 0 {
				m.statsTab--
			}
		case "right":
			if m.statsTab < 2 {
				m.statsTab++
			}
		}
	}
	return m, nil
}

func GetStatsView(m model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")

	header := m.appBoundaryView("Habit Statistics")
	b.WriteString(header)
	b.WriteString("\n\n")

	periods := getStatsPeriods()

	allCompletions, err := m.store.GetCompletionsByDateRange(context.Background(), periods[2].StartDate, periods[0].EndDate)
	if err != nil {
		log.Printf("Error fetching all completions for stats: %s", err)
	}

	tabNames := []string{"Last 7 Days", "Last 30 Days", "Last Year"}
	tabStyle := lipgloss.NewStyle().Foreground(text).Padding(0, 1)
	activeTabStyle := tabStyle.Foreground(lavender).Bold(true).Background(surface1)

	var tabs strings.Builder
	for i, name := range tabNames {
		if i == m.statsTab {
			tabs.WriteString(activeTabStyle.Render(name))
		} else {
			tabs.WriteString(tabStyle.Render(name))
		}
	}
	b.WriteString(tabs.String())
	b.WriteString("\n\n")

	period := periods[m.statsTab]

	periodLabelStyle := lipgloss.NewStyle().
		Foreground(lavender).
		Bold(true).
		Width(20)

	statLabelStyle := lipgloss.NewStyle().
		Foreground(subtext).
		Width(18)

	statValueStyle := lipgloss.NewStyle().
		Foreground(text).
		Width(10)

	var content strings.Builder

	content.WriteString(periodLabelStyle.Render(period.Name))
	content.WriteString("\n")

	separator := strings.Repeat("─", 50)
	content.WriteString(lipgloss.NewStyle().Foreground(overlay).Render(separator))
	content.WriteString("\n")

	if len(m.habits) == 0 {
		content.WriteString(s.Help.Render("No habits to show stats for."))
	} else {
		for _, habit := range m.habits {
			stats := calculateStatsForHabit(habit, allCompletions, period)

			habitColor := getHabitColor(habit.Color)
			habitNameStyle := lipgloss.NewStyle().Foreground(habitColor)

			habitName := habit.Name
			if len(habitName) > 17 {
				habitName = habitName[:15] + "…"
			}

			content.WriteString(habitNameStyle.Render(habitName))
			content.WriteString("\n")

			completedStr := fmt.Sprintf("%d/%d", stats.TotalCompletions, stats.ScheduledDays)
			content.WriteString(statLabelStyle.Render("  Completed: "))
			content.WriteString(statValueStyle.Render(completedStr))

			rateStr := fmt.Sprintf("%.0f%%", stats.CompletionRate)
			content.WriteString(statLabelStyle.Render("Rate: "))
			content.WriteString(statValueStyle.Render(rateStr))
			content.WriteString("\n")

			streakStr := fmt.Sprintf("%d days", stats.CurrentStreak)
			content.WriteString(statLabelStyle.Render("  Current Streak: "))
			content.WriteString(statValueStyle.Render(streakStr))

			longestStr := fmt.Sprintf("%d days", stats.LongestStreak)
			content.WriteString(statLabelStyle.Render("Best Streak: "))
			content.WriteString(statValueStyle.Render(longestStr))
			content.WriteString("\n")

			content.WriteString("\n")
		}
	}

	helpText := "←/→: switch tabs  |  esc: back  |  q: quit"
	help := s.Help.Render(helpText)
	content.WriteString(help)

	b.WriteString(s.ContentBox.Render(content.String()))

	return s.Base.Render(b.String())
}
