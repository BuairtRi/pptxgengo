package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// stageNativeDeck copies the validated bundle bytes into an existing PowerPoint
// workspace. It never replaces a file already present in that workspace.
func stageNativeDeck(workspace, source string, deckBytes []byte) (string, error) {
	dir, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", fmt.Errorf("native workspace must already exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("native workspace must be an existing directory, not a symlink: %s", dir)
	}
	dest := filepath.Join(dir, filepath.Base(source))
	info, err = os.Lstat(dest)
	if err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("native workspace deck must be a regular file, not a symlink: %s", dest)
		}
		got, readErr := os.ReadFile(dest)
		if readErr != nil {
			return "", readErr
		}
		if !bytes.Equal(got, deckBytes) {
			return "", fmt.Errorf("native workspace deck already exists with different bytes: %s", dest)
		}
		return dest, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	if err = writeNew(dest, deckBytes); err != nil {
		return "", fmt.Errorf("stage native workspace deck: %w", err)
	}
	return dest, nil
}

func checkNativeDeckUnchanged(path, expectedSHA string) error {
	got, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if hash(got) != expectedSHA {
		return fmt.Errorf("native workspace deck changed during measurement: %s", path)
	}
	return nil
}
