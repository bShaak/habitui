package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	"github.com/bShaak/habitui/internal/storage"
	"github.com/bShaak/habitui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

// Model
type formModel struct {
	habits    []types.Habit
	completions []types.Completion
	store  *storage.SQLiteStore
	newHabitForm *huh.Form
	viewState  string
	lg         *lipgloss.Renderer
	styles     *Styles
}

func InitialFormModel() formModel {
	store, err := storage.OpenSQLite()
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
		defer store.Close()
	}

	habits, err := store.ListHabits(context.Background())
	if err != nil {
		log.Fatalf("Error fetching habits: %s", err)
	}

	completions, err := store.ListCompletions(context.Background())
	if err != nil {
		log.Fatalf("Error fetching completions: %s", err)
	}

	lg := lipgloss.DefaultRenderer()

	return formModel{
		habits:    habits,
		completions: completions,
		newHabitForm: nil,
		store:     store,
		viewState: "main",
		lg:        lg,
		styles:    NewStyles(lg),
	}
}

// // Init
func (m formModel) Init() tea.Cmd {
	return nil
}

// // Update
func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			defer m.store.Close()
			return m, tea.Quit
		case "esc":
			if m.viewState == "form" {
				m.viewState = "main"
				return m, nil
			}
			return m, tea.Quit
		case "q":
			if m.viewState == "main" {
				defer m.store.Close()
				return m, tea.Quit
			}
		case "a":
			if m.viewState == "main" {
				m.newHabitForm = ui.CreateHabit()
				m.viewState = "form"
				return m, m.newHabitForm.Init()
			}
		}
	}

	if m.viewState == "form" {
		form, cmd := m.newHabitForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.newHabitForm = f
			cmds = append(cmds, cmd)
		}
		if m.newHabitForm.State == huh.StateCompleted {
			name := strings.TrimSpace(m.newHabitForm.GetString("name"))
			if name == "" {
				name = "Unnamed Habit"
			}
			description := m.newHabitForm.GetString("description")
			// frequency := m.newHabitForm.GetString("Frequency")
			freq := m.newHabitForm.Get("frequency").([]string)
			frequency := strings.Join(freq, ", ")
			if frequency == "" {
					frequency = "Daily"  // Default
			}
			fmt.Printf("Frequency: %s\n", frequency)
			goal := m.newHabitForm.GetString("goal")
			// startDate := m.newHabitForm.GetString("startDate")
			goalInt, err := strconv.Atoi(goal)
			if err != nil {
				log.Fatalf("Error converting goal to int: %s", err)
			}
			habit := types.Habit{
				Name:  name,
				Description: description,
				Frequency: frequency,
				Goal: goalInt,
				StartDate: time.Now().Format(time.RFC3339),
			}
			m.habits = append(m.habits, habit)
			_, err = m.store.CreateHabit(context.Background(), &habit)
			if err != nil {
				log.Fatalf("Error creating habit: %s", err)
			}
			m.viewState = "main"
		}
	}

	return m, tea.Batch(cmds...)
	// return m, cmd
}

// View
func (m formModel) View() string {
	if m.viewState == "main" {
		return m.mainView()
	}
	if m.viewState == "form" {
		return m.newHabitForm.View()
	}
	return m.mainView()
}

func (m formModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		80,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m formModel) mainView() string {
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
			// fmt.Fprintf(&b, "%d. %s the %s (Level %s)\n", i+1, name, c.Class, c.Level)
			// b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	help := s.Help.Render("a: Add new habit  |  q: Quit")
	b.WriteString(help)
	return s.Base.Render(b.String())
}