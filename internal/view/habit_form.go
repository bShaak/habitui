package view

import (
	"errors"
	"strconv"

	"github.com/bShaak/habitui/internal/models"
	"github.com/charmbracelet/huh"
)

const formSelectHeight = 6

type habitFormFields struct {
	Name        string
	Frequency   []string
	GoalString  string
	Description string
	Color       string
	Icon        string
	Confirm     bool
}

func newHabitFormFields() *habitFormFields {
	return &habitFormFields{
		Color: "purple",
	}
}

func habitFormFieldsFromHabit(habit models.Habit) *habitFormFields {
	color := habit.Color
	if color == "" {
		color = "purple"
	}
	return &habitFormFields{
		Name:        habit.Name,
		GoalString:  strconv.Itoa(effectiveGoal(habit.Goal)),
		Description: habit.Description,
		Frequency:   frequencyDaysForForm(habit.Frequency),
		Color:       color,
		Icon:        habit.Icon,
		Confirm:     false,
	}
}

func buildHabitForm(fields *habitFormFields, confirmTitle string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Name").
				Key("name").
				Value(&fields.Name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name must not be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Times per day").
				Key("goal").
				Value(&fields.GoalString).
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
				Title("Description").
				Key("description").
				CharLimit(400).
				Lines(2).
				Value(&fields.Description),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Schedule").
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
				Height(formSelectHeight).
				Value(&fields.Frequency),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Color").
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
				Height(formSelectHeight).
				Value(&fields.Color),
			huh.NewSelect[string]().
				Title("Icon").
				Key("icon").
				Options(habitIconOptions()...).
				Height(formSelectHeight).
				Value(&fields.Icon),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title(confirmTitle).
				Key("confirm").
				Affirmative("Yes").
				Negative("No").
				Value(&fields.Confirm),
		),
	).WithWidth(60).WithTheme(huh.ThemeCatppuccin())
}

func applyFormSize(form *huh.Form, width, height int) {
	if form == nil {
		return
	}
	if width > 0 {
		formWidth := width - 8
		if formWidth < 40 {
			formWidth = 40
		}
		if formWidth > 60 {
			formWidth = 60
		}
		form.WithWidth(formWidth)
	}
	// Intentionally skip WithHeight so the form (and outer box) shrink to content.
	// Field Heights + multi-page groups keep long lists scrollable without empty space.
	_ = height
}

func formFrequency(fields *habitFormFields) string {
	if fields == nil {
		return "daily"
	}
	return normalizeFrequency(fields.Frequency)
}
