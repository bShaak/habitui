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

func IsCompleted(completions []types.Completion, h *types.Habit) bool {
	completionCount := 0
	for _, c := range completions {
		if c.HabitID == h.ID {
			completionCount++
		}
	}
	return completionCount == h.Goal
}

func getHabitColor(color string) lipgloss.Color {
	return GetHabitColor(color)
}

// Update
func GetMainUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			defer m.store.Close()
			return m, tea.Quit
		case "a":
			m.form = CreateHabit()
			m.getView = GetCreateHabitView
			m.getUpdate = GetCreateHabitUpdate
			return m, m.form.Init()
		case "c":
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
			m.getView = GetCalendarView
			m.getUpdate = GetCalendarUpdate
			return m, nil
		case "s":
			habits, err := m.store.ListHabits(context.Background())
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
			}
			m.habits = habits
			if m.cursor >= len(m.habits) {
				m.cursor = 0
			}
			m.statsTab = 0
			m.getView = GetStatsView
			m.getUpdate = GetStatsUpdate
			return m, nil
		case "e":
			m.form = EditHabit(m.habits[m.cursor])
			m.getView = GetEditHabitView
			m.getUpdate = GetEditHabitUpdate
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
				err := m.store.DeleteHabit(context.Background(), m.habits[m.cursor].ID)
				if err != nil {
					log.Printf("Error deleting habit: %s", err)
					return m, nil
				}
				if m.cursor == len(m.habits)-1 {
					m.habits = m.habits[:m.cursor]
					// m.completions = m.completions[:m.cursor]
				} else {
					m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
					// m.completions = append(m.completions[:m.cursor], m.completions[m.cursor+1:]...)
				}
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "enter":
			if !IsCompleted(m.completions, &m.habits[m.cursor]) {
				c, err := m.store.CreateCompletion(context.Background(), &types.Completion{HabitID: m.habits[m.cursor].ID, CompletedAt: time.Now().Format(time.RFC3339)})
				if err != nil {
					log.Printf("Error creating completion: %s", err)
					return m, nil
				}
				m.completions = append(m.completions, *c)
			} else {
				completions, err := m.store.GetCompletionsByHabitId(context.Background(), m.habits[m.cursor].ID)
				if err != nil {
					log.Printf("Error retrieving completions: %s", err)
					return m, nil
				}
				now := time.Now()
				startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

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
				for _, c := range m.completions {
					completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
					if err != nil {
						updated = append(updated, c)
						continue
					}
					if !(c.HabitID == m.habits[m.cursor].ID &&
						(completedAt.Equal(startOfDay) || completedAt.After(startOfDay)) &&
						(completedAt.Equal(endOfDay) || completedAt.Before(endOfDay))) {
						updated = append(updated, c)
					}
				}
				m.completions = updated
			}
		}
	}
	return m, nil
}

// View
func GetMainView(m model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	header := m.appBoundaryView("Today's Habits")
	b.WriteString(header)
	b.WriteString("\n\n")
	var content strings.Builder
	if len(m.habits) == 0 {
		content.WriteString(s.Help.Render("No habits created yet.\n\nPress 'a' to create a new one."))
	} else {
		for i, h := range m.habits {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			habitColor := getHabitColor(h.Color)
			completed := ""
			if IsCompleted(m.completions, &h) {
				completed = "✓"
			} else {
				completed = "✗"
				completionCount := 0
				for _, c := range m.completions {
					if c.HabitID == h.ID {
						completionCount++
					}
				}
				completionCountText := fmt.Sprintf("(%d/%d)", completionCount, h.Goal)
				completed = fmt.Sprintf("%s %s", completed, completionCountText)
			}
			name := h.Name
			if name == "" {
				name = "Unnamed"
			}
			nameStyle := lipgloss.NewStyle().Foreground(habitColor)
			completedStyle := lipgloss.NewStyle().Foreground(habitColor)
			content.WriteString(fmt.Sprintf("%s %s %s\n", cursor, nameStyle.Render(name), completedStyle.Render(completed)))
		}
		content.WriteString("\n")
		help := s.Help.Render("a: Add new habit  |  c: Calendar  |  s: Stats  |  e: Edit  |  x: Delete  |  enter: Toggle  |  q: Quit")
		content.WriteString(help)
	}
	b.WriteString(s.ContentBox.Render(content.String()))
	return s.Base.Render(b.String())
}
