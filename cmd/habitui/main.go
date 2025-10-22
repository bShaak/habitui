package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model
type model struct {
	habits    []types.Habit
	habitLogs []types.HabitLog
	cursor    int
	textInput textinput.Model
	editing   bool
	store  *storage.SQLiteStore
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter new habit name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	store, err := storage.OpenSQLite()
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
		defer store.Close()
	}

	habits, err := store.ListHabits(context.Background())
	if err != nil {
		log.Fatalf("Error fetching habits: %s", err)
	}

	habitLogs := make([]types.HabitLog, len(habits))
	for i, h := range habits {
		habitLogs[i] = types.HabitLog{
			HabitID: h.ID,
			Timestamp: time.Time{},
		}
	}

	return model{
		habits:    habits,
		habitLogs: habitLogs,
		cursor:    0,
		textInput: ti,
		editing:   false,
		store:     store,
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
			defer m.store.Close()
			return m, tea.Quit
		case "q":
			if m.editing {
				break
			} else {
				defer m.store.Close()
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
				err := m.store.UpdateHabit(context.Background(), &m.habits[m.cursor])
				if err != nil {
					log.Printf("Error updating habit: %s", err)
				}
				m.editing = false
				m.textInput.Blur()
			} else {
				if m.habitLogs[m.cursor].Timestamp.IsZero() {
					m.habitLogs[m.cursor].Timestamp = time.Now()
				} else {
					m.habitLogs[m.cursor].Timestamp = time.Time{}
				}
			}
		case "a":
			if m.editing {
				break
			} else {
				habit, err := m.store.CreateHabit(context.Background(), &types.Habit{Name: "New habit"})
				if err != nil {
					log.Printf("Error creating habit: %s", err)
					break
				}
				m.habits = append(m.habits, *habit)
				m.habitLogs = append(m.habitLogs, types.HabitLog{HabitID: m.habits[len(m.habits)-1].ID, Timestamp: time.Now()})
			}
		case "x":
			// Delete a habit
			if m.editing {
				break
			} else {
				if len(m.habits) > 0 {
					err := m.store.DeleteHabit(context.Background(), m.habits[m.cursor].ID)
					if err != nil {
						log.Printf("Error deleting habit: %s", err.Error())
						break
					}
					m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
					m.habitLogs = append(m.habitLogs[:m.cursor], m.habitLogs[m.cursor+1:]...)
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
		if m.habitLogs[i].Timestamp.IsZero() {
			completed = "❌"
		} else {
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
	_, err := p.Run(); 
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
