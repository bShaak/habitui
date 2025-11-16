package view

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
			name := h.Name
			if name == "" {
				name = "Unnamed"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, name))
		}
	}
	b.WriteString("\n")
	help := s.Help.Render("a: Add new habit  |  q: Quit")
	b.WriteString(help)
	return s.Base.Render(b.String())
}