package view

import (
	"context"
	"log"
	"time"

	"github.com/bShaak/habitui/internal/models"
)

// toggleDayCompletion adds a completion for habit on day, or removes all completions that day if the goal is already met.
func (m Model) toggleDayCompletion(habit models.Habit, day time.Time, list []models.Completion) ([]models.Completion, error) {
	goal := effectiveGoal(habit.Goal)
	count := getCompletionsForHabitAndDate(list, habit.ID, day)

	if count >= goal {
		completions, err := m.store.GetCompletionsByHabitID(context.Background(), habit.ID)
		if err != nil {
			return list, err
		}
		for _, c := range completions {
			completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
			if err != nil {
				continue
			}
			if !inDayBounds(completedAt, day) {
				continue
			}
			if err := m.store.DeleteCompletion(context.Background(), c.ID); err != nil {
				log.Printf("Error deleting completion: %s", err)
			}
		}
		var updated []models.Completion
		for _, c := range list {
			completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
			if err != nil {
				updated = append(updated, c)
				continue
			}
			if c.HabitID == habit.ID && inDayBounds(completedAt, day) {
				continue
			}
			updated = append(updated, c)
		}
		return updated, nil
	}

	now := time.Now()
	completedAt := time.Date(
		day.Year(), day.Month(), day.Day(),
		now.Hour(), now.Minute(), now.Second(), 0, now.Location(),
	)
	c, err := m.store.CreateCompletion(context.Background(), &models.Completion{
		HabitID:     habit.ID,
		CompletedAt: completedAt.Format(time.RFC3339),
	})
	if err != nil {
		return list, err
	}
	return append(list, *c), nil
}
