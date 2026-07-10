package view

import (
	"strings"
)

var allWeekdays = []string{
	"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
}

func effectiveGoal(goal int) int {
	if goal < 1 {
		return 1
	}
	return goal
}

// normalizeFrequency stores empty or all-days schedules as "daily".
func normalizeFrequency(days []string) string {
	cleaned := make([]string, 0, len(days))
	seen := make(map[string]bool, len(days))
	for _, d := range days {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || d == "daily" || seen[d] {
			continue
		}
		seen[d] = true
		cleaned = append(cleaned, d)
	}
	if len(cleaned) == 0 || len(cleaned) == 7 {
		return "daily"
	}
	return strings.Join(cleaned, ",")
}

// frequencyDaysForForm expands "daily" into all weekdays so the multi-select shows a full schedule.
func frequencyDaysForForm(frequency string) []string {
	freq := strings.ToLower(strings.TrimSpace(frequency))
	if freq == "" || freq == "daily" {
		out := make([]string, len(allWeekdays))
		copy(out, allWeekdays)
		return out
	}
	parts := strings.Split(freq, ",")
	cleaned := make([]string, 0, len(parts))
	valid := make(map[string]bool, len(allWeekdays))
	for _, d := range allWeekdays {
		valid[d] = true
	}
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if !valid[p] || seen[p] {
			continue
		}
		seen[p] = true
		cleaned = append(cleaned, p)
	}
	if len(cleaned) == 0 {
		out := make([]string, len(allWeekdays))
		copy(out, allWeekdays)
		return out
	}
	return cleaned
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}
