package main

import (
	"fmt"
	"os"

	"github.com/bShaak/habitui/internal/view"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := view.InitViewState()
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if fm, ok := finalModel.(view.Model); ok {
		_ = fm.Close()
	}
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
