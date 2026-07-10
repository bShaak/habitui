package view

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cellWidth      = 8
	habitNameWidth = 12
	dayNames       = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
)

func updateCalendar(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
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
			updated, err := m.toggleDayCompletion(selectedHabit, selectedDate, m.weekCompletions)
			if err != nil {
				log.Printf("Error toggling completion: %s", err)
				return m, nil
			}
			m.weekCompletions = updated
		}
	}
	return m, nil
}

func viewCalendar(m Model) string {
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
		Foreground(blue).
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
		Foreground(overlay).
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
			habitColor := getHabitColor(habit.Color)
			nameStyle := lipgloss.NewStyle().Width(habitNameWidth).Foreground(habitColor)
			if row == m.cursor {
				nameStyle = nameStyle.Bold(true)
			}
			name := formatHabitLabel(habit)
			if len([]rune(name)) > habitNameWidth-1 {
				runes := []rune(name)
				name = string(runes[:habitNameWidth-2]) + "…"
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
				isComplete := completionCount >= effectiveGoal(habit.Goal)
				isPartial := completionCount > 0 && completionCount < effectiveGoal(habit.Goal)
				_, frequencyHasSpecificDays := frequencyDays[getDayName(date)]
				hasSpecificDays := len(frequencyDays) > 0 && frequencyHasSpecificDays

				cellStyle = lipgloss.NewStyle().Width(cellWidth).Align(lipgloss.Center)

				if row == m.cursor && col == m.calendarCol {
					if isComplete {
						cellStyle = cellStyle.
							Background(surface1).
							Foreground(green)
					} else if isPartial {
						cellStyle = cellStyle.
							Background(surface1).
							Foreground(yellow)
					} else {
						cellStyle = cellStyle.
							Background(surface1).
							Foreground(text)
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
						cellStyle = cellStyle.Foreground(overlay)
					}
				} else {
					cellContent = "-"
					if !(row == m.cursor && col == m.calendarCol) {
						cellStyle = cellStyle.Foreground(overlay)
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
