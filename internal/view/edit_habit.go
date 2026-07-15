package view

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func editHabitForm(habit models.Habit) (*huh.Form, *habitFormFields) {
	fields := habitFormFieldsFromHabit(habit)
	return buildHabitForm(fields, "Update Habit?"), fields
}

func updateEditHabit(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	// huh expands every group to the tallest group's height on WindowSizeMsg;
	// handle width ourselves and skip forwarding so the box stays content-sized.
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		applyFormSize(m.form, m.width, m.height)
		return m, nil
	}

	if len(m.habits) == 0 || m.cursor < 0 || m.cursor >= len(m.habits) {
		return returnToMain(m), nil
	}

	habit := m.habits[m.cursor]
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
		m.statusMsg = "Could not update habit: goal must be a number ≥ 1"
		return returnToMain(m), nil
	}
	color := m.formFields.Color
	if color == "" {
		color = "red"
	}
	habit.Name = name
	habit.Description = m.formFields.Description
	habit.Goal = goalInt
	habit.Frequency = formFrequency(m.formFields)
	habit.Color = color
	habit.Icon = m.formFields.Icon
	if err := m.store.UpdateHabit(context.Background(), &habit); err != nil {
		log.Printf("Error updating habit: %s", err)
		m.statusMsg = "Could not update habit"
		return returnToMain(m), nil
	}
	for i, h := range m.habits {
		if h.ID == habit.ID {
			m.habits[i] = habit
			break
		}
	}
	m.statusMsg = ""
	return returnToMain(m), nil
}

func viewEditHabit(m Model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	header := m.appBoundaryView("Edit Habit")
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
