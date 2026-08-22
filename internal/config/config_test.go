package config

import (
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, ".local", "share"))

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Theme != "#BD93F9" {
		t.Errorf("expected default theme #BD93F9, got %s", cfg.Theme)
	}
	if cfg.LogRetentionDays != 7 {
		t.Errorf("expected 7 retention days, got %d", cfg.LogRetentionDays)
	}
	if cfg.SortOrder != "name" {
		t.Errorf("expected sort order 'name', got %s", cfg.SortOrder)
	}
	if cfg.FileManager == "" {
		t.Errorf("expected detected file manager, got empty string")
	}

	// Verify persistence
	cfg.Theme = "#FF5555"
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if loaded.Theme != "#FF5555" {
		t.Errorf("expected theme #FF5555 after reload, got %s", loaded.Theme)
	}
}

func TestDetectDefaultFileManager(t *testing.T) {
	fm := DetectDefaultFileManager()
	if fm == "" {
		t.Errorf("DetectDefaultFileManager returned empty")
	}
}

func TestGetConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}
	expected := filepath.Join(tempDir, ".config", "vellum", "config.json")
	if path != expected {
		t.Errorf("expected config path %s, got %s", expected, path)
	}
}
