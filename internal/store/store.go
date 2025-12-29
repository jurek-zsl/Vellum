package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jurekzsl/vellum/internal/config"
	"github.com/jurekzsl/vellum/internal/model"
)

type Store struct {
	cfg *config.Config
}

func NewStore(cfg *config.Config) *Store {
	return &Store{cfg: cfg}
}

func (s *Store) CreateLogFile(item model.Metadata) (*os.File, error) {
	var dir string
	if item.Type == model.TypeScript {
		dir = filepath.Join(s.cfg.VellumDir, "scripts", item.Name, "logs")
	} else {
		dir = filepath.Join(s.cfg.VellumDir, "aliases", item.Name, "logs")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("02:01:2006-15:04:05")
	filename := fmt.Sprintf("%s.log", timestamp)
	return os.Create(filepath.Join(dir, filename))
}

func (s *Store) CreateAdvLogFile(item model.Metadata) (*os.File, error) {
	var dir string
	if item.Type == model.TypeScript {
		dir = filepath.Join(s.cfg.VellumDir, "scripts", item.Name, "logs")
	} else {
		dir = filepath.Join(s.cfg.VellumDir, "aliases", item.Name, "logs")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	timestamp := time.Now().Format("02:01:2006-15:04:05")
	filename := fmt.Sprintf("adv-%s.log", timestamp)
	return os.Create(filepath.Join(dir, filename))
}

func (s *Store) EnsureDirs() error {
	dirs := []string{
		s.cfg.VellumDir,
		filepath.Join(s.cfg.VellumDir, "scripts"),
		filepath.Join(s.cfg.VellumDir, "aliases"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListItems() ([]model.Metadata, error) {
	var items []model.Metadata

	// Load scripts
	scriptsDir := filepath.Join(s.cfg.VellumDir, "scripts")
	sEntries, err := os.ReadDir(scriptsDir)
	if err == nil {
		for _, e := range sEntries {
			if e.IsDir() {
				metaPath := filepath.Join(scriptsDir, e.Name(), "metadata.json")
				meta, err := s.loadMetadata(metaPath)
				if err == nil {
					// Find the script file
					files, _ := os.ReadDir(filepath.Join(scriptsDir, e.Name()))
					for _, f := range files {
						if strings.HasPrefix(f.Name(), "script.") {
							meta.FilePath = filepath.Join(scriptsDir, e.Name(), f.Name())
							break
						}
					}
					items = append(items, meta)
				}
			}
		}
	}

	// Load aliases
	aliasesDir := filepath.Join(s.cfg.VellumDir, "aliases")
	aEntries, err := os.ReadDir(aliasesDir)
	if err == nil {
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
			// Fallback to name if timestamps are equal
			return items[i].Name < items[j].Name
		case "type":
			if items[i].Type != items[j].Type {
				return items[i].Type < items[j].Type
			}
			return items[i].Name < items[j].Name
		default: // "name"
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
	})

	return items, nil
}

func (s *Store) CleanLogs() {
	if s.cfg.LogRetentionDays <= 0 && s.cfg.MaxLogFiles <= 0 {
		return
	}

	// Helper to clean logs in a specific directory
	cleanDir := func(logsDir string) {
		entries, err := os.ReadDir(logsDir)
		if err != nil {
			return
		}

		var logFiles []os.DirEntry
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
				logFiles = append(logFiles, e)
			}
		}

		// Sort by modification time (oldest first)
		sort.Slice(logFiles, func(i, j int) bool {
			iv, _ := logFiles[i].Info()
			jv, _ := logFiles[j].Info()
			return iv.ModTime().Before(jv.ModTime())
		})

		// 1. Retention Days Cleanup
		if s.cfg.LogRetentionDays > 0 {
			cutoff := time.Now().AddDate(0, 0, -s.cfg.LogRetentionDays)
			for _, f := range logFiles {
				info, _ := f.Info()
				if info.ModTime().Before(cutoff) {
					os.Remove(filepath.Join(logsDir, f.Name()))
				}
			}
		}

		// 2. Max Files Cleanup
		if s.cfg.MaxLogFiles > 0 {
			// Re-read entries remaining (or filter list)
			// Efficiently filter the already sorted list if we removed items?
			// Simpler to just re-scan or track deletions.
			// Let's re-scan to be safe and simple.
			entries, _ := os.ReadDir(logsDir)
			var validLogs []os.DirEntry
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
					validLogs = append(validLogs, e)
				}
			}

			if len(validLogs) > s.cfg.MaxLogFiles {
				sort.Slice(validLogs, func(i, j int) bool {
					iv, _ := validLogs[i].Info()
					jv, _ := validLogs[j].Info()
					return iv.ModTime().Before(jv.ModTime())
				})

				removeCount := len(validLogs) - s.cfg.MaxLogFiles
				for i := 0; i < removeCount; i++ {
					os.Remove(filepath.Join(logsDir, validLogs[i].Name()))
				}
			}
		}
	}

	// iterate scripts
	scriptsDir := filepath.Join(s.cfg.VellumDir, "scripts")
	if entries, err := os.ReadDir(scriptsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				cleanDir(filepath.Join(scriptsDir, e.Name(), "logs"))
			}
		}
	}

	// iterate aliases
	aliasesDir := filepath.Join(s.cfg.VellumDir, "aliases")
	if entries, err := os.ReadDir(aliasesDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				cleanDir(filepath.Join(aliasesDir, e.Name(), "logs"))
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
	// Create folder
	dir := filepath.Join(s.cfg.VellumDir, "scripts", meta.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Save metadata
	metaPath := filepath.Join(dir, "metadata.json")
	if err := s.saveMetadata(metaPath, meta); err != nil {
		return err
	}

	// Save script content
	scriptName := "script" + string(meta.ScriptType)
	scriptPath := filepath.Join(dir, scriptName)
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		return err
	}

	return nil
}

func (s *Store) SaveAlias(meta model.Metadata) error {
	dir := filepath.Join(s.cfg.VellumDir, "aliases", meta.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return s.saveMetadata(filepath.Join(dir, "metadata.json"), meta)
}

func (s *Store) saveMetadata(path string, meta model.Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Store) DeleteItem(item model.Metadata) error {
	dir := s.GetItemDir(item)
	return os.RemoveAll(dir)
}

func (s *Store) UpdateLastRun(item model.Metadata) error {
	item.LastRunAt = time.Now()
	dir := s.GetItemDir(item)
	return s.saveMetadata(filepath.Join(dir, "metadata.json"), item)
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

// CopyScript copies an existing script to a new name
func (s *Store) CopyScript(src model.Metadata, newName string) error {
	content, err := s.GetScriptContent(src)
	if err != nil {
		return err
	}

	newMeta := src
	newMeta.Name = newName
	newMeta.ID = strings.ToLower(newName)
	newMeta.LastRunAt = time.Time{} // Reset run time

	// Create new directory and save
	return s.SaveScript(newMeta, content)
}

// CopyAlias copies an existing alias to a new name
func (s *Store) CopyAlias(src model.Metadata, newName string) error {
	newMeta := src
	newMeta.Name = newName
	newMeta.ID = strings.ToLower(newName)
	newMeta.LastRunAt = time.Time{}

	return s.SaveAlias(newMeta)
}
