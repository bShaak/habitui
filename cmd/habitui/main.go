package main

import (
	"fmt"
	"os"

	"github.com/bShaak/habitui/internal/view"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(view.InitViewState())
	_, err := p.Run();
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}