package closedsessions

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseDateRequestExactDate(t *testing.T) {
	now := time.Date(2025, 4, 10, 12, 0, 0, 0, time.UTC)
	got, err := ParseDateRequest("20250404", now)
	if err != nil {
		t.Fatalf("ParseDateRequest returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 date, got %d", len(got))
	}
	if got[0].Format("20060102") != "20250404" {
		t.Fatalf("expected 20250404, got %s", got[0].Format("20060102"))
	}
}

func TestParseDateRequestRelativeDays(t *testing.T) {
	now := time.Date(2025, 4, 10, 12, 0, 0, 0, time.UTC)
	got, err := ParseDateRequest("4d", now)
	if err != nil {
		t.Fatalf("ParseDateRequest returned error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 dates, got %d", len(got))
	}
	wanted := []string{"20250410", "20250409", "20250408", "20250407"}
	for i, expected := range wanted {
		if got[i].Format("20060102") != expected {
			t.Fatalf("expected date %s at index %d, got %s", expected, i, got[i].Format("20060102"))
		}
	}
}

func TestSearchMatchesNestedPaths(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "nested", "archive"), 0o755); err != nil {
		t.Fatalf("creating nested directories: %v", err)
	}

	files := []string{
		filepath.Join(base, "nested", "restore-session-20250410.txt"),
		filepath.Join(base, "nested", "archive", "restore-session-20250409.log"),
		filepath.Join(base, "other", "restore-session-20250411.zip"),
		filepath.Join(base, "ignore", "restore-session-20250301.txt"),
		filepath.Join(base, "nested", "restore-session-auto-20250410.txt"),
	}
	if err := os.MkdirAll(filepath.Join(base, "other"), 0o755); err != nil {
		t.Fatalf("creating other directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(base, "ignore"), 0o755); err != nil {
		t.Fatalf("creating ignore directory: %v", err)
	}
	for _, path := range files {
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("writing file %s: %v", path, err)
		}
	}

	results, err := Search(base, "20250410")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
	if results[0].Directory != "nested" {
		t.Fatalf("expected directory nested, got %q", results[0].Directory)
	}
	if results[0].Filename != "restore-session-20250410.txt" {
		t.Fatalf("expected filename restore-session-20250410.txt, got %q", results[0].Filename)
	}
	if results[1].Filename != "restore-session-auto-20250410.txt" {
		t.Fatalf("expected filename restore-session-auto-20250410.txt, got %q", results[1].Filename)
	}
}

func TestLoadConfigCreatesDefaultFile(t *testing.T) {
	oldHome := os.Getenv("HOME")
	tempHome := t.TempDir()
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("setting HOME: %v", err)
	}
	defer func() { _ = os.Setenv("HOME", oldHome) }()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.BaseSearchPath != filepath.Join(tempHome, "Work") {
		t.Fatalf("expected default base path %s, got %s", filepath.Join(tempHome, "Work"), cfg.BaseSearchPath)
	}
	if _, err := os.Stat(ConfigFile()); err != nil {
		t.Fatalf("expected config file to exist at %s: %v", ConfigFile(), err)
	}
}
