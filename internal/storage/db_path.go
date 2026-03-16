package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDBFileName = "dreams.db"
	defaultDBPathEnv  = "DREAMS_DB_PATH"
)

func DefaultDBPath() (string, error) {
	envPath := strings.TrimSpace(os.Getenv(defaultDBPathEnv))
	if envPath != "" {
		return envPath, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	if isGoRunBuildPath(execDir) {
		return filepath.Join(".", defaultDBFileName), nil
	}

	return filepath.Join(execDir, defaultDBFileName), nil
}

func isGoRunBuildPath(path string) bool {
	tmpDir := os.TempDir()
	rel, err := filepath.Rel(tmpDir, path)
	if err != nil {
		return false
	}

	if rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}

	return strings.Contains(rel, "go-build")
}
