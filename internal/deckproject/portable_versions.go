package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const versionSchema = "pptxgengo.deck-version.v1"
const portableFileLimit int64 = 512 << 20
const portableTotalLimit uint64 = 8 << 30
const portableFileCountLimit = 100000
const portableManifestLimit uint64 = 16 << 20

var versionName = regexp.MustCompile(`^[0-9]{6}$`)

type VersionAsset struct {
	Object string `json:"object"`
	SHA256 string `json:"sha256"`
}
type DeckVersion struct {
	Schema         string                  `json:"schema"`
	Number         string                  `json:"number"`
	ProjectID      string                  `json:"project_id"`
	Created        string                  `json:"created"`
	Actor          string                  `json:"actor"`
	Message        string                  `json:"message"`
	Parent         string                  `json:"parent,omitempty"`
	ParentSHA256   string                  `json:"parent_sha256,omitempty"`
	SourceSHA256   string                  `json:"source_sha256"`
	SemanticSHA256 string                  `json:"semantic_sha256"`
	BuildID        string                  `json:"build_id"`
	DeckSHA256     string                  `json:"deck_sha256"`
	Files          map[string]string       `json:"files"`
	Assets         map[string]VersionAsset `json:"assets"`
	SharedBuilds   map[string]string       `json:"shared_builds"`
}
type VersionPointer struct {
	Schema         string `json:"schema"`
	Number         string `json:"number"`
	ManifestSHA256 string `json:"manifest_sha256"`
}
type VersionList struct {
	Schema   string         `json:"schema"`
	Current  VersionPointer `json:"current"`
	Versions []DeckVersion  `json:"versions"`
}

func readOptional(path string) ([]byte, error) {
	return readOptionalLimit(path, uint64(portableFileLimit))
}
func readOptionalLimit(path string, limit uint64) ([]byte, error) {
	_, _, b, e := hashReconcileFile(path, int64(limit), true)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e == nil && b == nil {
		b = []byte{}
	}
	return b, e
}
func readProjectFile(root, relative string) ([]byte, error) {
	return readProjectFileLimit(root, relative, uint64(portableFileLimit))
}
func readProjectFileLimit(root, relative string, limit uint64) ([]byte, error) {
	path, e := SafePath(root, relative)
	if e != nil {
		return nil, e
	}
	file, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer file.Close()
	info, e := file.Stat()
	if e != nil {
		return nil, e
	}
	if !info.Mode().IsRegular() || info.Size() < 0 || uint64(info.Size()) > limit {
		return nil, fmt.Errorf("nonregular/oversized project file %s", relative)
	}
	bytes, e := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if e != nil {
		return nil, e
	}
	if uint64(len(bytes)) > limit {
		return nil, fmt.Errorf("project file grew beyond size limit: %s", relative)
	}
	return bytes, nil
}
func portableName(relative string) error {
	for _, part := range strings.Split(relative, "/") {
		for _, r := range part {
			if r < 32 {
				return fmt.Errorf("nonportable control character in path %q", relative)
			}
		}
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, "\\:<>\"|?*\x00") || strings.TrimRight(part, " .") != part {
			return fmt.Errorf("nonportable path %q", relative)
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return fmt.Errorf("Windows-reserved project path %q", relative)
		}
	}
	return nil
}
func projectInventory(root string, exclude func(string, bool) bool) (map[string][]byte, error) {
	m := map[string][]byte{}
	names := map[string]string{}
	var total uint64
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("portable project refuses symlink %s", rel)
		}
		if exclude != nil && exclude(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if e := portableName(rel); e != nil {
			return e
		}
		folded := strings.ToLower(rel)
		if prior, exists := names[folded]; exists && prior != rel {
			return fmt.Errorf("case-colliding portable paths: %s and %s", prior, rel)
		}
		names[folded] = rel
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() || info.Size() > portableFileLimit {
			return fmt.Errorf("nonregular/oversized project file %s (limit512MiB)", rel)
		}
		if uint64(info.Size()) > portableTotalLimit-total || len(m) >= portableFileCountLimit {
			return fmt.Errorf("portable project inventory exceeds 8GiB or100000 files")
		}
		limit := portableTotalLimit - total
		if limit > uint64(portableFileLimit) {
			limit = uint64(portableFileLimit)
		}
		b, e := readProjectFileLimit(root, rel, limit)
		if e != nil {
			return e
		}
		total += uint64(len(b))
		m[rel] = b
		return nil
	})
	return m, e
}
func ignoredPortable(rel string, isDir bool) bool {
	first := strings.Split(rel, "/")[0]
	return first == ".git" || first == ".cache" || first == ".slotctl" || strings.HasPrefix(first, ".project-") || first == ".deck-source-mutation.lock" || strings.HasSuffix(rel, ".icloud") || strings.HasPrefix(filepath.Base(rel), "~$")
}
func conflictName(rel string) bool {
	s := strings.ToLower(filepath.Base(rel))
	return strings.Contains(s, "conflicted copy") || strings.Contains(s, "conflict copy") || strings.Contains(s, "(conflict")
}
func refuseSyncConflicts(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if rel == ".git" || rel == ".cache" || rel == ".slotctl" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(rel, ".icloud") || conflictName(rel) {
			return fmt.Errorf("unresolved synchronization conflict or placeholder: %s; preserve both copies and resolve explicitly", rel)
		}
		return nil
	})
}

func versionManifest(root, number string) (DeckVersion, []byte, error) {
	v := DeckVersion{}
	if !versionName.MatchString(number) {
		return v, nil, fmt.Errorf("version must be six digits")
	}
	raw, e := readProjectFileLimit(root, "versions/"+number+"/manifest.json", portableManifestLimit)
	if e != nil {
		return v, nil, e
	}
	if uint64(len(raw)) > portableManifestLimit {
		return v, nil, fmt.Errorf("version manifest exceeds 16MiB")
	}
	if e = strictInto(json.RawMessage(raw), &v); e != nil {
		return v, nil, e
	}
	if v.Schema != versionSchema || v.Number != number || !stableID.MatchString(v.ProjectID) || !shaPattern.MatchString(v.SourceSHA256) || !shaPattern.MatchString(v.SemanticSHA256) || !shaPattern.MatchString(v.DeckSHA256) || v.Actor == "" || v.Message == "" {
		return v, nil, fmt.Errorf("invalid version manifest %s", number)
	}
	return v, raw, nil
}
func VerifyVersion(root, number string) (DeckVersion, error) {
	v, _, e := versionManifest(root, number)
	if e != nil {
		return v, e
	}
	sourceRoot, e := SafePath(root, "versions/"+number+"/source")
	if e != nil {
		return v, e
	}
	sourceFiles, e := projectInventory(sourceRoot, nil)
	if e != nil {
		return v, e
	}
	if !reflectEqual(hashBytes(sourceFiles), v.Files) {
		return v, fmt.Errorf("immutable version source inventory changed: %s", number)
	}
	snapshotRoot, e := SafePath(root, "versions/"+number)
	if e != nil {
		return v, e
	}
	snapshotFiles, e := projectInventory(snapshotRoot, nil)
	if e != nil {
		return v, e
	}
	wantSnapshot := map[string]string{"deck.pptx": v.DeckSHA256}
	_, manifestBytes, e := versionManifest(root, number)
	if e != nil {
		return v, e
	}
	wantSnapshot["manifest.json"] = digest(manifestBytes)
	for relative, want := range v.Files {
		wantSnapshot["source/"+relative] = want
	}
	if !reflectEqual(hashBytes(snapshotFiles), wantSnapshot) {
		return v, fmt.Errorf("immutable version directory inventory changed: %s", number)
	}
	check := func(rel, want string) error {
		if !shaPattern.MatchString(want) {
			return fmt.Errorf("invalid version hash for %s", rel)
		}
		b, e := readProjectFile(root, rel)
		if e != nil {
			return e
		}
		if digest(b) != want {
			return fmt.Errorf("immutable version drift: %s", rel)
		}
		return nil
	}
	for rel, want := range v.Files {
		if _, e := SafePath(root, rel); e != nil {
			return v, e
		}
		if e = check("versions/"+number+"/source/"+rel, want); e != nil {
			return v, e
		}
	}
	for rel, a := range v.Assets {
		if _, e := SafePath(root, rel); e != nil {
			return v, e
		}
		if a.Object != "assets/objects/sha256/"+a.SHA256 {
			return v, fmt.Errorf("invalid asset object %s", a.Object)
		}
		if e = check(a.Object, a.SHA256); e != nil {
			return v, e
		}
	}
	for rel, want := range v.SharedBuilds {
		if !strings.HasPrefix(rel, "builds/") {
			return v, fmt.Errorf("invalid shared build path")
		}
		if e = check(rel, want); e != nil {
			return v, e
		}
	}
	if e = check("versions/"+number+"/deck.pptx", v.DeckSHA256); e != nil {
		return v, e
	}
	if v.Files["deck.yaml"] == "" || v.Files["state.json"] == "" || v.SharedBuilds["builds/"+v.BuildID+"/receipt.json"] == "" {
		return v, fmt.Errorf("incomplete version source/build snapshot")
	}
	return v, nil
}
func ListVersions(root string) (VersionList, error) {
	r := VersionList{Schema: "pptxgengo.deck-versions.v1", Versions: []DeckVersion{}}
	if e := refuseSyncConflicts(root); e != nil {
		return r, e
	}
	entries, e := os.ReadDir(filepath.Join(root, "versions"))
	if os.IsNotExist(e) {
		return r, nil
	}
	if e != nil {
		return r, e
	}
	for _, entry := range entries {
		if entry.Name() == "current.json" {
			continue
		}
		if !entry.IsDir() || !versionName.MatchString(entry.Name()) {
			return r, fmt.Errorf("unrecognized/interrupted version entry %s; retain it for explicit recovery", entry.Name())
		}
		v, e := VerifyVersion(root, entry.Name())
		if e != nil {
			return r, e
		}
		r.Versions = append(r.Versions, v)
	}
	sort.Slice(r.Versions, func(i, j int) bool { return r.Versions[i].Number < r.Versions[j].Number })
	pointerPath, e := SafePath(root, "versions/current.json")
	if e != nil {
		return r, e
	}
	raw, e := readOptionalLimit(pointerPath, portableManifestLimit)
	if e != nil {
		return r, e
	}
	if len(r.Versions) == 0 {
		if raw != nil {
			return r, fmt.Errorf("current version exists without snapshots")
		}
		return r, nil
	}
	if raw == nil {
		return r, fmt.Errorf("version snapshots exist without current pointer; retain interrupted publication")
	}
	if e = strictInto(json.RawMessage(raw), &r.Current); e != nil {
		return r, e
	}
	if r.Current.Schema != "pptxgengo.deck-version-pointer.v1" {
		return r, fmt.Errorf("invalid version pointer")
	}
	last := r.Versions[len(r.Versions)-1]
	_, lastRaw, e := versionManifest(root, last.Number)
	if e != nil {
		return r, e
	}
	if r.Current.Number != last.Number || r.Current.ManifestSHA256 != digest(lastRaw) {
		return r, fmt.Errorf("current version pointer diverges from latest immutable snapshot")
	}
	for i, v := range r.Versions {
		if v.Number != fmt.Sprintf("%06d", i+1) {
			return r, fmt.Errorf("version sequence is not contiguous at %s", v.Number)
		}
		if i == 0 {
			if v.Parent != "" || v.ParentSHA256 != "" {
				return r, fmt.Errorf("first version has unexpected predecessor")
			}
		} else {
			prior := r.Versions[i-1]
			_, b, e := versionManifest(root, prior.Number)
			if e != nil {
				return r, e
			}
			if v.Parent != prior.Number || v.ParentSHA256 != digest(b) {
				return r, fmt.Errorf("version predecessor divergence at %s", v.Number)
			}
		}
	}
	return r, nil
}

// SaveVersion stores asset bytes once per deck, independently of source versions.
// Number allocation and pointer publication occur under an exclusive owned lock.
// An interrupted publication is retained and diagnosed, never silently adopted.
func SaveVersion(p *Project, actor, message string) (DeckVersion, error) {
	v := DeckVersion{}
	if strings.TrimSpace(actor) == "" || strings.TrimSpace(message) == "" {
		return v, fmt.Errorf("version save requires actor and message")
	}
	if filepath.Base(p.SourcePath) != "deck.yaml" {
		return v, fmt.Errorf("portable versions require deck.yaml")
	}
	guard, e := SafePath(p.Root, ".project-version.lock")
	if e != nil {
		return v, e
	}
	if e = writeExclusive(guard, []byte(p.SourceHash()), 0600); e != nil {
		return v, fmt.Errorf("version publication busy; do not force lock takeover: %w", e)
	}
	defer os.Remove(guard)
	for _, rel := range []string{".project-build.lock", ".deck-source-mutation.lock"} {
		path, e := SafePath(p.Root, rel)
		if e != nil {
			return v, e
		}
		if b, e := readOptionalLimit(path, portableManifestLimit); e != nil || b != nil {
			return v, fmt.Errorf("project busy: %s; no forced lock takeover", rel)
		}
	}
	currentProject, e := Load(p.SourcePath)
	if e != nil {
		return v, e
	}
	if currentProject.SourceHash() != p.SourceHash() {
		return v, fmt.Errorf("authored source changed before version publication")
	}
	list, e := ListVersions(p.Root)
	if e != nil {
		return v, e
	}
	state, e := Status(p)
	if e != nil {
		return v, e
	}
	if e = protectBaseline(p, state); e != nil {
		return v, e
	}
	deps, e := dependencies(p)
	if e != nil {
		return v, e
	}
	if state.CurrentBuild == "" || state.SourceSHA256 != p.SourceHash() || state.SemanticSHA256 != digest(p.Canonical) || !reflectEqual(deps, state.Dependencies) {
		return v, fmt.Errorf("version save requires a current generated build; rebuild changed source/dependencies")
	}
	files, e := projectInventory(p.Root, func(rel string, dir bool) bool {
		return ignoredPortable(rel, dir) || rel == "versions" || rel == "assets" || rel == "builds"
	})
	if e != nil {
		return v, e
	}
	assets, e := projectInventory(p.Root, func(rel string, dir bool) bool {
		return ignoredPortable(rel, dir) || (!strings.HasPrefix(rel, "assets/") && rel != "assets")
	})
	if e != nil {
		return v, e
	}
	for _, a := range p.Document.Assets {
		if a.Path != "" {
			b, e := readProjectFile(p.Root, a.Path)
			if e != nil {
				return v, e
			}
			if a.SHA256 != "" && digest(b) != a.SHA256 {
				return v, fmt.Errorf("asset pin drift: %s", a.Path)
			}
			assets[a.Path] = b
		}
	}
	for relative := range assets {
		delete(files, relative)
	}
	builds, e := projectInventory(p.Root, func(rel string, dir bool) bool {
		return ignoredPortable(rel, dir) || (!strings.HasPrefix(rel, "builds/") && rel != "builds")
	})
	if e != nil {
		return v, e
	}
	var captureBytes uint64
	for _, group := range []map[string][]byte{files, assets, builds} {
		for _, b := range group {
			captureBytes += uint64(len(b))
		}
	}
	if captureBytes > portableTotalLimit {
		return v, fmt.Errorf("complete snapshot input exceeds 8GiB")
	}
	number := 1
	if list.Current.Number != "" {
		n, _ := strconv.Atoi(list.Current.Number)
		number = n + 1
	}
	if number > 999999 {
		return v, fmt.Errorf("version sequence exhausted")
	}
	v = DeckVersion{Schema: versionSchema, Number: fmt.Sprintf("%06d", number), ProjectID: p.Document.ID, Created: time.Now().UTC().Format(time.RFC3339Nano), Actor: actor, Message: message, Parent: list.Current.Number, ParentSHA256: list.Current.ManifestSHA256, SourceSHA256: p.SourceHash(), SemanticSHA256: digest(p.Canonical), BuildID: state.CurrentBuild, Files: map[string]string{}, Assets: map[string]VersionAsset{}, SharedBuilds: map[string]string{}}
	deck, e := readProjectFile(p.Root, "builds/"+state.CurrentBuild+"/deck.pptx")
	if e != nil {
		return v, e
	}
	v.DeckSHA256 = digest(deck)
	for rel, b := range assets {
		sha := digest(b)
		object := "assets/objects/sha256/" + sha
		path, e := SafePath(p.Root, object)
		if e != nil {
			return v, e
		}
		prior, e := readOptional(path)
		if e != nil {
			return v, e
		}
		if prior == nil {
			if e = writeExclusive(path, b, 0444); e != nil {
				return v, e
			}
		} else if !bytes.Equal(prior, b) {
			return v, fmt.Errorf("immutable asset object drift: %s", object)
		}
		v.Assets[rel] = VersionAsset{Object: object, SHA256: sha}
	}
	for rel, b := range files {
		v.Files[rel] = digest(b)
	}
	for rel, b := range builds {
		v.SharedBuilds[rel] = digest(b)
	}
	for rel := range builds {
		parts := strings.Split(rel, "/")
		if len(parts) < 3 || builds["builds/"+parts[1]+"/receipt.json"] == nil {
			return v, fmt.Errorf("incomplete retained build directory: %s", rel)
		}
	}
	for rel, b := range builds {
		if strings.HasSuffix(rel, "/receipt.json") {
			var receipt Receipt
			if e := strictInto(json.RawMessage(b), &receipt); e != nil {
				return v, e
			}
			prefix := strings.TrimSuffix(rel, "receipt.json")
			for output, want := range receipt.Outputs {
				data, exists := builds[prefix+output]
				if !exists || digest(data) != want {
					return v, fmt.Errorf("immutable retained build drift: %s", prefix+output)
				}
			}
		}
	}
	dir, e := SafePath(p.Root, "versions/"+v.Number)
	if e != nil {
		return v, e
	}
	if e = os.MkdirAll(filepath.Dir(dir), 0755); e != nil {
		return v, e
	}
	if e = os.Mkdir(dir, 0755); e != nil {
		return v, e
	}
	for _, rel := range sortedFileKeys(files) {
		if e = writeVersionFile(p.Root, v.Number, "source/"+rel, files[rel]); e != nil {
			return v, e
		}
	}
	if e = writeVersionFile(p.Root, v.Number, "deck.pptx", deck); e != nil {
		return v, e
	}
	manifest := canonical(v)
	if uint64(len(manifest)) > portableManifestLimit {
		return v, fmt.Errorf("version manifest exceeds 16MiB; snapshot retained")
	}
	if e = writeVersionFile(p.Root, v.Number, "manifest.json", manifest); e != nil {
		return v, e
	} // Retain any interrupted snapshot for explicit diagnosis.
	// Verify every captured predecessor again, including source, native edits,
	// state, build receipts and original assets. OneDrive does not honor locks.
	for rel, b := range files {
		actual, e := readProjectFile(p.Root, rel)
		if e != nil || !bytes.Equal(actual, b) {
			return v, fmt.Errorf("project changed during version publication: %s; snapshot retained", rel)
		}
	}
	for rel, b := range assets {
		actual, e := readProjectFile(p.Root, rel)
		if e != nil || !bytes.Equal(actual, b) {
			return v, fmt.Errorf("asset changed during version publication: %s; snapshot retained", rel)
		}
	}
	for rel, b := range builds {
		actual, e := readProjectFile(p.Root, rel)
		if e != nil || !bytes.Equal(actual, b) {
			return v, fmt.Errorf("build changed during version publication: %s; snapshot retained", rel)
		}
	}
	nowFiles, e := projectInventory(p.Root, func(rel string, dir bool) bool {
		return ignoredPortable(rel, dir) || rel == "versions" || rel == "assets" || rel == "builds"
	})
	if e != nil {
		return v, e
	}
	for relative := range assets {
		delete(nowFiles, relative)
	}
	if !reflectEqual(hashBytes(nowFiles), hashBytes(files)) {
		return v, fmt.Errorf("project inventory changed during publication; snapshot retained")
	}
	nowBuilds, e := projectInventory(p.Root, func(rel string, dir bool) bool {
		return ignoredPortable(rel, dir) || (!strings.HasPrefix(rel, "builds/") && rel != "builds")
	})
	if e != nil {
		return v, e
	}
	if !reflectEqual(hashBytes(nowBuilds), hashBytes(builds)) {
		return v, fmt.Errorf("build inventory changed during publication; snapshot retained")
	}
	currentProject, e = Load(p.SourcePath)
	if e != nil {
		return v, e
	}
	if currentProject.SourceHash() != p.SourceHash() {
		return v, fmt.Errorf("authored source changed during capture; snapshot retained")
	}
	currentDeps, e := dependencies(currentProject)
	if e != nil {
		return v, e
	}
	if !reflectEqual(currentDeps, state.Dependencies) {
		return v, fmt.Errorf("build dependencies changed during capture; snapshot retained")
	}
	pointer := VersionPointer{Schema: "pptxgengo.deck-version-pointer.v1", Number: v.Number, ManifestSHA256: digest(manifest)}
	dest := filepath.Join(p.Root, "versions/current.json")
	before, e := readOptionalLimit(dest, portableManifestLimit)
	if e != nil {
		return v, e
	}
	want := []byte(nil)
	if list.Current.Number != "" {
		want = canonical(list.Current)
	}
	if !bytes.Equal(bytes.TrimSpace(before), bytes.TrimSpace(want)) {
		return v, fmt.Errorf("current pointer changed during publication; snapshot retained")
	}
	tmp := filepath.Join(p.Root, "versions/.pointer-"+nonce())
	if e = writeExclusive(tmp, canonical(pointer), 0644); e != nil {
		return v, e
	}
	defer os.Remove(tmp)
	if e = os.Rename(tmp, dest); e != nil {
		return v, e
	}
	return v, nil
}

func MaterializeVersion(root, number, out string) (DeckVersion, error) {
	v, e := VerifyVersion(root, number)
	if e != nil {
		return v, e
	}
	abs, e := filepath.Abs(out)
	if e != nil {
		return v, e
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(abs))
	if e != nil {
		return v, e
	}
	abs = filepath.Join(parent, filepath.Base(abs))
	rootAbs, e := filepath.EvalSymlinks(root)
	if e != nil {
		return v, e
	}
	if strings.EqualFold(abs, rootAbs) || strings.HasPrefix(strings.ToLower(abs), strings.ToLower(rootAbs)+string(filepath.Separator)) {
		return v, fmt.Errorf("materialize requires a new directory outside the source project")
	}
	if e = os.Mkdir(abs, 0755); e != nil {
		return v, e
	}
	// Failed output is retained for explicit recovery, including any concurrent
	// colleague edits; never recursively delete synchronized project data.
	expected := map[string]string{}
	copy := func(from, to, want string) error {
		b, e := readVersionInput(root, from, want)
		if e != nil {
			return e
		}
		path, e := SafePath(abs, to)
		if e != nil {
			return e
		}
		if prior, exists := expected[to]; exists {
			if prior != want {
				return fmt.Errorf("conflicting materialized destination: %s", to)
			}
			return nil
		}
		expected[to] = want
		return writeExclusive(path, b, portableFileMode(to))
	}
	for rel, want := range v.Files {
		if e = copy("versions/"+number+"/source/"+rel, rel, want); e != nil {
			return v, e
		}
	}
	for rel, a := range v.Assets {
		if e = copy(a.Object, rel, a.SHA256); e != nil {
			return v, e
		}
	}
	for rel, want := range v.SharedBuilds {
		if e = copy(rel, rel, want); e != nil {
			return v, e
		}
	}
	materialized, e := projectInventory(abs, nil)
	if e != nil {
		return v, e
	}
	if !reflectEqual(hashBytes(materialized), expected) {
		return v, fmt.Errorf("materialized complete inventory changed")
	}
	p, e := Load(abs)
	if e != nil {
		return v, e
	}
	if p.SourceHash() != v.SourceSHA256 || digest(p.Canonical) != v.SemanticSHA256 {
		return v, fmt.Errorf("materialized source pins differ")
	}
	s, e := Status(p)
	if e != nil {
		return v, e
	}
	if e = protectBaseline(p, s); e != nil {
		return v, e
	}
	if e = VerifyMaterializedAssets(abs, v); e != nil {
		return v, e
	}
	return v, nil
}
func VerifyMaterializedAssets(root string, v DeckVersion) error {
	for rel, a := range v.Assets {
		b, e := readProjectFile(root, rel)
		if e != nil {
			return e
		}
		if digest(b) != a.SHA256 {
			return fmt.Errorf("materialized asset drift: %s", rel)
		}
	}
	return nil
}

// RecoverVersion publishes only an already complete, verified immediate child.
// It never changes authored files, deletes a conflicting snapshot or takes over
// another writer's lock. The exact pointer preimage must be explicitly supplied.
func RecoverVersion(root, number, expect string) (VersionPointer, error) {
	r := VersionPointer{}
	if e := refuseSyncConflicts(root); e != nil {
		return r, e
	}
	guard, e := SafePath(root, ".project-version.lock")
	if e != nil {
		return r, e
	}
	if e = writeExclusive(guard, []byte("explicit recovery"), 0600); e != nil {
		return r, fmt.Errorf("version publication busy; no lock takeover: %w", e)
	}
	defer os.Remove(guard)
	v, raw, e := versionManifest(root, number)
	if e != nil {
		return r, e
	}
	if _, e = VerifyVersion(root, number); e != nil {
		return r, e
	}
	dest, e := SafePath(root, "versions/current.json")
	if e != nil {
		return r, e
	}
	before, e := readOptionalLimit(dest, portableManifestLimit)
	if e != nil {
		return r, e
	}
	if (before == nil && expect != "absent") || (before != nil && digest(before) != expect) {
		return r, fmt.Errorf("recovery requires exact current-pointer SHA256, or absent")
	}
	prior := VersionPointer{}
	if before != nil {
		if e = strictInto(json.RawMessage(before), &prior); e != nil {
			return r, e
		}
	}
	if v.Parent != prior.Number || v.ParentSHA256 != prior.ManifestSHA256 {
		return r, fmt.Errorf("recovery snapshot is not the immediate child of current pointer")
	}
	entries, e := os.ReadDir(filepath.Join(root, "versions"))
	if e != nil {
		return r, e
	}
	for _, entry := range entries {
		if entry.Name() == "current.json" {
			continue
		}
		if !entry.IsDir() || !versionName.MatchString(entry.Name()) || entry.Name() > number {
			return r, fmt.Errorf("unresolved/interrupted version entry: %s", entry.Name())
		}
		if _, e = VerifyVersion(root, entry.Name()); e != nil {
			return r, e
		}
	}
	ordered := []DeckVersion{}
	for _, entry := range entries {
		if entry.IsDir() {
			candidate, _, e := versionManifest(root, entry.Name())
			if e != nil {
				return r, e
			}
			ordered = append(ordered, candidate)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })
	for i, candidate := range ordered {
		if candidate.Number != fmt.Sprintf("%06d", i+1) {
			return r, fmt.Errorf("recovery cannot bridge missing version history")
		}
		if i == 0 {
			if candidate.Parent != "" || candidate.ParentSHA256 != "" {
				return r, fmt.Errorf("first snapshot has invalid predecessor")
			}
		} else {
			_, b, e := versionManifest(root, ordered[i-1].Number)
			if e != nil {
				return r, e
			}
			if candidate.Parent != ordered[i-1].Number || candidate.ParentSHA256 != digest(b) {
				return r, fmt.Errorf("recovery predecessor divergence")
			}
		}
	}
	r = VersionPointer{Schema: "pptxgengo.deck-version-pointer.v1", Number: number, ManifestSHA256: digest(raw)}
	current, e := readOptionalLimit(dest, portableManifestLimit)
	if e != nil || !bytes.Equal(current, before) {
		return r, fmt.Errorf("pointer changed during recovery")
	}
	tmp, e := SafePath(root, "versions/.pointer-"+nonce())
	if e != nil {
		return r, e
	}
	if e = writeExclusive(tmp, canonical(r), 0644); e != nil {
		return r, e
	}
	defer os.Remove(tmp)
	if e = os.Rename(tmp, dest); e != nil {
		return r, e
	}
	return r, nil
}

func writeVersionFile(root, number, relative string, b []byte) error {
	path, e := SafePath(root, "versions/"+number+"/"+relative)
	if e != nil {
		return e
	}
	return writeExclusive(path, b, 0444)
}

func portableFileMode(relative string) os.FileMode {
	parts := strings.Split(relative, "/")
	if strings.HasPrefix(relative, "assets/objects/sha256/") || strings.HasPrefix(relative, "builds/") || strings.HasPrefix(relative, "decisions/sources/") || (len(parts) >= 3 && parts[0] == "versions" && versionName.MatchString(parts[1])) {
		return 0444
	}
	return 0644
}

func readVersionInput(root, relative, want string) ([]byte, error) {
	bytes, e := readProjectFile(root, relative)
	if e != nil {
		return nil, e
	}
	if digest(bytes) != want {
		return nil, fmt.Errorf("immutable input changed during materialization: %s", relative)
	}
	return bytes, nil
}
