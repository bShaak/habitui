package cli

import (
	"os"

	"github.com/bShaak/habitui/internal/api"
)

func openService(dbPath string) (api.Service, error) {
	return api.Open(resolveDBPath(dbPath))
}

func resolveDBPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}
	return os.Getenv("HABITUI_DB")
}
