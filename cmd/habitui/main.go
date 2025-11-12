package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// p := tea.NewProgram(InitialTextInputModel())
	p := tea.NewProgram(InitialFormModel())
	_, err := p.Run();
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}