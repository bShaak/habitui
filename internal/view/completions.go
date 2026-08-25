package view

import (
	"context"
	"time"

	"github.com/bShaak/habitui/internal/models"
)

// toggleDayCompletion toggles a completion for habit on day and returns the
// updated cached completion list.
func (m Model) toggleDayCompletion(habit models.Habit, day time.Time, list []models.Completion) ([]models.Completion, error) {
	res, err := m.svc.ToggleCompletionForDate(context.Background(), habit.ID, day)
	if err != nil {
		return list, err
	}

	if len(res.RemovedIDs) > 0 {
		removed := make(map[int64]bool, len(res.RemovedIDs))
		for _, id := range res.RemovedIDs {
			removed[id] = true
		}
		var updated []models.Completion
		for _, c := range list {
			if removed[c.ID] {
				continue
			}
			updated = append(updated, c)
		}
		return updated, nil
	}

	if res.Added != nil {
		return append(list, *res.Added), nil
	}
	return list, nil
}
