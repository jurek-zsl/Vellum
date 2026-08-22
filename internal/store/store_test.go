package store

import (
	"testing"
	"time"

	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
)

func setupTestStore(t *testing.T) (*Store, *config.Config) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		VellumDir:        tempDir,
		Theme:            "#BD93F9",
		LogRetentionDays: 7,
		MaxLogFiles:      5,
		SortOrder:        "name",
	}
	s := NewStore(cfg)
	if err := s.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}
	return s, cfg
}

func TestValidateName(t *testing.T) {
	s, _ := setupTestStore(t)

	validNames := []string{"myscript", "my-script_v2.0", "backup_db", "test123"}
	for _, name := range validNames {
		if err := s.ValidateName(name); err != nil {
			t.Errorf("expected '%s' to be valid, got error: %v", name, err)
		}
	}

	invalidNames := []string{
		"",
		" ",
		"../escape",
		"../../etc/passwd",
		"foo/bar",
		"foo\\bar",
		"bad$name",
		"test;rm -rf /",
	}
	for _, name := range invalidNames {
		if err := s.ValidateName(name); err == nil {
			t.Errorf("expected '%s' to fail validation, but it passed", name)
		}
	}
}

func TestScriptCRUD(t *testing.T) {
	s, _ := setupTestStore(t)

	meta := model.Metadata{
		ID:         "deploy",
		Name:       "deploy",
		Desc:       "Deploy script",
		Type:       model.TypeScript,
		ScriptType: model.ScriptTypeShell,
	}
	content := "#!/bin/bash\necho 'Deploying...'\n"

	// 1. Save
	if err := s.SaveScript(meta, content); err != nil {
		t.Fatalf("SaveScript failed: %v", err)
	}

	// 2. List
	items, err := s.ListItems()
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "deploy" {
		t.Errorf("expected item name 'deploy', got '%s'", items[0].Name)
	}

	// 3. Read Content
	gotContent, err := s.GetScriptContent(items[0])
	if err != nil {
		t.Fatalf("GetScriptContent failed: %v", err)
	}
	if gotContent != content {
		t.Errorf("content mismatch: got %q, want %q", gotContent, content)
	}

	// 4. Update Execution Stats
	if err := s.UpdateExecution(items[0], 0, 150); err != nil {
		t.Fatalf("UpdateExecution failed: %v", err)
	}
	itemsUpdated, _ := s.ListItems()
	if itemsUpdated[0].RunCount != 1 || itemsUpdated[0].LastExitCode != 0 || itemsUpdated[0].LastDurationMs != 150 {
		t.Errorf("stats not updated correctly: %+v", itemsUpdated[0])
	}

	// 5. Delete
	if err := s.DeleteItem(items[0]); err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}
	itemsAfterDelete, _ := s.ListItems()
	if len(itemsAfterDelete) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(itemsAfterDelete))
	}
}

func TestAliasCRUD(t *testing.T) {
	s, _ := setupTestStore(t)

	meta := model.Metadata{
		ID:      "ll",
		Name:    "ll",
		Desc:    "List files detailed",
		Type:    model.TypeAlias,
		Command: "ls -la",
	}

	if err := s.SaveAlias(meta); err != nil {
		t.Fatalf("SaveAlias failed: %v", err)
	}

	items, err := s.ListItems()
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}
	if len(items) != 1 || items[0].Name != "ll" {
		t.Fatalf("expected alias 'll', got %+v", items)
	}

	if err := s.DeleteItem(items[0]); err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}
}

func TestLoggingAndCleanLogs(t *testing.T) {
	s, cfg := setupTestStore(t)

	meta := model.Metadata{
		ID:         "logger",
		Name:       "logger",
		Type:       model.TypeScript,
		ScriptType: model.ScriptTypePython,
	}
	if err := s.SaveScript(meta, "print('hello')"); err != nil {
		t.Fatalf("SaveScript failed: %v", err)
	}

	// Create 7 log files
	for i := 0; i < 7; i++ {
		f, err := s.CreateLogFile(meta)
		if err != nil {
			t.Fatalf("CreateLogFile failed: %v", err)
		}
		f.WriteString("log output line\n")
		f.Close()
		time.Sleep(2 * time.Millisecond) // ensure unique timestamp
	}

	logs, err := s.ListLogs(meta)
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}
	if len(logs) != 7 {
		t.Fatalf("expected 7 logs before cleanup, got %d", len(logs))
	}

	// Test CleanLogs (MaxLogFiles is 5)
	s.CleanLogs()

	logsAfterMax, err := s.ListLogs(meta)
	if err != nil {
		t.Fatalf("ListLogs failed: %v", err)
	}
	if len(logsAfterMax) != cfg.MaxLogFiles {
		t.Errorf("expected %d logs after max cleanup, got %d", cfg.MaxLogFiles, len(logsAfterMax))
	}

	// Verify ReadLog
	data, err := s.ReadLog(meta, logsAfterMax[0])
	if err != nil {
		t.Fatalf("ReadLog failed: %v", err)
	}
	if data != "log output line\n" {
		t.Errorf("unexpected log content: %q", data)
	}
}

func TestExportMarkdown(t *testing.T) {
	s, _ := setupTestStore(t)

	meta := model.Metadata{
		ID:         "hello",
		Name:       "hello",
		Desc:       "Greeter",
		Type:       model.TypeScript,
		ScriptType: model.ScriptTypePython,
	}
	_ = s.SaveScript(meta, "print('hello world')")

	md, err := s.ExportMarkdown()
	if err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}
	if md == "" {
		t.Errorf("expected non-empty markdown output")
	}
}
