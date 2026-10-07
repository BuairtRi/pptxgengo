package modelpackage

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

//go:embed LICENSE.model.txt
var license []byte

const Schema = "pptxgengo.offline-embedding-package.v1"

func Source() string  { return "https://huggingface.co/" + ModelID + "/tree/" + ModelRevision }
func License() []byte { return append([]byte(nil), license...) }
func Readme() []byte {
	return []byte("Offline sentence-transformers/all-MiniLM-L6-v2\nModel source: " + Source() + "\nLicense: Apache-2.0 (see LICENSE).\nCreated by explicit pptxdesign library-model maintenance.\nWeights/tokenizer hashes, runtime, pooling and revision are in manifest.json.\nNormal library searches never download files.\n")
}

type Report struct {
	Schema    string   `json:"schema"`
	Directory string   `json:"directory"`
	Identity  Identity `json:"identity"`
	Source    string   `json:"source"`
	License   string   `json:"license"`
}

func fileHash(root, name string, limit int64) (string, int64, error) {
	path := filepath.Join(root, name)
	before, e := os.Lstat(path)
	if e != nil {
		return "", 0, e
	}
	if !before.Mode().IsRegular() || before.Size() > limit {
		return "", 0, fmt.Errorf("model package requires bounded regular file: %s", name)
	}
	input, e := os.Open(path)
	if e != nil {
		return "", 0, e
	}
	opened, e := input.Stat()
	if e != nil || !os.SameFile(before, opened) {
		input.Close()
		return "", 0, fmt.Errorf("model package file changed: %s", name)
	}
	h := sha256.New()
	count, readErr := io.Copy(h, io.LimitReader(input, limit+1))
	after, statErr := input.Stat()
	closeErr := input.Close()
	if readErr != nil {
		return "", 0, readErr
	}
	if closeErr != nil {
		return "", 0, closeErr
	}
	if statErr != nil || count != before.Size() || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return "", 0, fmt.Errorf("model package file changed: %s", name)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), count, nil
}

// FileBounds returns the closed root inventory and maximum bytes for each file.
func FileBounds() map[string]int64 {
	expected := map[string]int64{"manifest.json": 2 << 20, "LICENSE": int64(len(license)), "README.txt": int64(len(Readme()))}
	for _, a := range Artifacts() {
		expected[a.File] = a.Bytes
	}
	return expected
}

// Verify checks the closed package against compiled-in pins, including license.
// No model execution, network requests or trust in caller-provided artifact hashes.
func Verify(root string) (Report, map[string]string, error) {
	var report Report
	hashes := map[string]string{}
	st, e := os.Lstat(root)
	if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return report, nil, fmt.Errorf("model package directory invalid")
	}
	expected := FileBounds()
	entries, e := os.ReadDir(root)
	if e != nil {
		return report, nil, e
	}
	if len(entries) != len(expected) {
		return report, nil, fmt.Errorf("model package closure mismatch")
	}
	for _, entry := range entries {
		limit, ok := expected[entry.Name()]
		if !ok {
			return report, nil, fmt.Errorf("unexpected model package file: %s", entry.Name())
		}
		hash, _, e := fileHash(root, entry.Name(), limit)
		if e != nil {
			return report, nil, e
		}
		hashes[entry.Name()] = hash
	}
	manifest, e := os.Open(filepath.Join(root, "manifest.json"))
	if e != nil {
		return report, nil, e
	}
	raw, readErr := io.ReadAll(io.LimitReader(manifest, (2<<20)+1))
	closeErr := manifest.Close()
	if readErr != nil {
		return report, nil, readErr
	}
	if closeErr != nil {
		return report, nil, closeErr
	}
	if e != nil || int64(len(raw)) > expected["manifest.json"] || fmt.Sprintf("%x", sha256.Sum256(raw)) != hashes["manifest.json"] {
		return report, nil, fmt.Errorf("model manifest changed")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e = d.Decode(&report); e != nil {
		return report, nil, e
	}
	if e = d.Decode(new(any)); e != io.EOF {
		return report, nil, fmt.Errorf("model manifest trailing data")
	}
	if report.Schema != Schema || report.Source != Source() || report.License != "Apache-2.0" || !reflect.DeepEqual(report.Identity, PinnedIdentity()) {
		return report, nil, fmt.Errorf("model package identity mismatch")
	}
	for _, a := range Artifacts() {
		if hashes[a.File] != a.SHA256 {
			return report, nil, fmt.Errorf("model artifact pin mismatch: %s", a.File)
		}
	}
	for name, raw := range map[string][]byte{"LICENSE": License(), "README.txt": Readme()} {
		if hashes[name] != fmt.Sprintf("%x", sha256.Sum256(raw)) {
			return report, nil, fmt.Errorf("model attribution changed: %s", name)
		}
	}
	return report, hashes, nil
}
