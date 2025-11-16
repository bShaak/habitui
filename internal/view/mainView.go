package view

import (
	"fmt"
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
		for _, h := range m.habits {
			name := h.Name
			if name == "" {
				name = "Unnamed"
			}
			b.WriteString(fmt.Sprintf("%s\n", name))
		}
	}
	b.WriteString("\n")
	help := s.Help.Render("a: Add new habit  |  q: Quit")
	b.WriteString(help)
	return s.Base.Render(b.String())
}