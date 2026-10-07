package finishedslide

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadPayload materializes the exact closed revision within integration bounds.
// It verifies closure and revision before and after bounded regular-file reads.
func ReadPayload(root string, m Manifest) (map[string][]byte, error) {
	if e := m.Validate(); e != nil {
		return nil, e
	}
	var size int64
	for _, f := range m.Files {
		size += f.Bytes
		if f.Bytes > 64<<20 || size > 512<<20 {
			return nil, fmt.Errorf("finished-slide.payload_package_bounds")
		}
	}
	current, e := Read(root)
	if e != nil {
		return nil, e
	}
	if current.RevisionSHA256 != m.RevisionSHA256 {
		return nil, fmt.Errorf("finished-slide.payload_revision_mismatch")
	}
	files := map[string][]byte{}
	for _, f := range m.Files {
		file := filepath.Join(root, filepath.FromSlash(f.Path))
		before, e := os.Lstat(file)
		if e != nil {
			return nil, e
		}
		if !before.Mode().IsRegular() || before.Size() != f.Bytes {
			return nil, fmt.Errorf("finished-slide.payload_file_changed")
		}
		input, e := os.Open(file)
		if e != nil {
			return nil, e
		}
		opened, statErr := input.Stat()
		if statErr != nil || !os.SameFile(before, opened) {
			input.Close()
			return nil, fmt.Errorf("finished-slide.payload_file_changed")
		}
		raw, readErr := io.ReadAll(io.LimitReader(input, f.Bytes+1))
		after, afterErr := input.Stat()
		closeErr := input.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if afterErr != nil || int64(len(raw)) != f.Bytes || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) || sha(raw) != f.SHA256 {
			return nil, fmt.Errorf("finished-slide.payload_file_changed")
		}
		files[f.Path] = raw
	}
	current, e = Read(root)
	if e != nil || current.RevisionSHA256 != m.RevisionSHA256 {
		return nil, fmt.Errorf("finished-slide.payload_package_changed")
	}
	return files, nil
}
