package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
)

var validItemName = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

type Store struct {
	cfg *config.Config
	mu  sync.RWMutex
}

func NewStore(cfg *config.Config) *Store {
	return &Store{cfg: cfg}
}

// ValidateName ensures item names cannot escape the directory boundary or contain illegal characters
func (s *Store) ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if !validItemName.MatchString(trimmed) || strings.Contains(trimmed, "..") || strings.ContainsAny(trimmed, "/\\") {
		return fmt.Errorf("invalid name '%s': only alphanumeric, dash, dot, and underscore allowed", name)
	}
	return nil
}

// atomicWrite writes data to a temporary file and atomically renames it
func (s *Store) atomicWrite(targetPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmpFile := filepath.Join(dir, fmt.Sprintf(".%s.%d.tmp", filepath.Base(targetPath), time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, data, perm); err != nil {
		return err
	}

	return os.Rename(tmpFile, targetPath)
}

func (s *Store) CreateLogFile(item model.Metadata) (*os.File, error) {
	return s.createLogFileInternal(item, false)
}

func (s *Store) CreateAdvLogFile(item model.Metadata) (*os.File, error) {
	return s.createLogFileInternal(item, true)
}

func (s *Store) createLogFileInternal(item model.Metadata, isAdv bool) (*os.File, error) {
	if err := s.ValidateName(item.Name); err != nil {
		return nil, err
	}

	dir := filepath.Join(s.GetItemDir(item), "logs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	prefix := ""
	if isAdv {
		prefix = "adv-"
	}
	// Safe, ISO 8601 lexicographical timestamp without colons
	timestamp := time.Now().UTC().Format("2006-01-02_15-04-05.000")
	filename := fmt.Sprintf("%s%s.log", prefix, timestamp)
	return os.OpenFile(filepath.Join(dir, filename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
}

func (s *Store) EnsureDirs() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dirs := []string{
		s.cfg.VellumDir,
		filepath.Join(s.cfg.VellumDir, "scripts"),
		filepath.Join(s.cfg.VellumDir, "aliases"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListItems() ([]model.Metadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []model.Metadata

	// Load scripts
	scriptsDir := filepath.Join(s.cfg.VellumDir, "scripts")
	if sEntries, err := os.ReadDir(scriptsDir); err == nil {
		for _, e := range sEntries {
			if e.IsDir() {
				metaPath := filepath.Join(scriptsDir, e.Name(), "metadata.json")
				meta, err := s.loadMetadata(metaPath)
				if err == nil {
					// Find the script file
					if files, err := os.ReadDir(filepath.Join(scriptsDir, e.Name())); err == nil {
						for _, f := range files {
							if strings.HasPrefix(f.Name(), "script.") {
								meta.FilePath = filepath.Join(scriptsDir, e.Name(), f.Name())
								break
							}
						}
					}
					items = append(items, meta)
				}
			}
		}
	}

	// Load aliases
	aliasesDir := filepath.Join(s.cfg.VellumDir, "aliases")
	if aEntries, err := os.ReadDir(aliasesDir); err == nil {
		for _, e := range aEntries {
			if e.IsDir() {
				metaPath := filepath.Join(aliasesDir, e.Name(), "metadata.json")
				meta, err := s.loadMetadata(metaPath)
				if err == nil {
					items = append(items, meta)
				}
			}
		}
	}

	// Sort items
	sort.Slice(items, func(i, j int) bool {
		switch s.cfg.SortOrder {
		case "lastrun":
			if !items[i].LastRunAt.Equal(items[j].LastRunAt) {
				return items[i].LastRunAt.After(items[j].LastRunAt)
			}
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		case "type":
			if items[i].Type != items[j].Type {
				return items[i].Type < items[j].Type
			}
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		default: // "name"
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
	})

	return items, nil
}

type logStat struct {
	path    string
	name    string
	modTime time.Time
}

func (s *Store) CleanLogs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cfg.LogRetentionDays <= 0 && s.cfg.MaxLogFiles <= 0 {
		return
	}

	cleanDir := func(logsDir string) {
		entries, err := os.ReadDir(logsDir)
		if err != nil {
			return
		}

		var logFiles []logStat
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
				if info, err := e.Info(); err == nil {
					logFiles = append(logFiles, logStat{
						path:    filepath.Join(logsDir, e.Name()),
						name:    e.Name(),
						modTime: info.ModTime(),
					})
				}
			}
		}

		// Sort by modification time (oldest first)
		sort.Slice(logFiles, func(i, j int) bool {
			return logFiles[i].modTime.Before(logFiles[j].modTime)
		})

		// 1. Retention Days Cleanup
		var surviving []logStat
		if s.cfg.LogRetentionDays > 0 {
			cutoff := time.Now().AddDate(0, 0, -s.cfg.LogRetentionDays)
			for _, f := range logFiles {
				if f.modTime.Before(cutoff) {
					_ = os.Remove(f.path)
				} else {
					surviving = append(surviving, f)
				}
			}
		} else {
			surviving = logFiles
		}

		// 2. Max Files Cleanup
		if s.cfg.MaxLogFiles > 0 && len(surviving) > s.cfg.MaxLogFiles {
			excess := len(surviving) - s.cfg.MaxLogFiles
			for i := 0; i < excess; i++ {
				_ = os.Remove(surviving[i].path)
			}
		}
	}

	for _, sub := range []string{"scripts", "aliases"} {
		parent := filepath.Join(s.cfg.VellumDir, sub)
		if entries, err := os.ReadDir(parent); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					cleanDir(filepath.Join(parent, e.Name(), "logs"))
				}
			}
		}
	}
}

func (s *Store) loadMetadata(path string) (model.Metadata, error) {
	var m model.Metadata
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func (s *Store) SaveScript(meta model.Metadata, content string) error {
	if err := s.ValidateName(meta.Name); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.cfg.VellumDir, "scripts", meta.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}

	metaPath := filepath.Join(dir, "metadata.json")
	if err := s.saveMetadata(metaPath, meta); err != nil {
		return err
	}

	scriptName := "script" + string(meta.ScriptType)
	scriptPath := filepath.Join(dir, scriptName)
	return s.atomicWrite(scriptPath, []byte(content), 0755)
}

func (s *Store) SaveAlias(meta model.Metadata) error {
	if err := s.ValidateName(meta.Name); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Join(s.cfg.VellumDir, "aliases", meta.Name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}

	return s.saveMetadata(filepath.Join(dir, "metadata.json"), meta)
}

func (s *Store) saveMetadata(path string, meta model.Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return s.atomicWrite(path, data, 0600)
}

func (s *Store) DeleteItem(item model.Metadata) error {
	if err := s.ValidateName(item.Name); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if item.Type == model.TypeAlias {
		_ = s.RemoveAliasFromShell(item)
	}

	dir := s.GetItemDir(item)
	// Safety check: ensure target directory is strictly within VellumDir
	cleanDir := filepath.Clean(dir)
	cleanVellum := filepath.Clean(s.cfg.VellumDir)
	if !strings.HasPrefix(cleanDir, cleanVellum) || cleanDir == cleanVellum {
		return fmt.Errorf("refusing to delete unsafe path: %s", dir)
	}

	return os.RemoveAll(dir)
}

func (s *Store) UpdateExecution(item model.Metadata, exitCode int, durationMs int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.GetItemDir(item)
	metaPath := filepath.Join(dir, "metadata.json")

	meta, err := s.loadMetadata(metaPath)
	if err != nil {
		meta = item
	}

	meta.LastRunAt = time.Now()
	meta.LastExitCode = exitCode
	meta.LastDurationMs = durationMs
	meta.RunCount++

	return s.saveMetadata(metaPath, meta)
}

func (s *Store) UpdateLastRun(item model.Metadata) error {
	return s.UpdateExecution(item, 0, 0)
}

func (s *Store) GetScriptContent(item model.Metadata) (string, error) {
	if item.FilePath == "" {
		return "", fmt.Errorf("file path not set")
	}
	content, err := os.ReadFile(item.FilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *Store) GetItemDir(item model.Metadata) string {
	if item.Type == model.TypeScript {
		return filepath.Join(s.cfg.VellumDir, "scripts", item.Name)
	}
	return filepath.Join(s.cfg.VellumDir, "aliases", item.Name)
}

func (s *Store) ListLogs(item model.Metadata) ([]string, error) {
	if err := s.ValidateName(item.Name); err != nil {
		return nil, err
	}

	logsDir := filepath.Join(s.GetItemDir(item), "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var logs []logStat
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			if info, err := e.Info(); err == nil {
				logs = append(logs, logStat{
					name:    e.Name(),
					path:    filepath.Join(logsDir, e.Name()),
					modTime: info.ModTime(),
				})
			}
		}
	}

	// Sort newest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].modTime.After(logs[j].modTime)
	})

	var names []string
	for _, l := range logs {
		names = append(names, l.name)
	}
	return names, nil
}

func (s *Store) ReadLog(item model.Metadata, logFilename string) (string, error) {
	if err := s.ValidateName(item.Name); err != nil {
		return "", err
	}

	cleanFilename := filepath.Base(logFilename)
	if !strings.HasSuffix(cleanFilename, ".log") {
		return "", fmt.Errorf("invalid log filename")
	}

	logPath := filepath.Join(s.GetItemDir(item), "logs", cleanFilename)
	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CopyScript copies an existing script to a new name
func (s *Store) CopyScript(src model.Metadata, newName string) error {
	content, err := s.GetScriptContent(src)
	if err != nil {
		return err
	}

	newMeta := src
	newMeta.Name = newName
	newMeta.ID = strings.ToLower(newName)
	newMeta.LastRunAt = time.Time{}
	newMeta.LastExitCode = 0
	newMeta.LastDurationMs = 0
	newMeta.RunCount = 0

	return s.SaveScript(newMeta, content)
}

// CopyAlias copies an existing alias to a new name
func (s *Store) CopyAlias(src model.Metadata, newName string) error {
	newMeta := src
	newMeta.Name = newName
	newMeta.ID = strings.ToLower(newName)
	newMeta.LastRunAt = time.Time{}
	newMeta.LastExitCode = 0
	newMeta.LastDurationMs = 0
	newMeta.RunCount = 0

	return s.SaveAlias(newMeta)
}

// ExportMarkdown generates a comprehensive markdown summary of all scripts and aliases
func (s *Store) ExportMarkdown() (string, error) {
	items, err := s.ListItems()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Vellum Export - %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Total Items: %d\n\n", len(items)))

	sb.WriteString("## Scripts\n\n")
	for _, item := range items {
		if item.Type == model.TypeScript {
			sb.WriteString(fmt.Sprintf("### %s (`%s`)\n", item.Name, item.ScriptType))
			if item.Desc != "" {
				sb.WriteString(fmt.Sprintf("*%s*\n\n", item.Desc))
			}
			sb.WriteString(fmt.Sprintf("- **Requires Sudo:** %v\n", item.RequiresSudo))
			sb.WriteString(fmt.Sprintf("- **Runs:** %d\n", item.RunCount))
			if !item.LastRunAt.IsZero() {
				sb.WriteString(fmt.Sprintf("- **Last Run:** %s (Exit %d, %dms)\n", item.LastRunAt.Format(time.RFC3339), item.LastExitCode, item.LastDurationMs))
			}
			content, err := s.GetScriptContent(item)
			if err == nil && content != "" {
				ext := strings.TrimPrefix(string(item.ScriptType), ".")
				sb.WriteString(fmt.Sprintf("\n```%s\n%s\n```\n", ext, content))
			}
			sb.WriteString("\n---\n\n")
		}
	}

	sb.WriteString("## Aliases\n\n")
	for _, item := range items {
		if item.Type == model.TypeAlias {
			sb.WriteString(fmt.Sprintf("### %s\n", item.Name))
			if item.Desc != "" {
				sb.WriteString(fmt.Sprintf("*%s*\n\n", item.Desc))
			}
			sb.WriteString(fmt.Sprintf("- **Command:** `%s`\n", item.Command))
			sb.WriteString(fmt.Sprintf("- **Requires Sudo:** %v\n", item.RequiresSudo))
			sb.WriteString(fmt.Sprintf("- **Runs:** %d\n\n", item.RunCount))
		}
	}

	return sb.String(), nil
}

