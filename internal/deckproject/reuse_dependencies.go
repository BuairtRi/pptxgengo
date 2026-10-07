package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
)

// Candidate validation reads the exact bytes scheduled for the guarded commit.
// It never falls back to an existing file when an override is supplied.
func projectDependency(p *Project, relative string, limit int64) ([]byte, error) {
	path, err := SafePath(p.Root, relative)
	if err != nil {
		return nil, err
	}
	if raw, ok := p.sourceOverrides[filepath.ToSlash(filepath.Clean(relative))]; ok {
		if int64(len(raw)) > limit {
			return nil, fmt.Errorf("dependency exceeds size limit: %s", relative)
		}
		return raw, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("dependency must be a bounded regular file: %s", relative)
	}
	raw, err := os.ReadFile(path)
	if err == nil && int64(len(raw)) > limit {
		return nil, fmt.Errorf("dependency exceeds size limit: %s", relative)
	}
	return raw, err
}
