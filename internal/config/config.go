package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	VellumDir string `json:"vellum_dir"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "vellum")
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
		// Return default
		home, _ := os.UserHomeDir()
		cfg := &Config{
			VellumDir: filepath.Join(home, "vellum"),
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