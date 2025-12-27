package store

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jurekzsl/vellum/internal/model"
)

func (s *Store) CreateLogFile(item model.Metadata) (*os.File, error) {
	var dir string
	if item.Type == model.TypeScript {
		dir = filepath.Join(s.cfg.VellumDir, "scripts", item.Name)
	} else {
		dir = filepath.Join(s.cfg.VellumDir, "aliases", item.Name)
	}

	timestamp := time.Now().Format("02:01:2006-15:04:05")
	filename := fmt.Sprintf("%s.log", timestamp)
	return os.Create(filepath.Join(dir, filename))
}

func (s *Store) CleanLogs() error {
	dirs := []string{
		filepath.Join(s.cfg.VellumDir, "scripts"),
		filepath.Join(s.cfg.VellumDir, "aliases"),
	}

	cutoff := time.Now().AddDate(0, 0, -7)

	for _, rootDir := range dirs {
		entries, err := os.ReadDir(rootDir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if e.IsDir() {
				itemDir := filepath.Join(rootDir, e.Name())
				files, err := os.ReadDir(itemDir)
				if err != nil {
					continue
				}

				for _, f := range files {
					if filepath.Ext(f.Name()) == ".log" {
						info, err := f.Info()
						if err != nil {
							continue
						}
						if info.ModTime().Before(cutoff) {
							os.Remove(filepath.Join(itemDir, f.Name()))
						}
					}
				}
			}
		}
	}
	return nil
}
