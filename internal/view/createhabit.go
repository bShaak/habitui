package view

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

var (
	Name string
	Frequency []string
	GoalString string
	StartDate string
	Description string
	Confirm bool
)

func CreateHabit() *huh.Form {
	form := huh.NewForm(
	huh.NewGroup(
			huh.NewInput().
					Title("Habit Name").
					Key("name").
					Value(&Name).
					Validate(func(str string) error {
							if str == "" {
								return errors.New("name must not be empty")
							}
							return nil
					}),

					huh.NewInput().
					Title("How many times per day do you want to track this habit?").
					Key("goal").
					Value(&GoalString).
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
					Value(&Description),
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
					Value(&Frequency),

			huh.NewConfirm().
					Title("Create Habit?").
					Key("confirm").
					Affirmative("Yes").
					Negative("No").
					Value(&Confirm),
	),
)

	return form
}

func GetCreateHabitUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// frequency := m.newHabitForm.GetString("Frequency")
		freq := m.form.Get("frequency").([]string)
		frequency := strings.Join(freq, ", ")
		if frequency == "" {
				frequency = "Daily"  // Default
		}
		fmt.Printf("Frequency: %s\n", frequency)
		goal := m.form.GetString("goal")
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
		
		h, err := m.store.CreateHabit(context.Background(), &habit)
		m.habits = append(m.habits, *h)
		if err != nil {
			log.Fatalf("Error creating habit: %s", err)
		}

		// Return to main view after creating habit
		m.getView = GetMainView
		m.getUpdate = GetMainUpdate
	}

	return m, tea.Batch(cmds...)
}

func GetCreateHabitView(m model) string {
	return m.form.View()
}