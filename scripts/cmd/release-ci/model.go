package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/buairtri/pptxgengo/internal/modelpackage"
)

type ModelEvidence struct {
	Schema        string                `json:"schema"`
	Version       string                `json:"version"`
	Commit        string                `json:"commit"`
	Archive       string                `json:"archive"`
	ArchiveSHA256 string                `json:"archive_sha256"`
	Identity      modelpackage.Identity `json:"identity"`
	Files         map[string]string     `json:"files_sha256"`
	License       string                `json:"license"`
	Source        string                `json:"source"`
}

func offlineModelEnabled() bool              { return os.Getenv("PPTXGENGO_OFFLINE_MODEL") == "true" }
func modelArchiveName(version string) string { return "pptxgengo-" + version + "-offline-model.zip" }

func assembleModel(root, out, version, commit string) error {
	if e := identity("linux-amd64", version, commit); e != nil {
		return e
	}
	report, hashes, e := modelpackage.Verify(root)
	if e != nil {
		return e
	}
	if _, e := os.Lstat(out); !os.IsNotExist(e) {
		return fmt.Errorf("model release output must be new")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		return e
	}
	name := modelArchiveName(version)
	files := map[string]Input{}
	for path := range hashes {
		files[path] = Input{Path: filepath.Join(root, path)}
	}
	// Remove runner-specific directories from the distributable metadata.
	report.Directory = "."
	raw, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	raw = append(raw, '\n')
	files["manifest.json"] = Input{Data: raw}
	hashes["manifest.json"] = fmt.Sprintf("%x", sha256.Sum256(raw))
	if e := writeArchive(files, "windows-amd64", filepath.Join(out, name)); e != nil {
		return e
	}
	archiveHash, e := digest(filepath.Join(out, name))
	if e != nil {
		return e
	}
	ev := ModelEvidence{Schema: "pptxgengo.offline-model-release/v1", Version: version, Commit: commit, Archive: name, ArchiveSHA256: archiveHash, Identity: modelpackage.PinnedIdentity(), Files: hashes, License: "Apache-2.0", Source: modelpackage.Source()}
	if e := writeJSON(filepath.Join(out, name+".evidence.json"), ev); e != nil {
		return e
	}
	return verifyModelRelease(out, version, commit)
}
func verifyModelRelease(dir, version, commit string) error {
	if e := identity("linux-amd64", version, commit); e != nil {
		return e
	}
	name := modelArchiveName(version)
	var ev ModelEvidence
	if e := readJSON(filepath.Join(dir, name+".evidence.json"), &ev); e != nil {
		return e
	}
	if ev.Schema != "pptxgengo.offline-model-release/v1" || ev.Version != version || ev.Commit != commit || ev.Archive != name || ev.License != "Apache-2.0" || ev.Source != modelpackage.Source() || !reflect.DeepEqual(ev.Identity, modelpackage.PinnedIdentity()) {
		return fmt.Errorf("model release identity mismatch")
	}
	path := filepath.Join(dir, name)
	st, e := os.Lstat(path)
	if e != nil || !st.Mode().IsRegular() || st.Size() > 100<<20 {
		return fmt.Errorf("model release archive type/size invalid")
	}
	hash, e := digest(path)
	if e != nil || hash != ev.ArchiveSHA256 {
		return fmt.Errorf("model release archive hash mismatch")
	}
	z, e := zip.OpenReader(path)
	if e != nil {
		return e
	}
	bounds, seen := modelpackage.FileBounds(), map[string]bool{}
	if len(z.File) != len(bounds) {
		z.Close()
		return fmt.Errorf("model archive inventory mismatch")
	}
	for _, f := range z.File {
		limit, ok := bounds[f.Name]
		if !ok || seen[f.Name] || !f.Mode().IsRegular() || f.UncompressedSize64 > uint64(limit) {
			z.Close()
			return fmt.Errorf("model archive member invalid")
		}
		seen[f.Name] = true
	}
	z.Close()
	stage, e := os.MkdirTemp("", "pptx-model-release-verify-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	root := filepath.Join(stage, "package")
	if e := extract(path, root); e != nil {
		return e
	}
	_, files, e := modelpackage.Verify(root)
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(files, ev.Files) {
		return fmt.Errorf("model release payload hashes mismatch")
	}
	return nil
}
func copyModelRelease(from, out, version, commit string) error {
	if e := verifyModelRelease(from, version, commit); e != nil {
		return e
	}
	name := modelArchiveName(version)
	for _, suffix := range []string{"", ".evidence.json", ".sbom.json", ".vulnerabilities.json"} {
		raw, e := os.ReadFile(filepath.Join(from, name+suffix))
		if e != nil {
			return e
		}
		f, e := os.OpenFile(filepath.Join(out, name+suffix), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if e != nil {
			return e
		}
		_, e = f.Write(raw)
		ce := f.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	return verifyModelRelease(out, version, commit)
}

// modelReleaseDescription is based on the verified manifest, not ambient flags.
func modelReleaseDescription(m Manifest) string {
	if m.OfflineModelArchive == "" {
		return ""
	}
	return "\n\nOptional offline search model: `" + m.OfflineModelArchive + "`. Platform-independent pinned MiniLM weights/tokenizer with Apache-2.0 license and source attribution, covered by the signed release manifest. Extract separately; normal searches never download files. Source-bound embeddings must be generated for the selected library."
}
