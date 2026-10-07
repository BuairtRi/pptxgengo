package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ShareReceipt struct {
	Schema                  string                  `json:"schema"`
	CurrentVersion          string                  `json:"current_version"`
	SHA256                  string                  `json:"sha256"`
	Files                   map[string]string       `json:"files"`
	ContainsPrivateMaterial bool                    `json:"contains_private_material"`
	AssetAliases            map[string]VersionAsset `json:"asset_aliases,omitempty"`
}

func ShareProject(p *Project, out string) (ShareReceipt, error) {
	r := ShareReceipt{Schema: "pptxgengo.project-share.v1", Files: map[string]string{}, ContainsPrivateMaterial: true}
	if out == "" {
		return r, fmt.Errorf("share requires a new ZIP path")
	}
	list, e := ListVersions(p.Root)
	if e != nil {
		return r, e
	}
	if list.Current.Number == "" {
		return r, fmt.Errorf("save a numbered complete version before sharing")
	}
	r.CurrentVersion = list.Current.Number
	files, e := projectInventory(p.Root, ignoredPortable)
	if e != nil {
		return r, e
	}
	original := hashBytes(files)
	for rel, b := range files {
		r.Files[rel] = digest(b)
	}
	files["SHARE.md"] = []byte("# Complete private deck project\n\nUse `pptxdesign project share-extract --archive package.zip --out new-project` for a complete working project. Manual unzip can open versioned PowerPoint decks, but legacy asset aliases must be expanded before rebuilding. Asset originals with duplicate hashes are stored once in this ZIP and expanded deterministically to ordinary relative files for legacy path compatibility; no filesystem links. The working deck is deck.yaml; ordered slides are in slides/, local templates in slides/templates/, and deck-owned immutable assets in assets/. versions/current.json identifies the latest committed source-and-deck snapshot. Each version has a browsable deck.pptx and manifest.json. Asset bytes are shared by SHA256, never symlinks. Native working copies, receipts, approval history, contexts and evidence are private material.\n\nRun `pptxdesign project share-verify --project .` after synchronization/extraction. Resolve conflicting copies explicitly without deleting either collaborator's work. OneDrive is not a transaction service; interrupted publications are rejected. Materialize a saved version into a NEW folder with `pptxdesign project version materialize --project . --number 000001 --out ../restored-project`.\n\nToolchain executable/OS/architecture and template pins remain exact. Use the originally pinned runtime to rebuild, or explicitly migrate the toolchain using project migrate; sharing does not authorize a new runtime. Shared branding resources and fonts are still required unless an offline runtime package is also supplied. PowerPoint can open versions/<number>/deck.pptx without the CLI; native render quality and font installation remain separate qualification.\n")
	files["SHARE.md"] = append(files["SHARE.md"], []byte("\nCurrent saved PowerPoint: ["+r.CurrentVersion+"](versions/"+r.CurrentVersion+"/deck.pptx).\n")...)
	r.Files["SHARE.md"] = digest(files["SHARE.md"])
	delete(r.Files, "share-manifest.json")
	r.AssetAliases = map[string]VersionAsset{}
	assetPaths := map[string]bool{}
	for _, asset := range p.Document.Assets {
		if asset.Path != "" {
			assetPaths[asset.Path] = true
		}
	}
	for _, version := range list.Versions {
		for relative := range version.Assets {
			assetPaths[relative] = true
		}
	}
	for rel, b := range files {
		if strings.HasPrefix(rel, "assets/") || assetPaths[rel] {
			sha := digest(b)
			object := "assets/objects/sha256/" + sha
			if rel != object {
				r.AssetAliases[rel] = VersionAsset{Object: object, SHA256: sha}
				delete(files, rel)
				existing, e := readProjectFile(p.Root, object)
				if e != nil || !bytes.Equal(existing, b) {
					return r, fmt.Errorf("shared asset object unavailable: %s; save a new version first", rel)
				}
			}
		}
	}
	sizes := map[string]uint64{}
	for relative, b := range files {
		sizes[relative] = uint64(len(b))
	}
	if e = validateShareInventory(r, sizes); e != nil {
		return r, e
	}
	manifest := canonical(r)
	if uint64(len(manifest)) > portableManifestLimit {
		return r, fmt.Errorf("share manifest exceeds16MiB")
	}
	files["share-manifest.json"] = manifest
	abs, e := filepath.Abs(out)
	if e != nil {
		return r, e
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(abs))
	if e != nil {
		return r, e
	}
	abs = filepath.Join(parent, filepath.Base(abs))
	rootAbs, _ := filepath.EvalSymlinks(p.Root)
	if strings.EqualFold(abs, rootAbs) || strings.HasPrefix(strings.ToLower(abs), strings.ToLower(rootAbs)+string(filepath.Separator)) {
		return r, fmt.Errorf("share output must be outside project")
	}
	f, e := os.OpenFile(abs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return r, e
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(abs)
		}
	}()
	z := zip.NewWriter(f)
	for _, name := range sortedFileKeys(files) {
		if _, e := SafePath(p.Root, name); e != nil {
			return r, e
		}
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		h.SetMode(portableFileMode(name))
		w, e := z.CreateHeader(h)
		if e != nil {
			return r, e
		}
		if _, e = w.Write(files[name]); e != nil {
			return r, e
		}
	}
	if e = z.Close(); e != nil {
		return r, e
	}
	if e = f.Close(); e != nil {
		return r, e
	} // Detect changes during read/ZIP construction, including new files.
	current, e := projectInventory(p.Root, ignoredPortable)
	if e != nil {
		return r, e
	}
	if !reflectEqual(original, hashBytes(current)) {
		return r, fmt.Errorf("project changed during share; ZIP discarded")
	}
	// Stream the final ZIP hash; never allocate another archive-sized buffer.
	info, e := os.Stat(abs)
	if e != nil {
		return r, e
	}
	r.SHA256, _, _, e = hashReconcileFile(abs, info.Size(), false)
	if e != nil {
		return r, e
	}
	ok = true
	return r, nil
}
func hashBytes(files map[string][]byte) map[string]string {
	m := map[string]string{}
	for rel, b := range files {
		m[rel] = digest(b)
	}
	return m
}
func VerifyShare(root string) (ShareReceipt, error) {
	r := ShareReceipt{}
	if e := refuseSyncConflicts(root); e != nil {
		return r, e
	}
	raw, e := readProjectFileLimit(root, "share-manifest.json", portableManifestLimit)
	if e != nil {
		return r, e
	}
	if e = strictInto(json.RawMessage(raw), &r); e != nil {
		return r, e
	}
	if r.Schema != "pptxgengo.project-share.v1" || !r.ContainsPrivateMaterial {
		return r, fmt.Errorf("invalid private share manifest")
	}
	for rel, a := range r.AssetAliases {
		if a.Object != "assets/objects/sha256/"+a.SHA256 || r.Files[rel] != a.SHA256 || r.Files[a.Object] != a.SHA256 {
			return r, fmt.Errorf("invalid shared asset alias %s", rel)
		}
	}
	for rel, want := range r.Files {
		b, e := readProjectFile(root, rel)
		if e != nil {
			return r, e
		}
		if !shaPattern.MatchString(want) || digest(b) != want {
			return r, fmt.Errorf("shared project drift: %s", rel)
		}
	}
	actual, e := projectInventory(root, ignoredPortable)
	if e != nil {
		return r, e
	}
	delete(actual, "share-manifest.json")
	if !reflectEqual(hashBytes(actual), r.Files) {
		return r, fmt.Errorf("shared project inventory changed; preserve new or conflicting files separately")
	}
	list, e := ListVersions(root)
	if e != nil {
		return r, e
	}
	if list.Current.Number != r.CurrentVersion {
		return r, fmt.Errorf("share/current version disagreement")
	}
	return r, nil
}

// ExtractShare expands transport aliases to ordinary relative files in a NEW
// directory. ZIP storage retains one object per asset hash; old authored path
// contracts can require physical copies after extraction, never symlinks.
func ExtractShare(archive, out string) (ShareReceipt, error) {
	r := ShareReceipt{}
	if archive == "" || out == "" {
		return r, fmt.Errorf("share-extract requires archive and new output directory")
	}
	z, e := zip.OpenReader(archive)
	if e != nil {
		return r, e
	}
	defer z.Close()
	if len(z.File) > portableFileCountLimit {
		return r, fmt.Errorf("share archive has too many files")
	}
	entries := map[string]*zip.File{}
	folded := map[string]bool{}
	var total uint64
	for _, f := range z.File {
		if e = portableName(f.Name); e != nil {
			return r, e
		}
		if filepath.IsAbs(f.Name) || f.Mode()&os.ModeSymlink != 0 || f.FileInfo().IsDir() || !f.Mode().IsRegular() {
			return r, fmt.Errorf("share archive contains unsafe entry %s", f.Name)
		}
		key := strings.ToLower(f.Name)
		if folded[key] {
			return r, fmt.Errorf("share archive contains duplicate/case-colliding path %s", f.Name)
		}
		folded[key] = true
		total += f.UncompressedSize64
		if f.UncompressedSize64 > uint64(portableFileLimit) || total > portableTotalLimit {
			return r, fmt.Errorf("share archive exceeds extraction limits")
		}
		entries[f.Name] = f
	}
	read := func(name string) ([]byte, error) {
		f, exists := entries[name]
		if !exists {
			return nil, fmt.Errorf("share archive missing %s", name)
		}
		stream, e := f.Open()
		if e != nil {
			return nil, e
		}
		defer stream.Close()
		b, e := io.ReadAll(io.LimitReader(stream, int64(f.UncompressedSize64)+1))
		if e != nil {
			return nil, e
		}
		if uint64(len(b)) != f.UncompressedSize64 {
			return nil, fmt.Errorf("ZIP size mismatch: %s", name)
		}
		return b, nil
	}
	manifestEntry, exists := entries["share-manifest.json"]
	if !exists || manifestEntry.UncompressedSize64 > portableManifestLimit {
		return r, fmt.Errorf("share manifest missing or exceeds 16MiB")
	}
	manifest, e := read("share-manifest.json")
	if e != nil {
		return r, e
	}
	if e = strictInto(json.RawMessage(manifest), &r); e != nil {
		return r, e
	}
	if r.Schema != "pptxgengo.project-share.v1" || !r.ContainsPrivateMaterial {
		return r, fmt.Errorf("invalid private share archive manifest")
	}
	sizes := map[string]uint64{}
	for name, f := range entries {
		sizes[name] = f.UncompressedSize64
	}
	if e = validateShareInventory(r, sizes); e != nil {
		return r, e
	}
	for rel, a := range r.AssetAliases {
		if e = portableName(rel); e != nil {
			return r, e
		}
		if a.Object != "assets/objects/sha256/"+a.SHA256 || r.Files[rel] != a.SHA256 || r.Files[a.Object] != a.SHA256 {
			return r, fmt.Errorf("invalid shared asset alias %s", rel)
		}
		if _, exists := entries[rel]; exists {
			return r, fmt.Errorf("aliased asset redundantly stored: %s", rel)
		}
	}
	for name := range entries {
		if name != "share-manifest.json" && r.Files[name] == "" {
			return r, fmt.Errorf("unmanifested ZIP entry %s", name)
		}
	}
	for rel, want := range r.Files {
		if e = portableName(rel); e != nil {
			return r, e
		}
		if !shaPattern.MatchString(want) {
			return r, fmt.Errorf("invalid share file hash %s", rel)
		}
		if _, alias := r.AssetAliases[rel]; alias {
			continue
		}
		b, e := read(rel)
		if e != nil {
			return r, e
		}
		if digest(b) != want {
			return r, fmt.Errorf("share archive hash mismatch %s", rel)
		}
	}
	abs, e := filepath.Abs(out)
	if e != nil {
		return r, e
	}
	if e = os.Mkdir(abs, 0755); e != nil {
		return r, e
	}
	success := false
	defer func() {
		if !success {
			os.RemoveAll(abs)
		}
	}()
	for name := range entries {
		b, e := read(name)
		if e != nil {
			return r, e
		}
		path, e := SafePath(abs, name)
		if e != nil {
			return r, e
		}
		if e = writeExclusive(path, b, portableFileMode(name)); e != nil {
			return r, e
		}
	}
	for rel, a := range r.AssetAliases {
		b, e := readProjectFile(abs, a.Object)
		if e != nil {
			return r, e
		}
		if digest(b) != a.SHA256 {
			return r, fmt.Errorf("shared object drift")
		}
		path, e := SafePath(abs, rel)
		if e != nil {
			return r, e
		}
		if e = writeExclusive(path, b, portableFileMode(rel)); e != nil {
			return r, e
		}
	}
	if _, e = VerifyShare(abs); e != nil {
		return r, e
	}
	if _, e = Load(abs); e != nil {
		return r, fmt.Errorf("extracted working project invalid: %w", e)
	}
	success = true
	return r, nil
}

// validateShareInventory checks the LOGICAL expansion, not just compressed or
// physical ZIP entries. Run before any output creation or asset expansion.
func validateShareInventory(r ShareReceipt, sizes map[string]uint64) error {
	if len(r.Files) > portableFileCountLimit || len(r.AssetAliases) > portableFileCountLimit {
		return fmt.Errorf("share logical inventory exceeds100000 files/aliases")
	}
	names := make([]string, 0, len(r.Files))
	folded := map[string]string{}
	var total uint64
	for relative, want := range r.Files {
		if e := portableName(relative); e != nil {
			return e
		}
		if !shaPattern.MatchString(want) || relative == "share-manifest.json" {
			return fmt.Errorf("invalid logical share file %s", relative)
		}
		key := strings.ToLower(relative)
		if prior, exists := folded[key]; exists {
			return fmt.Errorf("logical share path/case collision: %s and %s", prior, relative)
		}
		folded[key] = relative
		names = append(names, key)
		source := relative
		if alias, exists := r.AssetAliases[relative]; exists {
			source = alias.Object
			if alias.Object != "assets/objects/sha256/"+alias.SHA256 || alias.SHA256 != want || r.Files[source] != want {
				return fmt.Errorf("invalid asset alias %s", relative)
			}
			if _, exists = sizes[relative]; exists {
				return fmt.Errorf("duplicate aliased asset storage %s", relative)
			}
		}
		size, exists := sizes[source]
		if !exists || size > uint64(portableFileLimit) {
			return fmt.Errorf("missing/oversized logical share file %s", relative)
		}
		if total > portableTotalLimit-size {
			return fmt.Errorf("expanded share exceeds8GiB")
		}
		total += size
	}
	folded["share-manifest.json"] = "share-manifest.json"
	for _, name := range names {
		parts := strings.Split(name, "/")
		for i := 1; i < len(parts); i++ {
			if _, exists := folded[strings.Join(parts[:i], "/")]; exists {
				return fmt.Errorf("logical share file/directory prefix collision")
			}
		}
	}
	for alias := range r.AssetAliases {
		if _, exists := r.Files[alias]; !exists {
			return fmt.Errorf("unmanifested asset alias %s", alias)
		}
	}
	return nil
}
