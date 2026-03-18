package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigureLogging_ShouldWriteToFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "var", "dreams.log")
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	previousPrefix := log.Prefix()

	closer, err := configureLogging(logPath)
	if err != nil {
		t.Fatalf("expected logging setup to succeed: %v", err)
	}

	t.Cleanup(func() {
		_ = closer.Close()
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	})

	log.Printf("priming-log-test")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file read to succeed: %v", err)
	}

	if !strings.Contains(string(data), "priming-log-test") {
		t.Fatalf("expected log file to contain emitted line, got %q", string(data))
	}
}

func TestLoadEnvFile_ShouldIgnoreMissingFile(t *testing.T) {
	err := loadEnvFile(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatalf("expected missing env file to be ignored: %v", err)
	}
}

func TestLoadEnvFile_ShouldReturnParseError(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), "broken.env")
	err := os.WriteFile(envPath, []byte("BROKEN='\n"), 0o600)
	if err != nil {
		t.Fatalf("expected broken env file write to succeed: %v", err)
	}

	err = loadEnvFile(envPath)
	if err == nil {
		t.Fatal("expected parse error when env file is invalid")
	}
	if !strings.Contains(err.Error(), "failed to load env file") {
		t.Fatalf("expected env load context in error, got %v", err)
	}
}

func TestLoadEnvFile_ShouldLoadVariables(t *testing.T) {
	const key = "DREAMS_TEST_ENV"

	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})

	envPath := filepath.Join(t.TempDir(), "valid.env")
	err := os.WriteFile(envPath, []byte(key+"=loaded\n"), 0o600)
	if err != nil {
		t.Fatalf("expected valid env file write to succeed: %v", err)
	}

	err = loadEnvFile(envPath)
	if err != nil {
		t.Fatalf("expected env file to load: %v", err)
	}

	if os.Getenv(key) != "loaded" {
		t.Fatalf("expected %s to be loaded from env file", key)
	}
}
