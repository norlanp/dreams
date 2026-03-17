package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dreams/internal/model"
)

func TestExportAll_ShouldCreateFileWithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "I was flying over mountains",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	filename := "2026-03-15-08-30-00-dream.md"
	path := filepath.Join(tmpDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	expectedFrontmatter := "---\ndate: 2026-03-15\n---\n"
	if !strings.HasPrefix(string(content), expectedFrontmatter) {
		t.Errorf("expected frontmatter %q, got content starting with %q", expectedFrontmatter, string(content)[:min(len(content), 50)])
	}

	if !strings.Contains(string(content), "I was flying over mountains") {
		t.Errorf("expected content to contain dream text")
	}
}

func TestExportAll_ShouldGenerateCorrectFilename(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		want      string
	}{
		{
			name:      "morning timestamp",
			createdAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			want:      "2026-03-15-08-30-00-dream.md",
		},
		{
			name:      "single digit values",
			createdAt: time.Date(2026, 1, 5, 9, 5, 5, 0, time.UTC),
			want:      "2026-01-05-09-05-05-dream.md",
		},
		{
			name:      "midnight",
			createdAt: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
			want:      "2026-12-25-00-00-00-dream.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dreams := []model.Dream{
				{ID: 1, Content: "test", CreatedAt: tt.createdAt, UpdatedAt: tt.createdAt},
			}

			_, err := ExportAll(dreams, tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			path := filepath.Join(tmpDir, tt.want)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("expected file %q to exist", tt.want)
			}
		})
	}
}

func TestExportAll_ShouldCreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "export", "path")

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "test",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, nestedDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if stat, err := os.Stat(nestedDir); err != nil || !stat.IsDir() {
		t.Errorf("expected directory to be created")
	}
}

func TestExportAll_ShouldOverwriteExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	createdAt := time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC)
	filename := "2026-03-15-08-30-00-dream.md"
	path := filepath.Join(tmpDir, filename)

	// Create an existing file
	if err := os.WriteFile(path, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "new dream content",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) == "old content" {
		t.Errorf("expected file to be overwritten, but old content remains")
	}

	if !strings.Contains(string(content), "new dream content") {
		t.Errorf("expected new content to be written")
	}
}

func TestExportAll_ShouldHandleEmptyDreamList(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 for empty list, got %d", count)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no files, found %d", len(entries))
	}
}

func TestExportAll_ShouldHandleSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	specialContent := `# Title with *emphasis* and **bold**
Contains "quotes" and 'apostrophes'
Line with \ backslash
	Tab at start
Code: \` + "`" + `func main() {}\` + "`" + `
[Link](http://example.com)`

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   specialContent,
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	filename := "2026-03-15-08-30-00-dream.md"
	path := filepath.Join(tmpDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Verify special characters are preserved
	if !strings.Contains(string(content), `# Title with *emphasis*`) {
		t.Errorf("special characters not preserved correctly")
	}
	if !strings.Contains(string(content), `"quotes"`) {
		t.Errorf("quotes not preserved")
	}
	if !strings.Contains(string(content), "[Link](http://example.com)") {
		t.Errorf("markdown link not preserved")
	}
}

func TestExportAll_ShouldExportMultipleDreams(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "First dream",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Content:   "Second dream",
			CreatedAt: time.Date(2026, 3, 14, 22, 15, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 14, 22, 15, 0, 0, time.UTC),
		},
		{
			ID:        3,
			Content:   "Third dream",
			CreatedAt: time.Date(2026, 3, 13, 6, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 13, 6, 0, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	expectedFiles := []string{
		"2026-03-15-08-30-00-dream.md",
		"2026-03-14-22-15-00-dream.md",
		"2026-03-13-06-00-00-dream.md",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %q to exist", filename)
		}
	}
}

func TestExportAll_ShouldPreserveNewlines(t *testing.T) {
	tmpDir := t.TempDir()
	content := "Line 1\nLine 2\n\nLine 4"

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   content,
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	_, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmpDir, "2026-03-15-08-30-00-dream.md")
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Extract just the content part (after frontmatter and blank line)
	contentStart := strings.Index(string(written), "---\n\n") + 5
	actualContent := string(written[contentStart:])

	if actualContent != content {
		t.Errorf("content mismatch\nexpected: %q\ngot: %q", content, actualContent)
	}
}

func TestExportAll_ShouldHandleUnicode(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "日本語の夢 🌙 émojis and ñ characters",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	_, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmpDir, "2026-03-15-08-30-00-dream.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "日本語の夢") {
		t.Errorf("unicode characters not preserved")
	}
	if !strings.Contains(string(content), "🌙") {
		t.Errorf("emoji not preserved")
	}
}

func TestExportAll_ShouldReturnErrorWhenDirectoryCannotBeCreated(t *testing.T) {
	// On Unix systems, we can't create directories in a read-only parent
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create read-only directory: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755)

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "test",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	_, err := ExportAll(dreams, filepath.Join(readOnlyDir, "subdir"))
	if err == nil {
		t.Skip("permission test skipped - may require elevated privileges")
	}

	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Errorf("expected error about directory creation, got: %v", err)
	}
}

func TestExportAll_ShouldHandleVeryLongContent(t *testing.T) {
	tmpDir := t.TempDir()
	longContent := strings.Repeat("This is a very long dream. ", 500)

	dreams := []model.Dream{
		{
			ID:        1,
			Content:   longContent,
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	path := filepath.Join(tmpDir, "2026-03-15-08-30-00-dream.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(content), longContent) {
		t.Errorf("expected long content to be written")
	}
}

func TestExportAll_ShouldContinueOnPartialFailures(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "First dream",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
		{
			ID:        2,
			Content:   "Second dream with same timestamp",
			CreatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC),
		},
	}

	count, err := ExportAll(dreams, tmpDir)
	// Should export at least one (last write wins for same timestamp)
	// or both if filesystem handles it
	if count == 0 {
		t.Errorf("expected at least one export, got %d", count)
	}
	// May or may not have error depending on filesystem behavior
	_ = err
}

func TestExportAll_ShouldIncludeDateInFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	dreams := []model.Dream{
		{
			ID:        1,
			Content:   "test",
			CreatedAt: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			UpdatedAt: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	_, err := ExportAll(dreams, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmpDir, "2026-12-31-23-59-59-dream.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// Verify date in frontmatter is just YYYY-MM-DD
	if !strings.Contains(string(content), "date: 2026-12-31") {
		t.Errorf("expected frontmatter to contain date only (YYYY-MM-DD)")
	}

	// Verify time is NOT in the frontmatter date
	if strings.Contains(string(content), "date: 2026-12-31 23:59:59") {
		t.Errorf("frontmatter should not include time")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
