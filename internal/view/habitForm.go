package view

import (
	"errors"
	"strconv"

	"github.com/charmbracelet/huh"
)

const (
	formSelectHeight = 6
)

func habitForm(name, goal, description, color, icon *string, frequency *[]string, confirm *bool, confirmTitle string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Habit Name").
				Key("name").
				Value(name).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("name must not be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Times per day").
				Key("goal").
				Value(goal).
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
				Value(description),
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
				Value(frequency),
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
				Value(color),
			huh.NewSelect[string]().
				Title("Icon").
				Key("icon").
				Options(habitIconOptions()...).
				Height(formSelectHeight).
				Value(icon),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title(confirmTitle).
				Key("confirm").
				Affirmative("Yes").
				Negative("No").
				Value(confirm),
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
