package view

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func createHabitForm() (*huh.Form, *habitFormFields) {
	fields := newHabitFormFields()
	return buildHabitForm(fields, "Create Habit?"), fields
}

func updateCreateHabit(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
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
	if m.form.State != huh.StateCompleted {
		return m, tea.Batch(cmds...)
	}
	if m.formFields == nil || !m.formFields.Confirm {
		return returnToMain(m), nil
	}

	name := strings.TrimSpace(m.formFields.Name)
	if name == "" {
		name = "Unnamed Habit"
	}
	goalInt, err := strconv.Atoi(m.formFields.GoalString)
	if err != nil || goalInt < 1 {
		log.Printf("Error converting goal to int: %v", err)
		m.statusMsg = "Could not create habit: goal must be a number ≥ 1"
		return returnToMain(m), nil
	}
	color := m.formFields.Color
	if color == "" {
		color = "purple"
	}
	habit := models.Habit{
		Name:        name,
		Description: m.formFields.Description,
		Frequency:   formFrequency(m.formFields),
		Goal:        goalInt,
		Color:       color,
		Icon:        m.formFields.Icon,
		StartDate:   time.Now().Format(time.RFC3339),
	}

	h, err := m.store.CreateHabit(context.Background(), &habit)
	if err != nil {
		log.Printf("Error creating habit: %s", err)
		m.statusMsg = "Could not create habit"
		return returnToMain(m), nil
	}
	m.habits = append(m.habits, *h)
	m.statusMsg = ""
	return returnToMain(m), nil
}

func viewCreateHabit(m Model) string {
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
