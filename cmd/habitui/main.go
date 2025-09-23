package main

import (
	"fmt"
	"os"

	// "github.com/bShaak/habitui/internal/db"
	types "github.com/bShaak/habitui/internal/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model
type model struct {
	habits    []types.Habit
	cursor    int
	textInput textinput.Model
	editing   bool
	// dbClient  *db.DBClient
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter new habit name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	//client := db.NewDBClient()
	// habits, err := client.GetHabits()
	// if err != nil {
	// 	log.Fatalf("Error fetching habits: %s", err)
	// }

	habits := []types.Habit{
		{Name: "Run", Completed: false},
		{Name: "Yoga", Completed: false},
		{Name: "Personal Project", Completed: false},
	}

	return model{
		habits:    habits,
		cursor:    0,
		textInput: ti,
		editing:   false,
		// dbClient:  client,
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
				m.habits[m.cursor].Name = m.textInput.Value()
				m.editing = false
				m.textInput.Blur()
			} else {
				if !m.habits[m.cursor].Completed {
					m.habits[m.cursor].Completed = true
				} else {
					m.habits[m.cursor].Completed = false
				}
			}
		case "a":
			if m.editing {
				break
			} else {
				m.habits = append(m.habits, types.Habit{Name: "New habit", Completed: false})
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
				m.textInput.SetValue(m.habits[m.cursor].Name)
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
		if h.Completed {
			completed = "✅"
		}
		s += fmt.Sprintf("%s %s %s\n", cursor, h.Name, completed)
	}

	if m.editing {
		s += "\nEditing: " + m.textInput.View()
	}

	s += "\nPress 'q' to quit\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
