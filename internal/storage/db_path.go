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
		return validateDBPath(envPath)
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)
	if isGoRunBuildPath(execDir) {
		dbPath := filepath.Join(".", defaultDBFileName)
		return validateDBPath(dbPath)
	}

	dbPath := filepath.Join(execDir, defaultDBFileName)
	return validateDBPath(dbPath)
}

func validateDBPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := os.Lstat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return absPath, nil
		}
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve symlink: %w", err)
		}

		if isUnsafeDBLocation(resolved) {
			return "", fmt.Errorf("database path resolves to unsafe location: %s", resolved)
		}

		return resolved, nil
	}

	return absPath, nil
}

func isUnsafeDBLocation(path string) bool {
	unsafePrefixes := []string{"/etc/", "/proc/", "/sys/", "/dev/", "/tmp/", "/var/tmp/"}
	for _, prefix := range unsafePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
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
