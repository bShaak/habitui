package view

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

var (
	Name        string
	Frequency   []string
	GoalString  string
	StartDate   string
	Description string
	Color       string
	Icon        string
	Confirm     bool
)

func resetCreateHabitFields() {
	Name = ""
	Frequency = nil
	GoalString = ""
	StartDate = ""
	Description = ""
	Color = "purple"
	Icon = ""
	Confirm = false
}

func CreateHabit() *huh.Form {
	resetCreateHabitFields()
	return habitForm(&Name, &GoalString, &Description, &Color, &Icon, &Frequency, &Confirm, "Create Habit?")
}

func GetCreateHabitUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	// huh expands every group to the tallest group's height on WindowSizeMsg;
	// handle width ourselves and skip forwarding so the box stays content-sized.
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		applyFormSize(m.form, m.width, m.height)
		return m, nil
	}

	var cmds []tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateCompleted {
		if !Confirm {
			m.form = nil
			m.getView = GetMainView
			m.getUpdate = GetMainUpdate
			return m, nil
		}
		name := strings.TrimSpace(m.form.GetString("name"))
		if name == "" {
			name = "Unnamed Habit"
		}
		description := m.form.GetString("description")
		freq := m.form.Get("frequency").([]string)
		frequency := strings.Join(freq, ",")
		if frequency == "" {
			frequency = "Daily"
		}
		goal := m.form.GetString("goal")
		goalInt, err := strconv.Atoi(goal)
		if err != nil {
			log.Fatalf("Error converting goal to int: %s", err)
		}
		color := m.form.GetString("color")
		if color == "" {
			color = "purple"
		}
		icon := m.form.GetString("icon")
		habit := types.Habit{
			Name:        name,
			Description: description,
			Frequency:   frequency,
			Goal:        goalInt,
			Color:       color,
			Icon:        icon,
			StartDate:   time.Now().Format(time.RFC3339),
		}

		h, err := m.store.CreateHabit(context.Background(), &habit)
		if err != nil {
			log.Fatalf("Error creating habit: %s", err)
		}
		m.habits = append(m.habits, *h)

		m.form = nil
		m.getView = GetMainView
		m.getUpdate = GetMainUpdate
	}

	return m, tea.Batch(cmds...)
}

func GetCreateHabitView(m model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	header := m.appBoundaryView("Create New Habit")
	b.WriteString(header)
	b.WriteString("\n\n")
	var content strings.Builder
	content.WriteString(m.form.View())
	content.WriteString("\n\n")
	help := s.Help.Render("tab: next  |  shift+tab: back  |  esc: cancel")
	content.WriteString(help)
	b.WriteString(s.ContentBox.Render(content.String()))
	return s.Base.Render(b.String())
}
