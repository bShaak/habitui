package ui

import (
	"github.com/charmbracelet/huh"
)

// type CreateHabitModel struct {
// 	Name string
// 	Description string
// 	Frequency string
// 	GoalString string
// 	StartDate string
// }

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
					Value(&Name),
					// Validate(func(str string) error {
					// 		if str == "" {
					// 			return errors.New("name must not be empty")
					// 		}
					// 		return nil
					// }),

					huh.NewInput().
					Title("How many times per day do you want to track this habit?").
					Key("goal").
					Value(&GoalString),
					// Validating fields is easy. The form will mark erroneous fields
					// and display error messages accordingly.
					// Validate(func(str string) error {
					// 		// TODO: Validate that the goal is a number
					// 		if _, err := strconv.Atoi(str); err != nil {
					// 			return errors.New("goal must be a number")
					// 		}
					// 		// if Goal < 1 {
					// 		// 	return errors.New("Goal must be at least 1")
					// 		// }
					// 		return nil
					// }),
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
			// huh.NewInput().
			// 	Title("Start Date (YYYY-MM-DD)").
			// 	Key("startDate").
			// 	Value(&StartDate).
			// 	Placeholder(time.Now().Format("2006-01-02")),

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