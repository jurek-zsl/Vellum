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

	// Sort by Name
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
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
	var dir string
	if item.Type == model.TypeScript {
		dir = filepath.Join(s.cfg.VellumDir, "scripts", item.Name)
	} else {
		dir = filepath.Join(s.cfg.VellumDir, "aliases", item.Name)
	}
	return os.RemoveAll(dir)
}

func (s *Store) UpdateLastRun(item model.Metadata) error {
	item.LastRunAt = time.Now()
	var dir string
	if item.Type == model.TypeScript {
		dir = filepath.Join(s.cfg.VellumDir, "scripts", item.Name)
	} else {
		dir = filepath.Join(s.cfg.VellumDir, "aliases", item.Name)
	}
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
