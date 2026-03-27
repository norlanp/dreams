package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDBPath_ShouldUseEnvOverride(t *testing.T) {
	t.Setenv(defaultDBPathEnv, "/custom/path/dreams.db")

	path, err := DefaultDBPath()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if path != "/custom/path/dreams.db" {
		t.Fatalf("expected env path, got %q", path)
	}
}

func TestIsGoRunBuildPath_ShouldDetectGoBuildTempPath(t *testing.T) {
	tmpDir := os.TempDir()
	goBuildPath := filepath.Join(tmpDir, "go-build12345", "b001", "exe")

	if !isGoRunBuildPath(goBuildPath) {
		t.Fatalf("expected %q to be detected as go-run build path", goBuildPath)
	}
}

func TestIsGoRunBuildPath_ShouldIgnoreNonTempPaths(t *testing.T) {
	if isGoRunBuildPath("/usr/local/bin") {
		t.Fatalf("expected non-temp path to not be detected as go-run build path")
	}
}

func TestValidateDBPath_ReturnsAbsolutePath(t *testing.T) {
	relPath := "./test.db"
	abs, _ := filepath.Abs(relPath)

	path, err := validateDBPath(relPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != abs {
		t.Errorf("expected %s, got %s", abs, path)
	}
}

func TestIsUnsafeDBLocation_DetectsUnsafeLocations(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"etc", "/etc/passwd", true},
		{"etc_subdir", "/etc/app/config", true},
		{"proc", "/proc/self", true},
		{"sys", "/sys/class", true},
		{"dev", "/dev/null", true},
		{"tmp", "/tmp/file", true},
		{"var_tmp", "/var/tmp/file", true},
		{"safe", "/home/user/db", false},
		{"safe_relative", "./db", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isUnsafeDBLocation(tc.path)
			if result != tc.expected {
				t.Errorf("isUnsafeDBLocation(%s) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestValidateDBPath_ResolvesNestedSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	realFile := filepath.Join(tmpDir, "real.db")
	link1 := filepath.Join(tmpDir, "link1.db")
	link2 := filepath.Join(tmpDir, "link2.db")

	if err := os.WriteFile(realFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if err := os.Symlink(realFile, link1); err != nil {
		t.Fatalf("failed to create link1: %v", err)
	}

	if err := os.Symlink(link1, link2); err != nil {
		t.Fatalf("failed to create link2: %v", err)
	}

	path, err := validateDBPath(link2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved, _ := filepath.EvalSymlinks(link2)
	if path != resolved {
		t.Errorf("expected resolved path %s, got %s", resolved, path)
	}
}

func TestValidateDBPath_NewFileReturnsAbsPath(t *testing.T) {
	tmpDir := t.TempDir()
	newPath := filepath.Join(tmpDir, "new.db")

	path, err := validateDBPath(newPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	abs, _ := filepath.Abs(newPath)
	if path != abs {
		t.Errorf("expected %s, got %s", abs, path)
	}
}

func TestValidateDBPath_RegularFileReturnsAbsPath(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "regular.db")
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	path, err := validateDBPath(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	abs, _ := filepath.Abs(tmpFile)
	if path != abs {
		t.Errorf("expected %s, got %s", abs, path)
	}
}
