package finishedslide

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Read verifies the complete directory closure before returning its manifest.
// The revision digest is integrity metadata, not a signature or human approval.
func Read(root string) (Manifest, error) {
	var m Manifest
	info, err := os.Lstat(root)
	if err != nil {
		return m, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return m, fmt.Errorf("finished-slide.package_directory_invalid")
	}
	manifest := filepath.Join(root, "manifest.json")
	info, err = os.Lstat(manifest)
	if err != nil {
		return m, err
	}
	if !info.Mode().IsRegular() || info.Size() > 2<<20 {
		return m, fmt.Errorf("finished-slide.manifest_file_invalid")
	}
	f, err := os.Open(manifest)
	if err != nil {
		return m, err
	}
	data, err := io.ReadAll(io.LimitReader(f, 2<<20+1))
	closeErr := f.Close()
	if err != nil {
		return m, err
	}
	if closeErr != nil {
		return m, closeErr
	}
	if len(data) > 2<<20 {
		return m, fmt.Errorf("finished-slide.manifest_too_large")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err = d.Decode(&m); err != nil {
		return m, fmt.Errorf("finished-slide.manifest_invalid: %w", err)
	}
	if err = d.Decode(new(any)); err != io.EOF {
		return m, fmt.Errorf("finished-slide.manifest_trailing_data")
	}
	if err = m.Validate(); err != nil {
		return m, err
	}
	inventory := map[string]File{}
	for _, f := range m.Files {
		inventory[f.Path] = f
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(root, func(file string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if file == root {
			return nil
		}
		relative, err := filepath.Rel(root, file)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("finished-slide.symlink_forbidden: %s", relative)
		}
		if entry.IsDir() {
			if !portablePath(relative) {
				return fmt.Errorf("finished-slide.directory_path_invalid: %s", relative)
			}
			used := false
			for name := range inventory {
				if strings.HasPrefix(name, relative+"/") {
					used = true
					break
				}
			}
			if !used {
				return fmt.Errorf("finished-slide.unlisted_directory: %s", relative)
			}
			return nil
		}
		if relative == "manifest.json" {
			return nil
		}
		expected, ok := inventory[relative]
		if !ok {
			return fmt.Errorf("finished-slide.unlisted_file: %s", relative)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() != expected.Bytes {
			return fmt.Errorf("finished-slide.file_type_or_size_changed: %s", relative)
		}
		input, err := os.Open(file)
		if err != nil {
			return err
		}
		hash := sha256.New()
		count, err := io.Copy(hash, io.LimitReader(input, expected.Bytes+1))
		closeErr := input.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if count != expected.Bytes || fmt.Sprintf("%x", hash.Sum(nil)) != expected.SHA256 {
			return fmt.Errorf("finished-slide.file_hash_changed: %s", relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return m, err
	}
	if len(seen) != len(inventory) {
		return m, fmt.Errorf("finished-slide.dependency_missing")
	}
	return m, nil
}
