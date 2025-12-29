package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	VellumDir        string `json:"-"`                  // Hardcoded to ~/vellum
	Theme            string `json:"theme"`              // hex color
	LogRetentionDays int    `json:"log_retention_days"` // default 7
	DefaultEditor    string `json:"default_editor"`     // nano/vim/micro/code
	SortOrder        string `json:"sort_order"`         // name/lastrun/type
	DefaultShell     string `json:"default_shell"`      // zsh/bash
	FileManager      string `json:"file_manager"`       // finder/explorer
	MaxLogFiles      int    `json:"max_log_files"`      // before cleanup
	LazyLoad         bool   `json:"lazy_load"`
	CacheEnabled     bool   `json:"cache_enabled"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// VellumDir is always ~/vellum and config is inside it.
	dir := filepath.Join(home, "vellum")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Hardcoded VellumDir
	home, _ := os.UserHomeDir()
	vellumDir := filepath.Join(home, "vellum")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Detect defaults
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		// Editor defaults to nano if EDITOR env not set
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		cfg := &Config{
			VellumDir:        vellumDir,
			Theme:            "#BD93F9", // Dracula Purple
			LogRetentionDays: 7,
			DefaultEditor:    editor,
			SortOrder:        "name",
			DefaultShell:     shell,
			FileManager:      "xdg-open",
			MaxLogFiles:      50,
			LazyLoad:         false,
			CacheEnabled:     true,
		}
		_ = SaveConfig(cfg)
		return cfg, nil
	} else if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Ensure VellumDir is set (it's ignored in JSON so it will be empty after Unmarshal)
	cfg.VellumDir = vellumDir

	// Set defaults for missing fields if older config exists
	if cfg.Theme == "" {
		cfg.Theme = "#BD93F9"
	}
	if cfg.LogRetentionDays == 0 {
		cfg.LogRetentionDays = 7
	}
	if cfg.DefaultShell == "" {
		cfg.DefaultShell = os.Getenv("SHELL")
	}
	if cfg.SortOrder == "" {
		cfg.SortOrder = "name"
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
