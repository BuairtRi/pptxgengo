package deckproject

import (
	"fmt"
	"io"
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
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("dependency must be a bounded regular file: %s", relative)
	}
	input, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, statErr := input.Stat()
	if statErr != nil || !os.SameFile(info, opened) {
		input.Close()
		return nil, fmt.Errorf("dependency changed during read: %s", relative)
	}
	raw, readErr := io.ReadAll(io.LimitReader(input, limit+1))
	after, afterErr := input.Stat()
	closeErr := input.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if afterErr != nil || int64(len(raw)) > limit || int64(len(raw)) != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return nil, fmt.Errorf("dependency changed or exceeds size limit: %s", relative)
	}
	return raw, nil
}
