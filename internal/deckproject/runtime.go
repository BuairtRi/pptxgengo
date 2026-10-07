package deckproject

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Lock struct {
	Schema           string            `json:"schema"`
	Runtime          string            `json:"runtime"`
	Go               string            `json:"go"`
	OS               string            `json:"os"`
	Architecture     string            `json:"architecture"`
	ExecutableSHA256 string            `json:"executable_sha256"`
	Engine           string            `json:"engine"`
	BundleSHA256     string            `json:"bundle_sha256"`
	BundleRevision   string            `json:"bundle_revision"`
	BundleCommit     string            `json:"bundle_commit"`
	BundleFiles      map[string]string `json:"bundle_files"`
	RuntimeFiles     map[string]string `json:"runtime_files"`
}
type Approval struct {
	ID             string            `json:"id"`
	Stage          string            `json:"stage"`
	Actor          string            `json:"actor"`
	Time           string            `json:"time"`
	ArtifactSHA256 string            `json:"artifact_sha256"`
	Slides         []string          `json:"slide_ids,omitempty"`
	Hashes         map[string]string `json:"dependency_hashes"`
	Valid          bool              `json:"valid"`
	InvalidReason  string            `json:"invalid_reason,omitempty"`
	SupersededBy   string            `json:"superseded_by,omitempty"`
}
type Invalidation struct {
	Stage      string `json:"stage"`
	SlideID    string `json:"slide_id,omitempty"`
	Reason     string `json:"reason"`
	NextAction string `json:"next_action"`
}
type State struct {
	Schema         string            `json:"schema"`
	ProjectID      string            `json:"project_id"`
	Stage          string            `json:"stage"`
	NextAction     string            `json:"next_action"`
	CurrentBuild   string            `json:"current_build,omitempty"`
	Baseline       string            `json:"baseline,omitempty"`
	ReceiptSHA256  string            `json:"receipt_sha256,omitempty"`
	SourceSHA256   string            `json:"source_sha256"`
	SemanticSHA256 string            `json:"semantic_sha256"`
	Dependencies   map[string]string `json:"dependencies"`
	Approvals      []Approval        `json:"approvals"`
	Invalidations  []Invalidation    `json:"invalidations"`
}
type Receipt struct {
	NativeLineageBuildToken string            `json:"native_lineage_build_token,omitempty"`
	Schema                  string            `json:"schema"`
	BuildID                 string            `json:"build_id"`
	ProjectID               string            `json:"project_id"`
	Created                 string            `json:"created"`
	SourceSHA256            string            `json:"source_sha256"`
	SemanticSHA256          string            `json:"semantic_sha256"`
	LockSHA256              string            `json:"lock_sha256"`
	Baseline                string            `json:"prior_baseline,omitempty"`
	AssetHashes             map[string]string `json:"asset_hashes"`
	Outputs                 map[string]string `json:"outputs"`
	Fit                     string            `json:"fit"`
	Native                  string            `json:"native"`
	Visual                  string            `json:"visual"`
	Reproducibility         string            `json:"reproducibility"`
}
type BuildOptions struct {
	Bundle string
	Engine string
}

func executableHash() (string, error) {
	path, e := os.Executable()
	if e != nil {
		return "", e
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return digest(b), nil
}
func treeHashes(root string) (map[string]string, error) {
	return filteredTreeHashes(root, nil)
}

func filteredTreeHashes(root string, excluded func(string) bool) (map[string]string, error) {
	m := map[string]string{}
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		if excluded != nil && excluded(filepath.ToSlash(rel)) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink forbidden in pinned tree: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular pinned file %s", path)
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		m[filepath.ToSlash(rel)] = digest(b)
		return nil
	})
	return m, e
}

// Discovery products and calibration are independently verified resources,
// not changes to the immutable source bundle recorded in existing locks.
func bundleAuxiliaryPath(rel string) bool {
	return rel == "catalog" || strings.HasPrefix(rel, "catalog/") ||
		rel == "library.sqlite" || rel == "typography" || strings.HasPrefix(rel, "typography/")
}

const calibrationLockKey = "typography-v2-candidate/calibration.json"

func runtimeFilePath(bundle, relative string) (string, error) {
	if relative == calibrationLockKey {
		path, err := wmdesign.CandidateCalibrationPath(filepath.Join(bundle, "fonts"))
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(filepath.Dir(bundle), path)
		if err != nil {
			return "", err
		}
		return SafePath(filepath.Dir(bundle), filepath.ToSlash(rel))
	}
	return SafePath(filepath.Dir(bundle), relative)
}
func makeLock(bundle, engine string) (Lock, error) {
	l := Lock{Schema: "pptxgengo.deck-toolchain-lock.v1", Runtime: RuntimeVersion, Go: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH, Engine: engine}
	s, e := wmdesign.Load(bundle, "")
	if e != nil {
		return l, e
	}
	if _, e = wmdesign.NewTypographyEngine(filepath.Join(bundle, "fonts"), engine); e != nil {
		return l, e
	}
	l.BundleRevision = s.Revision
	l.BundleCommit = s.Commit
	l.ExecutableSHA256, e = executableHash()
	if e != nil {
		return l, e
	}
	l.BundleFiles, e = filteredTreeHashes(bundle, bundleAuxiliaryPath)
	if e != nil {
		return l, e
	}
	l.RuntimeFiles = map[string]string{}
	if engine == wmdesign.CandidateEngine {
		relative := calibrationLockKey
		path, e := runtimeFilePath(bundle, relative)
		if e != nil {
			return l, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return l, e
		}
		l.RuntimeFiles[relative] = digest(data)
	}
	l.BundleSHA256 = digest(canonical(l.BundleFiles))
	return l, nil
}
func Pin(p *Project, bundle, engine string) (Lock, error) {
	l, e := makeLock(bundle, engine)
	if e != nil {
		return l, e
	}
	path, e := SafePath(p.Root, p.Document.Toolchain.Lockfile)
	if e != nil {
		return l, e
	}
	if e = writeExclusive(path, canonical(l), 0644); e != nil {
		return l, fmt.Errorf("pin creation requires an absent lockfile; migration must preserve prior lock explicitly: %w", e)
	}
	return l, nil
}
func ReadLock(p *Project) (Lock, []byte, error) {
	var l Lock
	path, e := SafePath(p.Root, p.Document.Toolchain.Lockfile)
	if e != nil {
		return l, nil, e
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return l, nil, e
	}
	if e = strictInto(json.RawMessage(b), &l); e != nil {
		return l, nil, e
	}
	if l.Schema != "pptxgengo.deck-toolchain-lock.v1" {
		return l, nil, fmt.Errorf("unsupported toolchain lock schema")
	}
	return l, b, nil
}
func Check(p *Project, bundle, engine string) (Compilation, error) {
	if e := ValidateEditorial(p); e != nil {
		return Compilation{}, e
	}
	l, _, e := ReadLock(p)
	if e != nil {
		return Compilation{}, e
	}
	actual, e := makeLock(bundle, engine)
	if e != nil {
		return Compilation{}, e
	}
	if !reflect.DeepEqual(l, actual) {
		return Compilation{}, fmt.Errorf("project.toolchain_drift: runtime/compiler, engine, bundle, font or asset differs from toolchain.lock.json; preserve old lock and explicitly re-pin after review")
	}
	return Compile(p, bundle, engine)
}
func writeExclusive(path string, b []byte, mode fs.FileMode) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func writeJSON(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return writeExclusive(path, append(b, '\n'), 0644)
}
func statePath(p *Project) (string, error) { return SafePath(p.Root, "state.json") }
func readState(p *Project) (State, error) {
	s := State{Schema: "pptxgengo.deck-project-state.v1", ProjectID: p.Document.ID, Stage: "intake", NextAction: "Check sources, content and template selection", Dependencies: map[string]string{}, Approvals: []Approval{}, Invalidations: []Invalidation{}}
	path, e := statePath(p)
	if e != nil {
		return s, e
	}
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		builds, pathErr := SafePath(p.Root, "builds")
		if pathErr != nil {
			return s, pathErr
		}
		entries, readErr := os.ReadDir(builds)
		if readErr == nil && len(entries) > 0 {
			return s, fmt.Errorf("project has build history but state.json is missing; restore state before resuming")
		}
		return s, nil
	}
	if e != nil {
		return s, e
	}
	if e = strictInto(json.RawMessage(b), &s); e != nil {
		return s, e
	}
	if s.Schema != "pptxgengo.deck-project-state.v1" || s.ProjectID != p.Document.ID {
		return s, fmt.Errorf("state schema/project identity mismatch")
	}
	return s, nil
}
func saveState(p *Project, s State) error {
	path, e := statePath(p)
	if e != nil {
		return e
	}
	b, e := json.MarshalIndent(s, "", "  ")
	if e != nil {
		return e
	}
	tmp := path + ".tmp-" + nonce()
	if e = writeExclusive(tmp, append(b, '\n'), 0644); e != nil {
		return e
	}
	defer os.Remove(tmp)
	return os.Rename(tmp, path)
}
func nonce() string {
	var b [8]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
func dependencies(p *Project) (map[string]string, error) {
	m := map[string]string{"deck:id": digest([]byte(p.Document.ID)), "deck:order": ""}
	composition, err := compositionPath(p)
	if err != nil {
		return nil, err
	}
	if composition != "" && p.Document.Context["composition_log"] == "" {
		path, e := SafePath(p.Root, composition)
		if e != nil {
			return nil, e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		m["context:composition_log"] = digest(data)
	}
	m["deck:copy"] = digest(canonical(map[string]any{"title": p.Document.Title, "year": p.Document.Year}))
	if len(p.Document.Sections) > 0 {
		m["deck:sections"] = digest(canonical(p.Document.Sections))
	}
	m["deck:media"] = digest(canonical(p.Document.MediaOptimization))
	order := []string{}
	for _, s := range p.Document.Slides {
		order = append(order, s.ID)
		m["slide:"+s.ID+":content"] = digest(canonical(s.Values))
		if s.Notes != "" {
			m["slide:"+s.ID+":content"] = digest(canonical(map[string]any{"values": s.Values, "notes": s.Notes}))
		}
		if s.DraftReview != nil {
			m["slide:"+s.ID+":draft-review"] = digest(canonical(s.DraftReview))
		}
		m["slide:"+s.ID+":selection"] = digest(canonical(s.Template))
		if s.Hidden {
			m["slide:"+s.ID+":selection"] = digest(canonical(map[string]any{"template": s.Template, "hidden": true}))
		}
		m["slide:"+s.ID+":evidence"] = digest(canonical(s.EvidenceRefs))
		for field, relative := range map[string]string{"source": p.SlideFiles[s.ID], "notes-file": p.NotesFiles[s.ID]} {
			if relative != "" {
				path, e := SafePath(p.Root, relative)
				if e != nil {
					return nil, e
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					return nil, e
				}
				m["slide:"+s.ID+":"+field] = digest(raw)
			}
		}
	}
	m["deck:order"] = digest(canonical(order))
	for k, t := range p.Document.LocalTemplates {
		m["template:"+k] = digest(canonical(t))
		if relative := p.TemplateFiles[k]; relative != "" {
			path, e := SafePath(p.Root, relative)
			if e != nil {
				return nil, e
			}
			raw, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			m["template:"+k+":source"] = digest(raw)
		}
		if t.Provenance != nil && t.Provenance.DefinitionSnapshot != "" {
			path, e := SafePath(p.Root, t.Provenance.DefinitionSnapshot)
			if e != nil {
				return nil, e
			}
			data, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			m["template:"+k+":ancestor"] = digest(data)
		}
	}
	for id, a := range p.Document.Assets {
		if a.DerivationReceipt != "" {
			path, e := SafePath(p.Root, a.DerivationReceipt)
			if e != nil {
				return nil, e
			}
			bytes, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			m["asset:"+id+":derivation"] = digest(bytes)
		}
		if a.Path != "" {
			path, e := SafePath(p.Root, a.Path)
			if e != nil {
				return nil, e
			}
			b, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			m["asset:"+id] = digest(b)
		} else {
			m["asset:"+id] = digest(canonical(a))
		}
	}
	for k, rel := range p.Document.Context {
		if k == "state" && filepath.Clean(rel) == "state.json" {
			continue
		}
		path, e := SafePath(p.Root, rel)
		if e != nil {
			return nil, e
		}
		st, e := os.Stat(path)
		if e != nil {
			return nil, e
		}
		if st.IsDir() {
			hashes, e := treeHashes(path)
			if e != nil {
				return nil, e
			}
			m["context:"+k] = digest(canonical(hashes))
		} else {
			b, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			m["context:"+k] = digest(b)
		}
	}
	lock, _, e := ReadLock(p)
	if e != nil {
		return nil, e
	}
	m["toolchain"] = digest(canonical(lock))
	return m, nil
}
func affectedKeys(p *Project, stage string, slides []string, deps map[string]string) map[string]string {
	m := map[string]string{}
	selected := map[string]bool{}
	for _, id := range slides {
		selected[id] = true
	}
	all := len(slides) == 0
	for k, v := range deps {
		include := false
		switch {
		case k == "deck:order", k == "deck:copy", k == "deck:sections":
			include = true
		case strings.HasPrefix(k, "context:"):
			include = true
		case k == "toolchain", k == "deck:media", strings.HasPrefix(k, "asset:"):
			include = stage == "selection" || stage == "build" || stage == "review" || stage == "delivered"
		case strings.HasPrefix(k, "slide:"):
			parts := strings.Split(k, ":")
			include = (all || selected[parts[1]]) && (parts[2] != "selection" || stage == "selection" || stage == "build" || stage == "review" || stage == "delivered")
		case strings.HasPrefix(k, "template:"):
			if stage != "selection" && stage != "build" && stage != "review" && stage != "delivered" {
				break
			}
			if all {
				include = true
			} else {
				for _, s := range p.Document.Slides {
					if selected[s.ID] && s.Template.Scope == "local" && (k == "template:"+s.Template.ID || strings.HasPrefix(k, "template:"+s.Template.ID+":")) {
						include = true
					}
				}
			}
		}
		if include {
			m[k] = v
		}
	}
	return m
}
func Status(p *Project) (State, error) {
	s, e := readState(p)
	if e != nil {
		return s, e
	}
	deps, e := dependencies(p)
	if e != nil {
		return s, e
	}
	s.Invalidations = []Invalidation{}
	for i := range s.Approvals {
		a := &s.Approvals[i]
		if a.SupersededBy != "" {
			a.Valid = false
			a.InvalidReason = "superseded by " + a.SupersededBy
			continue
		}
		current := affectedKeys(p, a.Stage, a.Slides, deps)
		a.Valid = reflect.DeepEqual(a.Hashes, current)
		a.InvalidReason = ""
		if !a.Valid {
			a.InvalidReason = "approved dependency hashes changed"
			s.Invalidations = append(s.Invalidations, Invalidation{Stage: a.Stage, Reason: a.InvalidReason, NextAction: "Review changed content or composition and approve its new hash"})
		}
	}
	if s.CurrentBuild != "" && !reflect.DeepEqual(s.Dependencies, deps) {
		s.Invalidations = append(s.Invalidations, Invalidation{Stage: "build", Reason: "build inputs changed", NextAction: "Build changed inputs in a new immutable directory"})
	}
	if s.SemanticSHA256 != "" && s.SemanticSHA256 != digest(p.Canonical) {
		s.Invalidations = append(s.Invalidations, Invalidation{Stage: "build", Reason: "authored source changed", NextAction: "Build current source in a new immutable directory"})
	}
	if e = protectBaseline(p, s); e != nil {
		s.Invalidations = append(s.Invalidations, Invalidation{Stage: "review", Reason: e.Error(), NextAction: "Preserve edited PPTX as a separate artifact; reconcile manually before building or delivery"})
		s.NextAction = "Resolve generated baseline divergence"
		return s, nil
	}
	if len(s.Invalidations) > 0 {
		s.Stage = "content"
		s.NextAction = s.Invalidations[0].NextAction
	}
	return s, nil
}
func Resume(p *Project) (State, error) {
	s, e := Status(p)
	if e != nil {
		return s, e
	}
	if e = saveState(p, s); e != nil {
		return s, e
	}
	return s, nil
}
func Approve(p *Project, stage, actor string, slides []string) (Approval, error) {
	a := Approval{}
	allowed := map[string]bool{"intake": true, "framing": true, "outline": true, "content": true, "selection": true, "build": true, "review": true, "delivered": true}
	if !allowed[stage] || strings.TrimSpace(actor) == "" {
		return a, fmt.Errorf("approval requires a known stage and named actor")
	}
	known := map[string]bool{}
	for _, s := range p.Document.Slides {
		known[s.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range slides {
		if !known[id] || seen[id] {
			return a, fmt.Errorf("invalid/duplicate approval slide ID %s", id)
		}
		seen[id] = true
	}
	s, e := Status(p)
	if e != nil {
		return a, e
	}
	if e = protectBaseline(p, s); e != nil {
		return a, e
	}
	if (stage == "build" || stage == "review" || stage == "delivered") && (s.CurrentBuild == "" || s.SemanticSHA256 != digest(p.Canonical)) {
		return a, fmt.Errorf("this approval requires a current generated build")
	}
	deps, e := dependencies(p)
	if e != nil {
		return a, e
	}
	if (stage == "build" || stage == "review" || stage == "delivered") && !reflect.DeepEqual(deps, s.Dependencies) {
		return a, fmt.Errorf("approval requires build inputs unchanged since current build; rebuild before approving")
	}
	if stage == "build" || stage == "review" || stage == "delivered" {
		lock, _, e := ReadLock(p)
		if e != nil {
			return a, e
		}
		hash, e := executableHash()
		if e != nil {
			return a, e
		}
		if hash != lock.ExecutableSHA256 {
			return a, fmt.Errorf("approval runtime differs from pinned build runtime")
		}
	}
	hashes := affectedKeys(p, stage, slides, deps)
	a = Approval{ID: "approval-" + nonce(), Stage: stage, Actor: actor, Time: time.Now().UTC().Format(time.RFC3339), ArtifactSHA256: digest(canonical(hashes)), Slides: slides, Hashes: hashes, Valid: true}
	path, e := SafePath(p.Root, "reviews/"+a.ID+".json")
	if e != nil {
		return a, e
	}
	if e = writeJSON(path, a); e != nil {
		return a, e
	}
	for i := range s.Approvals {
		old := &s.Approvals[i]
		if old.Stage == stage && sameScope(old.Slides, slides) {
			old.SupersededBy = a.ID
			old.Valid = false
			old.InvalidReason = "superseded by " + a.ID
		}
	}
	s.Approvals = append(s.Approvals, a)
	s.Stage = stage
	s.NextAction = "Continue with the next unresolved stage; run project status before resuming"
	if e = saveState(p, s); e != nil {
		return a, e
	}
	return a, nil
}
func protectBaseline(p *Project, s State) error {
	if s.CurrentBuild == "" {
		return nil
	}
	if !stableID.MatchString(s.CurrentBuild) {
		return fmt.Errorf("invalid build ID in state")
	}
	dir, e := SafePath(p.Root, "builds/"+s.CurrentBuild)
	if e != nil {
		return e
	}
	raw, e := os.ReadFile(filepath.Join(dir, "receipt.json"))
	if e != nil {
		return e
	}
	if digest(raw) != s.ReceiptSHA256 {
		return fmt.Errorf("project.baseline_divergence: receipt changed")
	}
	var r Receipt
	if e = strictInto(json.RawMessage(raw), &r); e != nil {
		return e
	}
	for rel, want := range r.Outputs {
		path, e := SafePath(dir, rel)
		if e != nil {
			return e
		}
		data, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if digest(data) != want {
			return fmt.Errorf("project.baseline_divergence: %s changed; manual edits require reconciliation", rel)
		}
	}
	return nil
}
func Build(p *Project, opts BuildOptions) (Receipt, error) {
	r := Receipt{}
	guard, e := SafePath(p.Root, ".project-build.lock")
	if e != nil {
		return r, e
	}
	if e = writeExclusive(guard, []byte(time.Now().UTC().Format(time.RFC3339)), 0600); e != nil {
		return r, fmt.Errorf("project busy (remove stale .project-build.lock only after confirming no build runs): %w", e)
	}
	defer os.Remove(guard)
	s, e := Status(p)
	if e != nil {
		return r, e
	}
	if e = protectBaseline(p, s); e != nil {
		return r, e
	}
	beforeDeps, e := dependencies(p)
	if e != nil {
		return r, e
	}
	c, e := Check(p, opts.Bundle, opts.Engine)
	if e != nil {
		return r, e
	}
	pptxBytes, report, e := wmdesign.BuildWithEngineAndAssets(opts.Bundle, "", c.Document, opts.Engine, c.Assets)
	if e != nil {
		return r, e
	}
	lock, lockBytes, e := ReadLock(p)
	if e != nil {
		return r, e
	}
	_ = lock
	id := "build-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce()
	dir, e := SafePath(p.Root, "builds/"+id)
	if e != nil {
		return r, e
	}
	if e = os.MkdirAll(filepath.Dir(dir), 0755); e != nil {
		return r, e
	}
	if e = os.Mkdir(dir, 0755); e != nil {
		return r, e
	}
	success := false
	defer func() {
		if !success {
			os.RemoveAll(dir)
		}
	}()
	objects, e := ObjectMap(p, c.Document, pptxBytes)
	if e != nil {
		return r, e
	}
	pptxBytes, objects, e = StampNativeLineage(pptxBytes, objects, digest(lockBytes))
	if e != nil {
		return r, e
	}
	outputs := map[string][]byte{"deck.pptx": pptxBytes, "deck.yaml": p.Raw, "source.canonical.json": p.Canonical, "scene.json": canonical(c.Document), "layout-report.json": canonical(report), "object-map.json": canonical(objects), "toolchain.lock.json": lockBytes}
	for relative, raw := range p.SourceFiles {
		if p.hasExternalSources() {
			outputs["authored/"+relative] = raw
		}
	}
	r = Receipt{NativeLineageBuildToken: objects.Lineage.BuildToken, Schema: "pptxgengo.deck-build-receipt.v1", BuildID: id, ProjectID: p.Document.ID, Created: time.Now().UTC().Format(time.RFC3339), SourceSHA256: p.SourceHash(), SemanticSHA256: digest(p.Canonical), LockSHA256: digest(lockBytes), Baseline: s.CurrentBuild, AssetHashes: c.AssetHashes, Outputs: map[string]string{}, Fit: "compiler_checks_passed_arbitrary_content_unqualified", Native: "not_reviewed", Visual: "not_reviewed", Reproducibility: "fixed build timestamp 2000-01-01T00:00:00Z; identity seed is canonical authored source SHA256; native lineage generation token pins source, lock and unstamped PPTX; execution receipt time is independent"}
	keys := []string{}
	for k := range outputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if e = writeExclusive(filepath.Join(dir, k), outputs[k], 0444); e != nil {
			return r, e
		}
		r.Outputs[k] = digest(outputs[k])
	}
	if e = writeJSON(filepath.Join(dir, "receipt.json"), r); e != nil {
		return r, e
	}
	if e = os.Chmod(filepath.Join(dir, "receipt.json"), 0444); e != nil {
		return r, e
	}
	current, e := Load(p.SourcePath)
	if e != nil {
		return r, e
	}
	if current.SourceHash() != p.SourceHash() {
		return r, fmt.Errorf("source changed during build; retry")
	}
	currentLock, e := os.ReadFile(filepath.Join(p.Root, p.Document.Toolchain.Lockfile))
	if e != nil || !bytesEqual(currentLock, lockBytes) {
		return r, fmt.Errorf("toolchain lock changed during build")
	}
	receiptBytes, e := os.ReadFile(filepath.Join(dir, "receipt.json"))
	if e != nil {
		return r, e
	}
	s.CurrentBuild = id
	s.Baseline = id
	s.ReceiptSHA256 = digest(receiptBytes)
	s.SourceSHA256 = p.SourceHash()
	s.SemanticSHA256 = digest(p.Canonical)
	s.Stage = "build"
	s.NextAction = "Review actual deck fit and native rendering; approval is separate from compiler success"
	s.Dependencies, e = dependencies(p)
	if e != nil {
		return r, e
	}
	if !reflect.DeepEqual(beforeDeps, s.Dependencies) {
		return r, fmt.Errorf("project dependencies changed during build; retry")
	}
	if e = protectBaseline(p, State{CurrentBuild: r.Baseline, ReceiptSHA256: func() string { old, _ := readState(p); return old.ReceiptSHA256 }()}); e != nil {
		return r, e
	}
	if e = saveState(p, s); e != nil {
		return r, e
	}
	success = true
	return r, nil
}
func bytesEqual(a, b []byte) bool { return string(a) == string(b) }

func sameScope(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aa := append([]string{}, a...)
	bb := append([]string{}, b...)
	sort.Strings(aa)
	sort.Strings(bb)
	return reflect.DeepEqual(aa, bb)
}
