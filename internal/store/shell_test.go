package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jurekzsl/vellum/internal/model"
)

func TestFormatAliasDef(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		shellType string
		want      string
		wantErr   bool
	}{
		{
			name:      "simple",
			command:   "ls -la",
			shellType: "bash",
			want:      "alias simple='ls -la'",
			wantErr:   false,
		},
		{
			name:      "with_quotes",
			command:   "echo 'hello world'",
			shellType: "bash",
			want:      `alias with_quotes='echo '\''hello world'\'''`,
			wantErr:   false,
		},
		{
			name:      "fish_shell",
			command:   "echo 'hello'",
			shellType: "fish",
			want:      `alias fish_shell 'echo \'hello\''`,
			wantErr:   false,
		},
		{
			name:      "bad name with spaces",
			command:   "ls",
			shellType: "bash",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		got, err := formatAliasDef(tt.name, tt.command, tt.shellType)
		if (err != nil) != tt.wantErr {
			t.Errorf("formatAliasDef(%q, %q, %q) error = %v, wantErr %v", tt.name, tt.command, tt.shellType, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("formatAliasDef(%q, %q, %q) = %q, want %q", tt.name, tt.command, tt.shellType, got, tt.want)
		}
	}
}

func TestShellRCHandling(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("SHELL", "/bin/bash")

	s, _ := setupTestStore(t)

	meta := model.Metadata{
		ID:      "testalias",
		Name:    "testalias",
		Type:    model.TypeAlias,
		Command: "echo 'vellum alias works'",
	}

	// Add alias
	if err := s.AddAliasToShell(meta); err != nil {
		t.Fatalf("AddAliasToShell failed: %v", err)
	}

	rcData, err := os.ReadFile(filepath.Join(tempHome, ".bashrc"))
	if err != nil {
		t.Fatalf("Failed to read .bashrc: %v", err)
	}
	if !strings.Contains(string(rcData), "testalias") {
		t.Errorf("expected .bashrc to contain testalias, got: %s", string(rcData))
	}

	// Remove alias
	if err := s.RemoveAliasFromShell(meta); err != nil {
		t.Fatalf("RemoveAliasFromShell failed: %v", err)
	}

	rcDataAfter, _ := os.ReadFile(filepath.Join(tempHome, ".bashrc"))
	if strings.Contains(string(rcDataAfter), "testalias") {
		t.Errorf("expected testalias to be removed from .bashrc, got: %s", string(rcDataAfter))
	}
}
