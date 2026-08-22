package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	VellumDir        string `json:"vellum_dir,omitempty"` // Base data directory for scripts/aliases
	Theme            string `json:"theme"`                // Hex color code
	LogRetentionDays int    `json:"log_retention_days"`   // Default 7
	DefaultEditor    string `json:"default_editor"`       // nano/vim/micro/code
	SortOrder        string `json:"sort_order"`           // name/lastrun/type
	DefaultShell     string `json:"default_shell"`        // zsh/bash/fish
	FileManager      string `json:"file_manager"`         // open/xdg-open/explorer
	MaxLogFiles      int    `json:"max_log_files"`        // Limit before cleanup
	LazyLoad         bool   `json:"lazy_load"`
	CacheEnabled     bool   `json:"cache_enabled"`
}

// GetDefaultDataDir returns the base directory for scripts, aliases, and logs.
func GetDefaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "vellum"
	}

	// Legacy backward compatibility: if ~/vellum exists, keep using it
	legacyDir := filepath.Join(home, "vellum")
	if info, err := os.Stat(legacyDir); err == nil && info.IsDir() {
		return legacyDir
	}

	// XDG Data Home
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "vellum")
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "vellum")
	}

	return filepath.Join(home, ".local", "share", "vellum")
}

// GetConfigPath returns the path to config.json respecting XDG standards.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check legacy location first
	legacyConfig := filepath.Join(home, "vellum", "config.json")
	if _, err := os.Stat(legacyConfig); err == nil {
		return legacyConfig, nil
	}

	// XDG Config Home
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(home, ".config")
	}
	vellumConfigDir := filepath.Join(configDir, "vellum")
	if err := os.MkdirAll(vellumConfigDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(vellumConfigDir, "config.json"), nil
}

// DetectDefaultFileManager selects the appropriate file manager for the OS.
func DetectDefaultFileManager() string {
	switch runtime.GOOS {
	case "darwin":
		return "open"
	case "windows":
		return "explorer"
	default:
		return "xdg-open"
	}
}

// LoadConfig loads the configuration from disk, applying sane defaults.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	defaultDataDir := GetDefaultDataDir()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		cfg := &Config{
			VellumDir:        defaultDataDir,
			Theme:            "#BD93F9", // Dracula Purple
			LogRetentionDays: 7,
			DefaultEditor:    editor,
			SortOrder:        "name",
			DefaultShell:     shell,
			FileManager:      DetectDefaultFileManager(),
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

	if cfg.VellumDir == "" {
		cfg.VellumDir = defaultDataDir
	}
	if cfg.Theme == "" {
		cfg.Theme = "#BD93F9"
	}
	if cfg.LogRetentionDays == 0 {
		cfg.LogRetentionDays = 7
	}
	if cfg.DefaultShell == "" {
		cfg.DefaultShell = os.Getenv("SHELL")
		if cfg.DefaultShell == "" {
			cfg.DefaultShell = "/bin/bash"
		}
	}
	if cfg.FileManager == "" {
		cfg.FileManager = DetectDefaultFileManager()
	}
	if cfg.SortOrder == "" {
		cfg.SortOrder = "name"
	}
	if cfg.DefaultEditor == "" {
		cfg.DefaultEditor = os.Getenv("EDITOR")
		if cfg.DefaultEditor == "" {
			cfg.DefaultEditor = "nano"
		}
	}

	return &cfg, nil
}

// SaveConfig persists the configuration to disk.
func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
