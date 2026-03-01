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

var (
	cellWidth      = 8
	habitNameWidth = 12
	dayNames       = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	yellow         = lipgloss.AdaptiveColor{Light: "#F5A623", Dark: "#F5A623"}
)

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

func getDayName(t time.Time) string {
	return strings.ToLower(t.Weekday().String())
}

func getCompletionsForHabitAndDate(completions []types.Completion, habitID int64, date time.Time) int {
	count := 0
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	for _, c := range completions {
		if c.HabitID != habitID {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
		if err != nil {
			continue
		}
		if (completedAt.Equal(startOfDay) || completedAt.After(startOfDay)) &&
			(completedAt.Equal(endOfDay) || completedAt.Before(endOfDay)) {
			count++
		}
	}
	return count
}

func isScheduledOnDay(frequency string, dayName string) bool {
	if frequency == "" || strings.ToLower(frequency) == "daily" {
		return true
	}
	days := parseFrequency(frequency)
	return days[dayName]
}

func GetCalendarUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			defer m.store.Close()
			return m, tea.Quit
		case "j":
			if len(m.habits) > 0 && m.cursor < len(m.habits)-1 {
				m.cursor++
			}
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "l":
			if m.calendarCol < 6 {
				m.calendarCol++
			}
		case "h":
			if m.calendarCol > 0 {
				m.calendarCol--
			}
		case "L":
			m.weekStart = m.weekStart.AddDate(0, 0, 7)
			weekEnd := m.weekStart.AddDate(0, 0, 6)
			completions, err := m.store.GetCompletionsByDateRange(context.Background(), m.weekStart, weekEnd)
			if err != nil {
				log.Printf("Error fetching week completions: %s", err)
				return m, nil
			}
			m.weekCompletions = completions
		case "H":
			m.weekStart = m.weekStart.AddDate(0, 0, -7)
			weekEnd := m.weekStart.AddDate(0, 0, 6)
			completions, err := m.store.GetCompletionsByDateRange(context.Background(), m.weekStart, weekEnd)
			if err != nil {
				log.Printf("Error fetching week completions: %s", err)
				return m, nil
			}
			m.weekCompletions = completions
		case "enter":
			if len(m.habits) == 0 {
				return m, nil
			}
			selectedHabit := m.habits[m.cursor]
			selectedDate := m.weekStart.AddDate(0, 0, m.calendarCol)
			dayName := getDayName(selectedDate)

			if !isScheduledOnDay(selectedHabit.Frequency, dayName) {
				return m, nil
			}

			completionCount := getCompletionsForHabitAndDate(m.weekCompletions, selectedHabit.ID, selectedDate)

			if completionCount >= selectedHabit.Goal {
				completions, err := m.store.GetCompletionsByHabitId(context.Background(), selectedHabit.ID)
				if err != nil {
					log.Printf("Error retrieving completions: %s", err)
					return m, nil
				}
				startOfDay := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), 0, 0, 0, 0, selectedDate.Location())
				endOfDay := time.Date(selectedDate.Year(), selectedDate.Month(), selectedDate.Day(), 23, 59, 59, 999999999, selectedDate.Location())

				for _, c := range completions {
					completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
					if err != nil {
						continue
					}
					if (completedAt.Equal(startOfDay) || completedAt.After(startOfDay)) &&
						(completedAt.Equal(endOfDay) || completedAt.Before(endOfDay)) {
						err := m.store.DeleteCompletion(context.Background(), c.ID)
						if err != nil {
							log.Printf("Error deleting completion: %s", err)
						}
					}
				}
				var updated []types.Completion
				for _, c := range m.weekCompletions {
					completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
					if err != nil {
						updated = append(updated, c)
						continue
					}
					if !(c.HabitID == selectedHabit.ID &&
						(completedAt.Equal(startOfDay) || completedAt.After(startOfDay)) &&
						(completedAt.Equal(endOfDay) || completedAt.Before(endOfDay))) {
						updated = append(updated, c)
					}
				}
				m.weekCompletions = updated
			} else {
				now := time.Now()
				selectedDateTime := time.Date(
					selectedDate.Year(), selectedDate.Month(), selectedDate.Day(),
					now.Hour(), now.Minute(), now.Second(), 0, now.Location(),
				)
				c, err := m.store.CreateCompletion(context.Background(), &types.Completion{
					HabitID:     selectedHabit.ID,
					CompletedAt: selectedDateTime.Format(time.RFC3339),
				})
				if err != nil {
					log.Printf("Error creating completion: %s", err)
					return m, nil
				}
				m.weekCompletions = append(m.weekCompletions, *c)
			}
		}
	}
	return m, nil
}

func GetCalendarView(m model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")

	weekEnd := m.weekStart.AddDate(0, 0, 6)
	headerText := fmt.Sprintf("Week of %s %d - %s %d, %d",
		m.weekStart.Format("Jan"), m.weekStart.Day(),
		weekEnd.Format("Jan"), weekEnd.Day(),
		m.weekStart.Year())
	header := m.appBoundaryView(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	var content strings.Builder

	headerCell := lipgloss.NewStyle().
		Foreground(indigo).
		Bold(true).
		Width(cellWidth).
		Align(lipgloss.Center)

	emptyCell := lipgloss.NewStyle().Width(habitNameWidth)
	content.WriteString(emptyCell.Render(" "))

	for i := 0; i < 7; i++ {
		date := m.weekStart.AddDate(0, 0, i)
		dayHeader := fmt.Sprintf("%s %d", dayNames[i], date.Day())
		content.WriteString(headerCell.Render(dayHeader))
	}
	content.WriteString("\n")

	separatorCell := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(cellWidth).
		Align(lipgloss.Center)
	content.WriteString(emptyCell.Render(" "))
	for i := 0; i < 7; i++ {
		content.WriteString(separatorCell.Render(strings.Repeat("-", 6)))
	}
	content.WriteString("\n")

	if len(m.habits) == 0 {
		content.WriteString(s.Help.Render("No habits created yet.\n\nPress 'a' from main view to create a new one."))
	} else {
		for row, habit := range m.habits {
			nameStyle := lipgloss.NewStyle().Width(habitNameWidth)
			if row == m.cursor {
				nameStyle = nameStyle.Foreground(lipgloss.Color("212")).Bold(true)
			}
			name := habit.Name
			if len(name) > habitNameWidth-1 {
				name = name[:habitNameWidth-2] + "…"
			}
			content.WriteString(nameStyle.Render(name))

			frequencyDays := parseFrequency(habit.Frequency)

			for col := 0; col < 7; col++ {
				date := m.weekStart.AddDate(0, 0, col)
				dayName := getDayName(date)
				completionCount := getCompletionsForHabitAndDate(m.weekCompletions, habit.ID, date)

				var cellContent string
				var cellStyle lipgloss.Style

				isScheduled := isScheduledOnDay(habit.Frequency, dayName)
				isComplete := completionCount >= habit.Goal
				isPartial := completionCount > 0 && completionCount < habit.Goal
				_, frequencyHasSpecificDays := frequencyDays[getDayName(date)]
				hasSpecificDays := len(frequencyDays) > 0 && frequencyHasSpecificDays

				cellStyle = lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)

				if row == m.cursor && col == m.calendarCol {
					if isComplete {
						cellStyle = cellStyle.
							Background(lipgloss.Color("57")).
							Foreground(green)
					} else if isPartial {
						cellStyle = cellStyle.
							Background(lipgloss.Color("57")).
							Foreground(yellow)
					} else {
						cellStyle = cellStyle.
							Background(lipgloss.Color("57")).
							Foreground(lipgloss.Color("#FFFFFF"))
					}
				}

				if isComplete {
					cellContent = "✓"
					if !(row == m.cursor && col == m.calendarCol) {
						cellStyle = cellStyle.Foreground(green)
					}
				} else if isPartial {
					cellContent = "✓"
					if !(row == m.cursor && col == m.calendarCol) {
						cellStyle = cellStyle.Foreground(yellow)
					}
				} else if isScheduled || hasSpecificDays {
					cellContent = "·"
					if !(row == m.cursor && col == m.calendarCol) {
						cellStyle = cellStyle.Foreground(lipgloss.Color("248"))
					}
				} else {
					cellContent = "-"
					if !(row == m.cursor && col == m.calendarCol) {
						cellStyle = cellStyle.Foreground(lipgloss.Color("238"))
					}
				}

				content.WriteString(cellStyle.Render(cellContent))
			}
			content.WriteString("\n")
		}
	}

	content.WriteString("\n")
	helpText := "h/l: navigate days  |  j/k: navigate habits  |  enter: toggle  |  H/L: prev/next week  |  esc: back  |  q: quit"
	help := s.Help.Render(helpText)
	content.WriteString(help)

	b.WriteString(s.ContentBox.Render(content.String()))

	return s.Base.Render(b.String())
}
