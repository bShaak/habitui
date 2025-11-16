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
					goalInt, err := strconv.Atoi(str);
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

			huh.NewConfirm().
				Title("Update Habit?").
				Key("confirm").
				Affirmative("Yes").
				Negative("No").
				Value(&Confirm),
		),
	)

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
				frequency = "monday,tuesday,wednesday,thursday,friday,saturday,sunday"  // Default frequency is daily
		}
		goal := m.form.GetString("goal")
		goalInt, err := strconv.Atoi(goal)
		if err != nil {
			log.Fatalf("Error converting goal to int: %s", err)
		}
		habit.Name = name
		habit.Description = description
		habit.Goal = goalInt
		habit.Frequency = frequency
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
		

		// Return to main view after creating habit
		m.getView = GetMainView
		m.getUpdate = GetMainUpdate
	}

	return m, tea.Batch(cmds...)
}

func GetEditHabitView(m model) string {
	return m.form.View()
}