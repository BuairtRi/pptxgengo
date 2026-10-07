package finishedslide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Create materializes a new immutable revision from an already closed byte set.
// The deck publisher must validate compilation, pins and editorial facts first.
// Lifecycle/approval are explicit caller metadata; publication grants neither.
func Create(out string, m Manifest, files map[string][]byte) (Manifest, error) {
	if out == "" {
		return m, fmt.Errorf("finished-slide.output_required")
	}
	m.Files = append([]File(nil), m.Files...)
	if len(files) != len(m.Files) {
		return m, fmt.Errorf("finished-slide.inventory_content_mismatch")
	}
	for i, f := range m.Files {
		data, ok := files[f.Path]
		if !ok {
			return m, fmt.Errorf("finished-slide.content_missing: %s", f.Path)
		}
		m.Files[i].Bytes = int64(len(data))
		m.Files[i].SHA256 = sha(data)
	}
	if err := m.Seal(); err != nil {
		return m, err
	}
	dest, err := filepath.Abs(out)
	if err != nil {
		return m, err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return m, fmt.Errorf("finished-slide.output_exists: %s", dest)
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return m, err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".revision-stage-")
	if err != nil {
		return m, err
	}
	defer os.RemoveAll(stage)
	for _, f := range m.Files {
		file := filepath.Join(stage, filepath.FromSlash(f.Path))
		if err = os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			return m, err
		}
		if err = writeNew(file, files[f.Path]); err != nil {
			return m, err
		}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return m, err
	}
	if err = writeNew(filepath.Join(stage, "manifest.json"), append(data, '\n')); err != nil {
		return m, err
	}
	if _, err = Read(stage); err != nil {
		return m, err
	}
	if err = publishDirectory(stage, dest); err != nil {
		return m, fmt.Errorf("finished-slide.publish_failed: %w", err)
	}
	return m, nil
}

func writeNew(path string, data []byte) (err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
