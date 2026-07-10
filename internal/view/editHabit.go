package view

import (
	"context"
	"log"
	"strconv"
	"strings"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func EditHabit(habit types.Habit) *huh.Form {
	goalString := strconv.Itoa(effectiveGoal(habit.Goal))
	color := habit.Color
	if color == "" {
		color = "purple"
	}

	// Bind to package-level vars so values survive across form pages.
	Name = habit.Name
	GoalString = goalString
	Description = habit.Description
	Frequency = frequencyDaysForForm(habit.Frequency)
	Color = color
	Icon = habit.Icon
	Confirm = false

	return habitForm(&Name, &GoalString, &Description, &Color, &Icon, &Frequency, &Confirm, "Update Habit?")
}

func GetEditHabitUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
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
	if m.form.State == huh.StateCompleted {
		if !Confirm {
			return returnToMain(m), nil
		}
		name := strings.TrimSpace(m.form.GetString("name"))
		if name == "" {
			name = "Unnamed Habit"
		}
		description := m.form.GetString("description")
		frequency := formFrequency(m.form)
		goal := m.form.GetString("goal")
		goalInt, err := strconv.Atoi(goal)
		if err != nil || goalInt < 1 {
			log.Printf("Error converting goal to int: %v", err)
			m.statusMsg = "Could not update habit: goal must be a number ≥ 1"
			return returnToMain(m), nil
		}
		color := m.form.GetString("color")
		if color == "" {
			color = "purple"
		}
		icon := m.form.GetString("icon")
		habit.Name = name
		habit.Description = description
		habit.Goal = goalInt
		habit.Frequency = frequency
		habit.Color = color
		habit.Icon = icon
		err = m.store.UpdateHabit(context.Background(), &habit)
		if err != nil {
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

	return m, tea.Batch(cmds...)
}

func GetEditHabitView(m model) string {
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
