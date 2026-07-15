package view

import (
	"errors"
	"strconv"

	"github.com/bShaak/habitui/internal/models"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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
		Color: "red",
	}
}

func habitFormFieldsFromHabit(habit models.Habit) *habitFormFields {
	color := habit.Color
	if color == "" {
		color = "red"
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
	).WithWidth(60).WithTheme(formTheme())
}

// formTheme builds a huh theme from the active Habitui palette so create/edit
// forms follow theme cycling instead of staying locked to Catppuccin.
func formTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Base = t.Focused.Base.BorderForeground(subtext)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(primary)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(primary)
	t.Focused.Directory = t.Focused.Directory.Foreground(primary)
	t.Focused.Description = t.Focused.Description.Foreground(muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(pink)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(pink)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(pink)
	t.Focused.Option = t.Focused.Option.Foreground(text)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(pink)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(green)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(text)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(text)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(surface).Background(pink)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(text).Background(surface)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(orange)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(pink)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base

	t.Help = help.New().Styles
	t.Help.Ellipsis = t.Help.Ellipsis.Foreground(muted)
	t.Help.ShortKey = t.Help.ShortKey.Foreground(muted)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(subtext)
	t.Help.ShortSeparator = t.Help.ShortSeparator.Foreground(muted)
	t.Help.FullKey = t.Help.FullKey.Foreground(muted)
	t.Help.FullDesc = t.Help.FullDesc.Foreground(subtext)
	t.Help.FullSeparator = t.Help.FullSeparator.Foreground(muted)

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	return t
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
