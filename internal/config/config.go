package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	VellumDir        string `json:"vellum_dir"`
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
	// Per user request, check ~/vellum/config.json first or create it there?
	// Request said: "create config.json in ~/vellum/config.json when there is no config.js there"
	// and "Remove changing main vellum directory at all".
	// So we should assume VellumDir is ~/vellum and config is inside it.

	// Let's standardise on ~/vellum/config.json
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

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Detect defaults
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		var fileManager string
		// Simple heuristic
		if _, err := os.Stat("/Applications"); err == nil {
			fileManager = "open" // MacOS
		} else {
			fileManager = "xdg-open" // Linux fallback
		}

		home, _ := os.UserHomeDir()
		cfg := &Config{
			VellumDir:        filepath.Join(home, "vellum"),
			Theme:            "#BD93F9",
			LogRetentionDays: 7,
			DefaultEditor:    editor,
			SortOrder:        "name",
			DefaultShell:     shell,
			FileManager:      fileManager,
			MaxLogFiles:      100,
			LazyLoad:         true,
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
