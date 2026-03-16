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
