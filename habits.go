package main

import (
	"fmt"
	"log"

	// import charm bubbletea package
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type habit struct {
	name      string
	completed bool
}

// Model
type model struct {
	habits    []habit
	cursor    int
	textInput textinput.Model
	editing   bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter new habit name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		habits:    []habit{},
		cursor:    0,
		textInput: ti,
		editing:   false,
	}
}

// Init
func (m model) Init() tea.Cmd {
	return nil
}

// Update
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.editing {
				break
			} else {
				return m, tea.Quit
			}
		case "up", "k":
			if m.editing {
				break
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "down", "j":
			if m.editing {
				break
			} else {
				if m.cursor < len(m.habits)-1 {
					m.cursor++
				}
			}
		case "enter":
			if m.editing {
				m.habits[m.cursor].name = m.textInput.Value()
				m.editing = false
				m.textInput.Blur()
			} else {
				if !m.habits[m.cursor].completed {
					m.habits[m.cursor].completed = true
				} else {
					m.habits[m.cursor].completed = false
				}
			}
		case "a":
			if m.editing {
				break
			} else {
				m.habits = append(m.habits, habit{name: "New habit", completed: false})
			}
		case "d":
			// Delete a habit
			if m.editing {
				break
			} else {
				if len(m.habits) > 0 {
					m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
					if m.cursor > 0 {
						m.cursor--
					}
				}
			}

		}
	}

	if m.editing {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			// Start editing the habit name
			if m.editing {
				break
			} else {
				m.editing = true
				m.textInput.SetValue(m.habits[m.cursor].name)
				m.textInput.Focus()
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

	if m.editing {
		s += "\nEditing: " + m.textInput.View()
	}

	s += "\nPress 'q' to quit\n"

	return s
}

func main() {
	// Initialize the model
	m := initialModel()
	m.habits = []habit{
		{name: "Read for 30 minutes", completed: false},
		{name: "Exercise", completed: false},
	}

	// Start the program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
