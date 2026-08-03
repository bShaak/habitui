package cli

import (
	"os"

	"github.com/bShaak/habitui/internal/storage"
)

func openStore(dbPath string) (storage.Store, error) {
	if dbPath == "" {
		return storage.OpenSQLite()
	}
	return storage.OpenSQLiteAt(dbPath)
}

func resolveDBPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}
	return os.Getenv("HABITUI_DB")
}
