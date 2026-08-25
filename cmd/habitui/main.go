package main

import (
	"fmt"
	"os"

	"github.com/bShaak/habitui/internal/cli"
	"github.com/bShaak/habitui/internal/server"
	"github.com/bShaak/habitui/internal/view"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "cli" {
		if err := cli.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "serve" {
		if err := server.Run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	m := view.InitViewState()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithReportFocus())
	finalModel, err := p.Run()
	if fm, ok := finalModel.(view.Model); ok {
		_ = fm.Close()
	}
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
