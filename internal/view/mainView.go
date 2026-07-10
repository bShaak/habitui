package view

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	types "github.com/bShaak/habitui/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func IsCompleted(completions []types.Completion, h *types.Habit) bool {
	completionCount := 0
	for _, c := range completions {
		if c.HabitID == h.ID {
			completionCount++
		}
	}
	goal := h.Goal
	if goal < 1 {
		goal = 1
	}
	return completionCount >= goal
}

func todayCompletionCount(completions []types.Completion, habitID int64) int {
	count := 0
	for _, c := range completions {
		if c.HabitID == habitID {
			count++
		}
	}
	return count
}

func getHabitColor(color string) lipgloss.Color {
	return GetHabitColor(color)
}

// Update
func GetMainUpdate(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmingDelete {
			switch msg.String() {
			case "y", "enter":
				return deleteSelectedHabit(m), nil
			case "n", "x":
				m.confirmingDelete = false
				return m, nil
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "q":
			defer m.store.Close()
			return m, tea.Quit
		case "a":
			m.form = CreateHabit()
			applyFormSize(m.form, m.width, m.height)
			m.scrollOffset = 0
			m.getView = GetCreateHabitView
			m.getUpdate = GetCreateHabitUpdate
			return m, m.form.Init()
		case "c":
			habits, err := m.store.ListHabits(context.Background())
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
			}
			m.habits = habits
			weekEnd := m.weekStart.AddDate(0, 0, 6)
			completions, err := m.store.GetCompletionsByDateRange(context.Background(), m.weekStart, weekEnd)
			if err != nil {
				log.Printf("Error fetching week completions: %s", err)
			}
			m.weekCompletions = completions
			if m.cursor >= len(m.habits) {
				m.cursor = 0
			}
			m.calendarCol = 0
			m.scrollOffset = 0
			m.getView = GetCalendarView
			m.getUpdate = GetCalendarUpdate
			return m, nil
		case "s":
			habits, err := m.store.ListHabits(context.Background())
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
			}
			m.habits = habits
			if m.cursor >= len(m.habits) {
				m.cursor = 0
			}
			m.statsTab = 0
			m.scrollOffset = 0
			m.getView = GetStatsView
			m.getUpdate = GetStatsUpdate
			return m, nil
		case "e":
			if len(m.habits) == 0 {
				return m, nil
			}
			m.form = EditHabit(m.habits[m.cursor])
			applyFormSize(m.form, m.width, m.height)
			m.scrollOffset = 0
			m.getView = GetEditHabitView
			m.getUpdate = GetEditHabitUpdate
			return m, m.form.Init()
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "j":
			if m.cursor < len(m.habits)-1 {
				m.cursor++
			}
		case "x":
			if len(m.habits) > 0 {
				m.confirmingDelete = true
			}
		case "enter":
			if len(m.habits) == 0 {
				return m, nil
			}
			if !IsCompleted(m.completions, &m.habits[m.cursor]) {
				c, err := m.store.CreateCompletion(context.Background(), &types.Completion{HabitID: m.habits[m.cursor].ID, CompletedAt: time.Now().Format(time.RFC3339)})
				if err != nil {
					log.Printf("Error creating completion: %s", err)
					return m, nil
				}
				m.completions = append(m.completions, *c)
			} else {
				completions, err := m.store.GetCompletionsByHabitId(context.Background(), m.habits[m.cursor].ID)
				if err != nil {
					log.Printf("Error retrieving completions: %s", err)
					return m, nil
				}
				now := time.Now()
				dayStart := startOfDay(now)
				dayEnd := endOfDay(now)

				for _, c := range completions {
					completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
					if err != nil {
						continue
					}
					if (completedAt.Equal(dayStart) || completedAt.After(dayStart)) &&
						(completedAt.Equal(dayEnd) || completedAt.Before(dayEnd)) {
						err := m.store.DeleteCompletion(context.Background(), c.ID)
						if err != nil {
							log.Printf("Error deleting completion: %s", err)
						}
					}
				}
				var updated []types.Completion
				for _, c := range m.completions {
					completedAt, err := time.Parse(time.RFC3339, c.CompletedAt)
					if err != nil {
						updated = append(updated, c)
						continue
					}
					if !(c.HabitID == m.habits[m.cursor].ID &&
						(completedAt.Equal(dayStart) || completedAt.After(dayStart)) &&
						(completedAt.Equal(dayEnd) || completedAt.Before(dayEnd))) {
						updated = append(updated, c)
					}
				}
				m.completions = updated
			}
			m = refreshStreakCompletions(m)
		}
	}
	return m, nil
}

func deleteSelectedHabit(m model) model {
	if len(m.habits) == 0 {
		m.confirmingDelete = false
		return m
	}
	err := m.store.DeleteHabit(context.Background(), m.habits[m.cursor].ID)
	if err != nil {
		log.Printf("Error deleting habit: %s", err)
		m.confirmingDelete = false
		return m
	}
	if m.cursor == len(m.habits)-1 {
		m.habits = m.habits[:m.cursor]
	} else {
		m.habits = append(m.habits[:m.cursor], m.habits[m.cursor+1:]...)
	}
	if m.cursor > 0 && m.cursor >= len(m.habits) {
		m.cursor--
	}
	m.confirmingDelete = false
	return m
}

// View
func GetMainView(m model) string {
	s := m.styles
	var b strings.Builder
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	header := m.appBoundaryView("Today's Habits")
	b.WriteString(header)
	b.WriteString("\n\n")
	var content strings.Builder
	if len(m.habits) == 0 {
		content.WriteString(s.Help.Render("No habits created yet.\n\nPress 'a' to create a new one."))
	} else {
		for i, h := range m.habits {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			habitColor := getHabitColor(h.Color)
			completed := ""
			if IsCompleted(m.completions, &h) {
				completed = "✓"
			} else {
				completionCount := todayCompletionCount(m.completions, h.ID)
				completed = fmt.Sprintf("✗ (%d/%d)", completionCount, h.Goal)
			}
			name := formatHabitLabel(h)
			if name == "" || strings.TrimSpace(h.Name) == "" {
				name = "Unnamed"
				if h.Icon != "" {
					name = h.Icon + " Unnamed"
				}
			}
			currentStreak, _ := getHabitStreak(h, m.streakCompletions, time.Now())
			streakText := ""
			if currentStreak >= 3 {
				streakText = fmt.Sprintf(" 🔥 %d", currentStreak)
			}
			nameStyle := lipgloss.NewStyle().Foreground(habitColor)
			completedStyle := lipgloss.NewStyle().Foreground(habitColor)
			streakStyle := lipgloss.NewStyle().Foreground(peach)
			content.WriteString(fmt.Sprintf("%s %s %s%s\n",
				cursor,
				nameStyle.Render(name),
				completedStyle.Render(completed),
				streakStyle.Render(streakText),
			))
		}
		content.WriteString("\n")
		if m.confirmingDelete {
			habitName := formatHabitLabel(m.habits[m.cursor])
			confirm := lipgloss.NewStyle().Foreground(red).Bold(true).Render(
				fmt.Sprintf("Delete %q?  y: confirm  |  n/esc: cancel", habitName),
			)
			content.WriteString(confirm)
		} else {
			help := s.Help.Render("a: Add  |  c: Calendar  |  s: Stats  |  e: Edit  |  x: Delete  |  enter: Toggle  |  q: Quit")
			content.WriteString(help)
		}
	}
	b.WriteString(s.ContentBox.Render(content.String()))
	return s.Base.Render(b.String())
}
