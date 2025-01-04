package main

import (
	"fmt"
	"log"

	// import charm bubbletea package
	tea "github.com/charmbracelet/bubbletea"
)

type habit struct {
	name      string
	completed bool
}

// Model
type model struct {
	habits []habit
	cursor int
}

// Init
func (m model) Init() tea.Cmd {
	return nil
}

// Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.habits)-1 {
				m.cursor++
			}
		case "enter", " ":
			if !m.habits[m.cursor].completed {
				m.habits[m.cursor].completed = true
			} else {
				m.habits[m.cursor].completed = false
			}
		}
	}
	return m, nil
}

// View
func (m model) View() string {
	s := "Habits\n\n"

	for i, h := range m.habits {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		completed := ""
		if h.completed {
			completed = "✅"
		}
		s += fmt.Sprintf("%s %s %s\n", cursor, h.name, completed)
	}

	s += "\nPress 'q' to quit\n"

	return s
}

func main() {
	// Initialize the model
	m := model{
		habits: []habit{
			{name: "Read for 30 minutes", completed: false},
			{name: "Exercise", completed: false},
		},
	}

	// Start the program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
