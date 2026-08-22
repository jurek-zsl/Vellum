package runner

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jurekzsl/vellum/internal/model"
)

func TestPrepareCommand(t *testing.T) {
	tempDir := t.TempDir()
	scriptFile := filepath.Join(tempDir, "test.sh")
	if err := os.WriteFile(scriptFile, []byte("#!/bin/sh\necho 'hello'\n"), 0755); err != nil {
		t.Fatalf("failed to create script: %v", err)
	}

	meta := model.Metadata{
		ID:         "test",
		Name:       "test",
		Type:       model.TypeScript,
		ScriptType: model.ScriptTypeShell,
		FilePath:   scriptFile,
	}

	cmd, err := PrepareCommand(meta, []string{"--arg1", "val"}, nil, "/bin/sh", "")
	if err != nil {
		t.Fatalf("PrepareCommand failed: %v", err)
	}
	if cmd == nil {
		t.Fatalf("expected non-nil cmd")
	}
}

func TestExecuteHeadless(t *testing.T) {
	meta := model.Metadata{
		ID:      "echo_test",
		Name:    "echo_test",
		Type:    model.TypeAlias,
		Command: "echo 'vellum headless works'",
	}

	var stdout, stderr bytes.Buffer
	exitCode, durationMs, err := ExecuteHeadless(meta, nil, "/bin/sh", "", &stdout, &stderr)
	if err != nil {
		t.Fatalf("ExecuteHeadless failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exitCode 0, got %d", exitCode)
	}
	if durationMs < 0 {
		t.Errorf("expected duration >= 0, got %d", durationMs)
	}
	if stdout.String() != "vellum headless works\n" {
		t.Errorf("unexpected stdout: %q", stdout.String())
	}
}

func TestGetInterpreter(t *testing.T) {
	types := []model.ScriptType{
		// Coding
		model.ScriptTypePython,
		model.ScriptTypeJavascript,
		model.ScriptTypeTypescript,
		model.ScriptTypeJSX,
		model.ScriptTypeTSX,
		model.ScriptTypeRuby,
		model.ScriptTypePerl,
		model.ScriptTypePHP,
		model.ScriptTypeLua,
		model.ScriptTypeTcl,
		// Shell
		model.ScriptTypeShell,
		model.ScriptTypeBash,
		model.ScriptTypeZsh,
		model.ScriptTypeFish,
		model.ScriptTypePowershell,
		// Compiled
		model.ScriptTypeGo,
		model.ScriptTypeRust,
		model.ScriptTypeJava,
		model.ScriptTypeKotlin,
		model.ScriptTypeSwift,
		// System
		model.ScriptTypeAppleScript,
		model.ScriptTypeBat,
		model.ScriptTypeCmd,
		model.ScriptTypeVBS,
	}

	for _, st := range types {
		interp := getInterpreter(st)
		if interp == "" {
			t.Errorf("expected interpreter for script type %s, got empty", st)
		}
	}
}
