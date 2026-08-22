package store

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jurekzsl/vellum/internal/model"
)

var validAliasName = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

func getShellRcInfo() (string, string) {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return ".zshrc", "zsh"
	} else if strings.Contains(shell, "fish") {
		return ".config/fish/config.fish", "fish"
	}
	return ".bashrc", "bash"
}

// formatAliasDef safely formats an alias definition avoiding shell injection
func formatAliasDef(name, command, shellType string) (string, error) {
	if !validAliasName.MatchString(name) {
		return "", fmt.Errorf("invalid alias name '%s'", name)
	}

	if shellType == "fish" {
		escaped := strings.ReplaceAll(command, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "\\'")
		return fmt.Sprintf("alias %s '%s'", name, escaped), nil
	}

	// POSIX / bash / zsh
	escaped := strings.ReplaceAll(command, "'", `'\''`)
	return fmt.Sprintf("alias %s='%s'", name, escaped), nil
}

func (s *Store) AddAliasToShell(meta model.Metadata) error {
	rcFile, shellType := getShellRcInfo()
	aliasCmd, err := formatAliasDef(meta.Name, meta.Command, shellType)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(home, rcFile)

	// Remove any existing entry first to prevent duplicates
	_ = s.RemoveAliasFromShell(meta)

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := fmt.Sprintf("\n# [Vellum Alias: %s]\n%s\n", meta.Name, aliasCmd)
	_, err = f.WriteString(entry)
	return err
}

func (s *Store) RemoveAliasFromShell(meta model.Metadata) error {
	rcFile, _ := getShellRcInfo()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(home, rcFile)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	skipNext := false
	tag := fmt.Sprintf("# [Vellum Alias: %s]", meta.Name)

	for _, line := range lines {
		if strings.TrimSpace(line) == tag {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(fullPath, []byte(strings.Join(newLines, "\n")), 0644)
}

