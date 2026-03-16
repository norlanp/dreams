package main

import (
	"fmt"
	"os"

	"dreams/internal/storage"
	"dreams/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	repo, err := storage.NewRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer repo.Close()

	return tui.Run(repo)
}
