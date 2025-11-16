package view

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
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
			// Delete all completions for this habit (un-complete)
			completions, err := m.store.GetCompletionsByHabitId(context.Background(), m.habits[m.cursor].ID)
			if err != nil {
				log.Printf("Error retrieving completions: %s", err)
				return m, nil
			}
			for _, c := range completions {
				err := m.store.DeleteCompletion(context.Background(), c.ID)
				if err != nil {
					log.Printf("Error deleting completion: %s", err)
				}
			}
			// Remove from local completions slice
			// TODO: just query the db for completions by date
			var updated []types.Completion
			for _, c := range m.completions {
				if c.HabitID != m.habits[m.cursor].ID {
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
	header := m.appBoundaryView("Your Habits")
	b.WriteString(header)
	b.WriteString("\n\n")
	if len(m.habits) == 0 {
		b.WriteString(s.Help.Render("No habits created yet.\n\nPress 'a' to create a new one."))
	} else {
		for i, h := range m.habits {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			completed := ""
			if IsCompleted(m.completions, &h) {
				completed = "✅"
			} else {
				completed = "❌"
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
			b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, name, completed))
		}
	}
	b.WriteString("\n")
	help := s.Help.Render("a: Add new habit  |  q: Quit")
	b.WriteString(help)
	return s.Base.Render(b.String())
}