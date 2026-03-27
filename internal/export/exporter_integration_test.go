package export_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dreams/internal/export"
	"dreams/internal/model"
)

func setupExportDir(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	exportDir := filepath.Join(cwd, "test-export-"+t.Name())
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("failed to create export dir: %v", err)
	}
	return exportDir
}

func TestExportAll_Integration(t *testing.T) {
	exportDir := setupExportDir(t)
	defer os.RemoveAll(exportDir)

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "First dream content",
			CreatedAt: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Content:   "Second dream with more details",
			CreatedAt: time.Date(2026, 3, 16, 11, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 16, 11, 30, 0, 0, time.UTC),
		},
	}

	count, err := export.ExportAll(dreams, exportDir)
	if err != nil {
		t.Fatalf("ExportAll failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 dreams exported, got %d", count)
	}

	entries, err := os.ReadDir(exportDir)
	if err != nil {
		t.Fatalf("failed to read export directory: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 files in directory, got %d", len(entries))
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".md" {
			t.Errorf("expected markdown file, got %s", entry.Name())
		}

		content, err := os.ReadFile(filepath.Join(exportDir, entry.Name()))
		if err != nil {
			t.Errorf("failed to read exported file %s: %v", entry.Name(), err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("exported file %s is empty", entry.Name())
		}
	}
}

func TestExportAll_EmptyDreams(t *testing.T) {
	exportDir := setupExportDir(t)
	defer os.RemoveAll(exportDir)

	count, err := export.ExportAll([]model.Dream{}, exportDir)
	if err != nil {
		t.Fatalf("ExportAll failed with empty dreams: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 dreams exported, got %d", count)
	}
}

func TestExportAll_InvalidDirectory(t *testing.T) {
	dreams := []model.Dream{
		{ID: 1, Content: "Test", CreatedAt: time.Now()},
	}

	invalidDir := "/root/invalid/path/that/does/not/exist"
	_, err := export.ExportAll(dreams, invalidDir)
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
}

func TestExportAll_PartialFailure(t *testing.T) {
	exportDir := setupExportDir(t)
	defer os.RemoveAll(exportDir)

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "Valid dream",
			CreatedAt: time.Now(),
		},
	}

	count, err := export.ExportAll(dreams, exportDir)
	if err != nil {
		t.Fatalf("ExportAll failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 dream exported, got %d", count)
	}

	files, _ := os.ReadDir(exportDir)
	if len(files) != 1 {
		t.Errorf("expected 1 file created, got %d", len(files))
	}
}

func TestExportFilenameGeneration(t *testing.T) {
	exportDir := setupExportDir(t)
	defer os.RemoveAll(exportDir)

	dream := model.Dream{
		ID:        42,
		Content:   "Test content",
		CreatedAt: time.Date(2026, 1, 15, 14, 30, 45, 0, time.UTC),
	}

	_, err := export.ExportAll([]model.Dream{dream}, exportDir)
	if err != nil {
		t.Fatalf("ExportAll failed: %v", err)
	}

	expectedPrefix := "2026-01-15-14-30-45-42-dream.md"

	entries, _ := os.ReadDir(exportDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	filename := entries[0].Name()
	if filename != expectedPrefix {
		t.Errorf("expected filename %s, got %s", expectedPrefix, filename)
	}
}

func TestExportContentFormat(t *testing.T) {
	exportDir := setupExportDir(t)
	defer os.RemoveAll(exportDir)

	dream := model.Dream{
		ID:        1,
		Content:   "Dream content here",
		CreatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
	}

	_, err := export.ExportAll([]model.Dream{dream}, exportDir)
	if err != nil {
		t.Fatalf("ExportAll failed: %v", err)
	}

	entries, _ := os.ReadDir(exportDir)
	content, err := os.ReadFile(filepath.Join(exportDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	contentStr := string(content)

	expectedElements := []string{
		"---",
		"date: 2026-03-20",
		"---",
		"Dream content here",
	}

	for _, element := range expectedElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("expected content to contain %q", element)
		}
	}
}
