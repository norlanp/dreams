package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"dreams/internal/export"
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
	var exportDir string
	flag.StringVar(&exportDir, "export", "", "Export all dreams to the specified directory as markdown files")
	flag.Parse()

	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	repo, err := storage.NewRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer repo.Close()

	if exportDir != "" {
		dreams, err := repo.ListDreams(context.Background())
		if err != nil {
			return fmt.Errorf("failed to fetch dreams: %w", err)
		}

		count, err := export.ExportAll(dreams, exportDir)
		if err != nil {
			return fmt.Errorf("failed to export dreams: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Exported %d dreams to %s\n", count, exportDir)
		return nil
	}

	return tui.Run(repo)
}
