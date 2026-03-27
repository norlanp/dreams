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

func TestExportAll_RejectsSystemPaths(t *testing.T) {
	dreams := []model.Dream{
		{ID: 1, Content: "test", CreatedAt: time.Now()},
	}

	unsafePaths := []struct {
		name string
		path string
	}{
		{"etc", "/etc"},
		{"proc", "/proc"},
		{"sys", "/sys"},
		{"dev", "/dev"},
		{"tmp", "/tmp"},
		{"var_tmp", "/var/tmp"},
		{"etc_subdir", "/etc/app"},
		{"proc_subdir", "/proc/self"},
	}

	for _, tc := range unsafePaths {
		t.Run(tc.name, func(t *testing.T) {
			_, err := export.ExportAll(dreams, tc.path)
			if err == nil {
				t.Errorf("expected error for path %s, got nil", tc.path)
			}
			if err != nil && !strings.Contains(err.Error(), "working directory") {
				t.Errorf("expected working directory error, got: %v", err)
			}
		})
	}
}

func TestExportAll_RejectsParentDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	parentDir := filepath.Dir(cwd)
	if parentDir == cwd {
		t.Skip("cannot test parent: already at root")
	}

	dreams := []model.Dream{
		{ID: 1, Content: "test", CreatedAt: time.Now()},
	}

	_, err = export.ExportAll(dreams, parentDir)
	if err == nil {
		t.Errorf("expected error for parent directory, got nil")
	}
}

func TestExportAll_AllowsSubdirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	subDir := filepath.Join(cwd, "test-export-subdir")
	defer os.RemoveAll(subDir)

	dreams := []model.Dream{
		{ID: 1, Content: "test dream content", CreatedAt: time.Now()},
	}

	count, err := export.ExportAll(dreams, subDir)
	if err != nil {
		t.Fatalf("expected success for subdirectory, got error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 dream exported, got %d", count)
	}
}

func TestExportAll_AllowsCurrentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(cwd)

	dreams := []model.Dream{
		{ID: 1, Content: "test", CreatedAt: time.Now()},
	}

	count, err := export.ExportAll(dreams, ".")
	if err != nil {
		t.Fatalf("expected success for current directory, got error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 dream exported, got %d", count)
	}
}

func TestExportAll_NormalizesPathBeforeCheck(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	subDir := filepath.Join(cwd, "test-norm-export")
	defer os.RemoveAll(subDir)

	dreams := []model.Dream{
		{ID: 1, Content: "test", CreatedAt: time.Now()},
	}

	relativePath := "./test-norm-export"
	count, err := export.ExportAll(dreams, relativePath)
	if err != nil {
		t.Fatalf("expected success for relative subdirectory, got error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 dream exported, got %d", count)
	}
}

func TestExportAll_EmptyDreamsNoOp(t *testing.T) {
	count, err := export.ExportAll([]model.Dream{}, "/tmp")
	if err != nil {
		t.Fatalf("unexpected error for empty dreams: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for empty dreams, got %d", count)
	}
}
