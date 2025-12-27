package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jurekzsl/vellum/internal/model"
)

func (s *Store) AddAliasToShell(meta model.Metadata) error {
	shell := os.Getenv("SHELL")
	rcFile := ""
	aliasCmd := ""

	if strings.Contains(shell, "zsh") {
		rcFile = ".zshrc"
		aliasCmd = fmt.Sprintf("alias %s='%s'", meta.Name, meta.Command)
	} else if strings.Contains(shell, "bash") {
		rcFile = ".bashrc"
		aliasCmd = fmt.Sprintf("alias %s='%s'", meta.Name, meta.Command)
	} else if strings.Contains(shell, "fish") {
		rcFile = ".config/fish/config.fish"
		aliasCmd = fmt.Sprintf("alias %s '%s'", meta.Name, meta.Command)
	} else {
		// Default to bashrc if unknown
		rcFile = ".bashrc"
		aliasCmd = fmt.Sprintf("alias %s='%s'", meta.Name, meta.Command)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(home, rcFile)

	// Check if already exists?
	// For now, just append.
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("\n# Added by Vellum\n%s\n", aliasCmd)); err != nil {
		return err
	}

	return nil
}
