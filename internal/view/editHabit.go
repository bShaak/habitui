package view

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func EditHabit(habit types.Habit) *huh.Form {
	goalString := strconv.Itoa(habit.Goal)
	frequency := strings.Split(habit.Frequency, ",")
	color := habit.Color
	if color == "" {
		color = "purple"
	}

	for i, f := range frequency {
		frequency[i] = strings.TrimSpace(f)
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Name").
				Key("name").
				Value(&habit.Name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name must not be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("How many times per day do you want to track this habit?").
				Key("goal").
				Value(&goalString).
				Validate(func(str string) error {
					goalInt, err := strconv.Atoi(str)
					if err != nil {
						return errors.New("goal must be a number")
					}
					if goalInt < 1 {
						return errors.New("goal must be at least 1")
					}
					return nil
				}),
			huh.NewText().
				Title("Habit Description").
				Key("description").
				CharLimit(400).
				Value(&habit.Description),
			huh.NewMultiSelect[string]().
				Title("What days of the week do you want to track this habit?").
				Key("frequency").
				Options(
					huh.NewOption("Monday", "monday"),
					huh.NewOption("Tuesday", "tuesday"),
					huh.NewOption("Wednesday", "wednesday"),
					huh.NewOption("Thursday", "thursday"),
					huh.NewOption("Friday", "friday"),
					huh.NewOption("Saturday", "saturday"),
					huh.NewOption("Sunday", "sunday"),
				).
				Value(&frequency),
			huh.NewSelect[string]().
				Title("What color should this habit be?").
				Key("color").
				Options(
					huh.NewOption("Red", "red"),
					huh.NewOption("Blue", "blue"),
					huh.NewOption("Green", "green"),
					huh.NewOption("Yellow", "yellow"),
					huh.NewOption("Orange", "orange"),
					huh.NewOption("Purple", "purple"),
					huh.NewOption("Pink", "pink"),
				).
				Value(&color),

			huh.NewConfirm().
				Title("Update Habit?").
				Key("confirm").
				Affirmative("Yes").
				Negative("No").
				Value(&Confirm),
		),
	).WithWidth(60).WithTheme(huh.ThemeCatppuccin())

	return form
}

func GetEditHabitUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	habit := m.habits[m.cursor]
	var cmds []tea.Cmd
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}
	if m.form.State == huh.StateCompleted {
		if !Confirm {
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
			frequency = "monday,tuesday,wednesday,thursday,friday,saturday,sunday"
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
		habit.Name = name
		habit.Description = description
		habit.Goal = goalInt
		habit.Frequency = frequency
		habit.Color = color
		err = m.store.UpdateHabit(context.Background(), &habit)
		if err != nil {
			log.Fatalf("Error updating habit: %s", err)
		}
		for i, h := range m.habits {
			if h.ID == habit.ID {
				m.habits[i] = habit
				break
			}
		}

		m.getView = GetMainView
		m.getUpdate = GetMainUpdate
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
	help := s.Help.Render("esc: cancel  |  enter: confirm")
	content.WriteString(help)
	b.WriteString(s.ContentBox.Render(content.String()))
	return s.Base.Render(b.String())
}
