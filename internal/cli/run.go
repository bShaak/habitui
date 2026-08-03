package cli

import (
	"fmt"
	"os"
)

// Run executes habitui cli subcommands. Pass os.Args[2:] when invoked as `habitui cli ...`.
func Run(args []string) error {
	if len(args) == 0 {
		usage(os.Stderr)
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "list":
		return runList(args[1:])
	case "complete":
		return runComplete(args[1:])
	case "-h", "--help", "help":
		usage(os.Stdout)
		return nil
	default:
		usage(os.Stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(w interface{ Write([]byte) (int, error) }) {
	fmt.Fprintln(w, "usage: habitui cli <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  list      list habits with due/complete state")
	fmt.Fprintln(w, "  complete  record a completion for a habit on a date")
}
