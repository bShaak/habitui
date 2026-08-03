package cli

import "github.com/bShaak/habitui/internal/models"

type listResponse struct {
	Date   string         `json:"date"`
	Habits []habitSummary `json:"habits"`
}

type habitSummary struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Frequency       string `json:"frequency"`
	Goal            int    `json:"goal"`
	Color           string `json:"color,omitempty"`
	Icon            string `json:"icon,omitempty"`
	Due             bool   `json:"due"`
	Complete        bool   `json:"complete"`
	CompletionCount int    `json:"completion_count"`
}

type completeResponse struct {
	HabitID          int64              `json:"habit_id"`
	Date             string             `json:"date"`
	AlreadyComplete  bool               `json:"already_complete"`
	Completion       *models.Completion `json:"completion,omitempty"`
	CompletionCount  int                `json:"completion_count"`
	Goal             int                `json:"goal"`
}
