package view

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

			habitName := formatHabitLabel(habit)
			if len(habitName) > 20 {
				habitName = habitName[:18] + "…"
			}

			content.WriteString(habitNameStyle.Render(habitName))
			content.WriteString("\n")

			completedStr := fmt.Sprintf("%d/%d", stats.GoalDaysMet, stats.ScheduledDays)
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
