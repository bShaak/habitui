package view

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func updateMain(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "enter":
				return deleteSelectedHabit(m), nil
			case "n", "x":
				m.confirmingDelete = false
				return m, nil
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "a":
			m.statusMsg = ""
			form, fields := createHabitForm()
			m.form = form
			m.formFields = fields
			applyFormSize(m.form, m.width, m.height)
			m.scrollOffset = 0
			m.screen = screenCreateHabit
			return m, m.form.Init()
		case "c":
			m.statusMsg = ""
			habits, err := m.store.ListHabits(context.Background())
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
			}
			m.habits = habits
			weekEnd := m.weekStart.AddDate(0, 0, 6)
			completions, err := m.store.GetCompletionsByDateRange(context.Background(), m.weekStart, weekEnd)
			if err != nil {
				log.Printf("Error fetching week completions: %s", err)
			}
			m.weekCompletions = completions
			if m.cursor >= len(m.habits) {
				m.cursor = 0
			}
			m.calendarCol = 0
			m.scrollOffset = 0
			m.screen = screenCalendar
			return m, nil
		case "s":
			m.statusMsg = ""
			habits, err := m.store.ListHabits(context.Background())
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
			}
			m.habits = habits
			if m.cursor >= len(m.habits) {
				m.cursor = 0
			}
			statsCompletions, err := m.store.ListCompletions(context.Background())
			if err != nil {
				log.Printf("Error fetching completions for stats: %s", err)
			}
			m.statsCompletions = statsCompletions
			m.statsTab = 0
			m.scrollOffset = 0
			m.screen = screenStats
			return m, nil
		case "e":
			m.statusMsg = ""
			if len(m.habits) == 0 {
				return m, nil
			}
			form, fields := editHabitForm(m.habits[m.cursor])
			m.form = form
			m.formFields = fields
			applyFormSize(m.form, m.width, m.height)
			m.scrollOffset = 0
			m.screen = screenEditHabit
			return m, m.form.Init()
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "j":
			if m.cursor < len(m.habits)-1 {
				m.cursor++
			}
		case "x":
			if len(m.habits) > 0 {
				m.confirmingDelete = true
			}
		case "t":
			return cycleTheme(m), nil
		case "enter":
			if len(m.habits) == 0 {
				return m, nil
			}
			selected := m.habits[m.cursor]
			m.statusMsg = ""
			updated, err := m.toggleDayCompletion(selected, time.Now(), m.completions)
			if err != nil {
				log.Printf("Error toggling completion: %s", err)
				return m, nil
			}
			m.completions = updated
			m = refreshStreakCompletions(m)
		}
	}
	return m, nil
}

func deleteSelectedHabit(m Model) Model {
	if len(m.habits) == 0 {
		m.confirmingDelete = false
		return m
	}
	err := m.store.DeleteHabit(context.Background(), m.habits[m.cursor].ID)
	if err != nil {
		log.Printf("Error deleting habit: %s", err)
		m.confirmingDelete = false
		return m
	}
	if m.cursor == len(m.habits)-1 {
		m.habits = m.habits[:m.cursor]
	} else {
		m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
	}
	if m.cursor > 0 && m.cursor >= len(m.habits) {
		m.cursor--
	}
	m.confirmingDelete = false
	return m
}

func statusStyle(msg string) lipgloss.Style {
	if strings.HasPrefix(msg, "Theme:") {
		return lipgloss.NewStyle().Foreground(primary)
	}
	return lipgloss.NewStyle().Foreground(red)
}

func viewMain(m Model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	header := m.appBoundaryView("Today's Habits")
	b.WriteString(header)
	b.WriteString("\n\n")
	var content strings.Builder
	if len(m.habits) == 0 {
		content.WriteString(s.Help.Render("No habits created yet.\n\nPress 'a' to create a new one.  |  t: Theme"))
		if m.statusMsg != "" {
			content.WriteString("\n\n")
			content.WriteString(statusStyle(m.statusMsg).Render(m.statusMsg))
		}
	} else {
		for i, h := range m.habits {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			habitColor := getHabitColor(h.Color)
			scheduledToday := isScheduledOnDay(h.Frequency, getDayName(time.Now()))
			completed := ""
			if isCompleted(m.completions, h) {
				completed = "✓"
			} else if scheduledToday {
				completionCount := todayCompletionCount(m.completions, h.ID)
				completed = fmt.Sprintf("✗ (%d/%d)", completionCount, effectiveGoal(h.Goal))
			}
			name := formatHabitLabel(h)
			if name == "" || strings.TrimSpace(h.Name) == "" {
				name = "Unnamed"
				if h.Icon != "" {
					name = h.Icon + " Unnamed"
				}
			}
			currentStreak, _ := getHabitStreak(h, m.streakCompletions, time.Now())
			streakText := ""
			if currentStreak >= 3 {
				streakText = fmt.Sprintf(" 🔥 %d", currentStreak)
			}
			nameStyle := lipgloss.NewStyle().Foreground(habitColor)
			completedStyle := lipgloss.NewStyle().Foreground(habitColor)
			streakStyle := lipgloss.NewStyle().Foreground(orange)
			content.WriteString(fmt.Sprintf("%s %s %s%s\n",
				cursor,
				nameStyle.Render(name),
				completedStyle.Render(completed),
				streakStyle.Render(streakText),
			))
		}
		content.WriteString("\n")
		if m.confirmingDelete {
			habitName := formatHabitLabel(m.habits[m.cursor])
			confirm := lipgloss.NewStyle().Foreground(red).Bold(true).Render(
				fmt.Sprintf("Delete %q?  y: confirm  |  n/esc: cancel", habitName),
			)
			content.WriteString(confirm)
		} else {
			if m.statusMsg != "" {
				content.WriteString(statusStyle(m.statusMsg).Render(m.statusMsg))
				content.WriteString("\n")
			}
			help := s.Help.Render("a: Add  |  c: Calendar  |  s: Stats  |  e: Edit  |  x: Delete  |  t: Theme  |  enter: Toggle  |  q: Quit")
			content.WriteString(help)
		}
	}
	b.WriteString(s.ContentBox.Render(content.String()))
	return s.Base.Render(b.String())
}
