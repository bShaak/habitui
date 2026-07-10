package view

import "github.com/charmbracelet/huh"

func habitIconOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("None", ""),
		huh.NewOption("🏃 Running", "🏃"),
		huh.NewOption("💪 Strength", "💪"),
		huh.NewOption("🧘 Meditation", "🧘"),
		huh.NewOption("📚 Reading", "📚"),
		huh.NewOption("✍️ Writing", "✍️"),
		huh.NewOption("💧 Water", "💧"),
		huh.NewOption("🥗 Nutrition", "🥗"),
		huh.NewOption("😴 Sleep", "😴"),
		huh.NewOption("🎯 Focus", "🎯"),
		huh.NewOption("💻 Coding", "💻"),
		huh.NewOption("🎵 Music", "🎵"),
		huh.NewOption("🧹 Chores", "🧹"),
		huh.NewOption("🚶 Walk", "🚶"),
		huh.NewOption("🧠 Study", "🧠"),
		huh.NewOption("❤️ Health", "❤️"),
	}
}
