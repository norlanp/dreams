package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

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
	if err := loadEnvFile(".env"); err != nil {
		return err
	}

	logFile, err := configureLogging("./var/dreams.log")
	if err != nil {
		return err
	}
	defer logFile.Close()

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

	if err := repo.SeedPrimingContent(context.Background()); err != nil {
		log.Printf("warning: failed to seed priming content: %v", err)
	}

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

	return tui.Run(repo, dbPath)
}

func configureLogging(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	info, err := os.Stat(path)
	if err == nil && info.Size() > 10*1024*1024 {
		backup := path + ".old"
		_ = os.Remove(backup)
		if err := os.Rename(path, backup); err != nil {
			// Rotation failed, continue with existing file
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags)
	return file, nil
}

func loadEnvFile(path string) error {
	err := godotenv.Load(path)
	if err == nil || errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("failed to load env file %s: %w", path, err)
}
