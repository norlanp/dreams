package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dreams/internal/model"
)

const maxDreamContentLength = 100000 // 100KB limit

func ExportAll(dreams []model.Dream, directory string) (int, error) {
	if len(dreams) == 0 {
		return 0, nil
	}

	// Path traversal protection: ensure directory is within reasonable bounds
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return 0, fmt.Errorf("invalid directory path: %w", err)
	}

	if containsPathTraversal(absDir) {
		return 0, fmt.Errorf("export directory contains path traversal sequences")
	}

	directory = absDir

	if err := os.MkdirAll(directory, 0700); err != nil {
		return 0, fmt.Errorf("failed to create directory: %w", err)
	}

	exported := 0
	var lastErr error

	for _, dream := range dreams {
		if err := exportDream(dream, directory); err != nil {
			lastErr = err
			continue
		}
		exported++
	}

	return exported, lastErr
}

func exportDream(dream model.Dream, directory string) error {
	filename := generateFilename(dream)
	filepath := filepath.Join(directory, filename)
	content := generateContent(dream)

	tmpFile, err := os.CreateTemp(directory, ".export-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write content: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, filepath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func generateFilename(dream model.Dream) string {
	return fmt.Sprintf("%s-%d-dream.md",
		dream.CreatedAt.Format("2006-01-02-15-04-05"),
		dream.ID)
}

func generateContent(dream model.Dream) string {
	return fmt.Sprintf("---\ndate: %s\n---\n\n%s",
		dream.CreatedAt.Format("2006-01-02"),
		dream.Content)
}

func containsPathTraversal(path string) bool {
	// Check for common path traversal patterns
	return strings.Contains(path, "..") || strings.Contains(path, "~") ||
		strings.HasPrefix(path, "/etc") || strings.HasPrefix(path, "/proc") ||
		strings.HasPrefix(path, "/sys") || strings.HasPrefix(path, "/dev")
}
