package installstate

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	Root     string
	BinDir   string
	SkillDir string
	Launcher string
}
type Selection struct {
	Package     Package `json:"package"`
	Release     string  `json:"release"`
	SkillDir    string  `json:"skill_directory,omitempty"`
	SkillSHA256 string  `json:"skill_sha256,omitempty"`
	ActivatedAt string  `json:"activated_at"`
}
type OriginalSkill struct {
	Backup      string `json:"backup,omitempty"`
	Hash        string `json:"hash,omitempty"`
	ManagedHash string `json:"managed_hash"`
}
type State struct {
	Schema         string                   `json:"schema"`
	Current        *Selection               `json:"current"`
	Previous       *Selection               `json:"previous,omitempty"`
	OriginalSkills map[string]OriginalSkill `json:"original_skills,omitempty"`
}
type LinkChange struct {
	Path   string  `json:"path"`
	Before *string `json:"before"`
	After  string  `json:"after"`
}
type pathSnapshot struct {
	Value  string `json:"value"`
	Kind   uint32 `json:"kind"`
	Exists bool   `json:"exists"`
}
type transaction struct {
	Schema          string       `json:"schema"`
	Before          State        `json:"before"`
	After           State        `json:"after"`
	SkillPath       string       `json:"skill_path,omitempty"`
	SkillStage      string       `json:"skill_stage,omitempty"`
	SkillBackup     string       `json:"skill_backup,omitempty"`
	BeforeSkillHash string       `json:"before_skill_hash,omitempty"`
	AfterSkillHash  string       `json:"after_skill_hash,omitempty"`
	Links           []LinkChange `json:"links,omitempty"`
	ChangePath      bool         `json:"change_path"`
	BeforePath      pathSnapshot `json:"before_path,omitempty"`
	AfterPath       pathSnapshot `json:"after_path,omitempty"`
}
type Options struct {
	SkipSkill bool
	NoPath    bool
	StageOnly bool
}

func Defaults() (Config, error) {
	home, e := os.UserHomeDir()
	if e != nil {
		return Config{}, e
	}
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		data = filepath.Join(home, ".local", "share")
	}
	if runtime.GOOS == "windows" {
		data = os.Getenv("LOCALAPPDATA")
		if data == "" {
			return Config{}, fmt.Errorf("LOCALAPPDATA is required")
		}
	}
	root := filepath.Join(data, "pptxgengo")
	bin := filepath.Join(home, ".local", "bin")
	if runtime.GOOS == "windows" {
		bin = filepath.Join(root, "bin")
	}
	codex := os.Getenv("CODEX_HOME")
	if codex == "" {
		codex = filepath.Join(home, ".codex")
	}
	exe, e := os.Executable()
	if e != nil {
		return Config{}, e
	}
	exe, e = filepath.EvalSymlinks(exe)
	if e != nil {
		return Config{}, e
	}
	if hint := os.Getenv("PPTXGENGO_INSTALLATION_ROOT"); hint != "" {
		root = hint
	}
	managedRoot := filepath.Dir(filepath.Dir(exe))
	if _, e = os.Lstat(filepath.Join(managedRoot, "launchers.json")); e == nil {
		root = managedRoot
	}
	if runtime.GOOS == "windows" {
		bin = filepath.Join(root, "bin")
	}
	return Config{Root: root, BinDir: bin, SkillDir: filepath.Join(codex, "skills", "west-monroe-presentations"), Launcher: exe}, nil
}
func (c Config) normalized() (Config, error) {
	for _, field := range []*string{&c.Root, &c.BinDir, &c.Launcher} {
		if *field == "" {
			return c, fmt.Errorf("installation paths must be explicit")
		}
		p, e := canonicalPath(*field)
		if e != nil {
			return c, e
		}
		*field = filepath.Clean(p)
	}
	if c.SkillDir == "" {
		return c, fmt.Errorf("skill directory must be explicit")
	}
	skill, e := filepath.Abs(c.SkillDir)
	if e != nil {
		return c, e
	}
	parent, e := canonicalPath(filepath.Dir(skill))
	if e != nil {
		return c, e
	}
	// Preserve the final skill entry: an existing user symlink must be backed up
	// as a link, never resolved and replaced at its external target.
	c.SkillDir = filepath.Join(parent, filepath.Base(skill))
	overlaps := func(a, b string) bool { return a == b || within(a, b) || within(b, a) }
	if overlaps(c.Root, c.SkillDir) || overlaps(c.BinDir, c.SkillDir) {
		return c, fmt.Errorf("skill directory must be separate from installation state and launchers")
	}
	if runtime.GOOS != "windows" && overlaps(c.Root, c.BinDir) {
		return c, fmt.Errorf("Unix launcher directory must be separate from installation state")
	}
	return c, nil
}

// Resolve the existing ancestor too: macOS /var is an alias for /private/var,
// and user data roots may have other directory aliases. Store one physical path.
func canonicalPath(path string) (string, error) {
	absolute, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	parent := filepath.Clean(absolute)
	missing := []string{}
	for {
		if _, e = os.Lstat(parent); e == nil {
			break
		} else if !os.IsNotExist(e) {
			return "", e
		}
		next := filepath.Dir(parent)
		if next == parent {
			return "", fmt.Errorf("path has no existing ancestor")
		}
		missing = append(missing, filepath.Base(parent))
		parent = next
	}
	resolved, e := filepath.EvalSymlinks(parent)
	if e != nil {
		return "", e
	}
	for i := len(missing) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, missing[i])
	}
	return resolved, nil
}
func token() string {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func atomicJSON(path string, value any) error {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	tmp := path + ".tmp-" + token()
	f, e := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer os.Remove(tmp)
	if _, e = f.Write(append(b, '\n')); e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	return replaceFile(tmp, path)
}
func within(root, path string) bool {
	rel, e := filepath.Rel(root, path)
	return e == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func regularDirectory(path string) error {
	info, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("not a real directory: %s", path)
	}
	return nil
}
func loadState(root string) (State, error) {
	var e error
	root, e = canonicalPath(root)
	if e != nil {
		return State{}, e
	}
	s := State{Schema: "pptxgengo.installation/v1"}
	e = readJSON(filepath.Join(root, "active.json"), &s)
	if os.IsNotExist(e) {
		return s, nil
	}
	if e != nil {
		return s, e
	}
	if s.Schema != "pptxgengo.installation/v1" {
		return s, fmt.Errorf("unsupported installation state")
	}
	for _, v := range []*Selection{s.Current, s.Previous} {
		if v != nil && (v.Release != filepath.Join(root, "releases", v.Package.Version) || !versionPattern.MatchString(v.Package.Version)) {
			return s, fmt.Errorf("active release escapes managed releases")
		}
	}
	return s, nil
}
func (c Config) mutate(work func(Config) error) error {
	var e error
	c, e = c.normalized()
	if e != nil {
		return e
	}
	if e = os.MkdirAll(c.Root, 0700); e != nil {
		return e
	}
	if e = regularDirectory(c.Root); e != nil {
		return e
	}
	unlock, e := lockRoot(c.Root)
	if e != nil {
		return fmt.Errorf("another installer is active: %w", e)
	}
	defer unlock()
	if _, e = os.Lstat(filepath.Join(c.Root, "pending.json")); e == nil {
		if e = c.recover(); e != nil {
			return fmt.Errorf("unfinished activation needs recovery: %w", e)
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	return work(c)
}

// Install stages and verifies a package before selecting it. Retrying an already
// staged identical package is safe; an existing version with other bytes is refused.
func (c Config) Install(ctx context.Context, source string, o Options) (*Selection, error) {
	var selected *Selection
	e := c.mutate(func(c Config) error {
		source, e := canonicalPath(source)
		if e != nil {
			return e
		}
		p, e := Verify(source)
		if e != nil {
			return e
		}
		parent := filepath.Join(c.Root, "releases")
		if e = os.MkdirAll(parent, 0700); e != nil {
			return e
		}
		if e = regularDirectory(parent); e != nil {
			return e
		}
		dest := filepath.Join(parent, p.Version)
		if _, e = os.Lstat(dest); os.IsNotExist(e) {
			if within(source, parent) || within(source, dest) {
				return fmt.Errorf("installation destination must be outside the source package")
			}
			stage := filepath.Join(parent, ".stage-"+token())
			defer os.RemoveAll(stage)
			if e = copyTree(source, stage); e != nil {
				return e
			}
			checked, e := Verify(stage)
			if e != nil {
				return e
			}
			if checked.ManifestSHA256 != p.ManifestSHA256 || checked.ContentSHA256 != p.ContentSHA256 {
				return fmt.Errorf("source manifest changed while copying")
			}
			if e = probe(ctx, stage, p); e != nil {
				return e
			}
			if e = publishDirectory(stage, dest); e != nil {
				return e
			}
		} else if e != nil {
			return e
		} else {
			checked, e := Verify(dest)
			if e != nil {
				return e
			}
			if checked.ManifestSHA256 != p.ManifestSHA256 || checked.ContentSHA256 != p.ContentSHA256 {
				return fmt.Errorf("existing release differs; choose a new version")
			}
			if e = probe(ctx, dest, checked); e != nil {
				return e
			}
		}
		selected = &Selection{Package: p, Release: dest, ActivatedAt: time.Now().UTC().Format(time.RFC3339)}
		if o.StageOnly {
			return nil
		}
		return c.activate(ctx, selected, o)
	})
	return selected, e
}
func (c Config) Rollback(ctx context.Context, o Options) (*Selection, error) {
	var selected *Selection
	e := c.mutate(func(c Config) error {
		s, e := loadState(c.Root)
		if e != nil {
			return e
		}
		if s.Previous == nil {
			return fmt.Errorf("no previous active release is recorded")
		}
		p, e := Verify(s.Previous.Release)
		if e != nil {
			return e
		}
		if p.ManifestSHA256 != s.Previous.Package.ManifestSHA256 || p.ContentSHA256 != s.Previous.Package.ContentSHA256 {
			return fmt.Errorf("previous release manifest changed")
		}
		selected = &Selection{Package: p, Release: s.Previous.Release, ActivatedAt: time.Now().UTC().Format(time.RFC3339)}
		o.SkipSkill = s.Previous.SkillDir == ""
		if !o.SkipSkill {
			c.SkillDir = s.Previous.SkillDir
		}
		return c.activate(ctx, selected, o)
	})
	return selected, e
}
func (c Config) ensureLaunchers() error {
	dir := filepath.Join(c.Root, "bin")
	if e := os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	if e := regularDirectory(dir); e != nil {
		return e
	}
	marker := filepath.Join(c.Root, "launchers.json")
	var expected map[string]string
	e := readJSON(marker, &expected)
	if e != nil && !os.IsNotExist(e) {
		return e
	}
	if os.IsNotExist(e) {
		expected = map[string]string{}
		h, e := hashFile(c.Launcher)
		if e != nil {
			return e
		}
		for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
			name := toolName(tool)
			if _, e = os.Lstat(filepath.Join(dir, name)); !os.IsNotExist(e) {
				return fmt.Errorf("refusing to claim an existing launcher: %s", name)
			}
			expected[name] = h
		}
		// Claim absent paths before creation so a crash can resume the same setup.
		if e = atomicJSON(marker, expected); e != nil {
			return e
		}
	}
	if len(expected) != 3 {
		return fmt.Errorf("invalid launcher ownership record")
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		name := toolName(tool)
		path := filepath.Join(dir, name)
		if !digestPattern.MatchString(expected[name]) {
			return fmt.Errorf("invalid launcher digest")
		}
		if _, e = os.Lstat(path); os.IsNotExist(e) {
			sourceHash, e := hashFile(c.Launcher)
			if e != nil {
				return e
			}
			if sourceHash != expected[name] {
				return fmt.Errorf("resume launcher setup with the original installer executable")
			}
			temp := path + ".stage-" + token()
			if e = copyRegular(c.Launcher, temp, 0755); e != nil {
				return e
			}
			h, e := hashFile(temp)
			if e != nil {
				os.Remove(temp)
				return e
			}
			if h != expected[name] {
				os.Remove(temp)
				return fmt.Errorf("launcher changed while copying")
			}
			// Hard-link publication refuses overwrite and never exposes a partial EXE.
			e = os.Link(temp, path)
			os.Remove(temp)
			if e != nil {
				return e
			}
		} else if e != nil {
			return e
		}
		info, e := os.Lstat(path)
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("launcher is not regular")
		}
		h, e := hashFile(path)
		if e != nil {
			return e
		}
		if h != expected[name] {
			return fmt.Errorf("launcher changed: %s", path)
		}
	}
	return nil
}
func copyRegular(from, to string, mode os.FileMode) error {
	in, e := os.Open(from)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if e != nil {
		return e
	}
	_, e = io.Copy(out, in)
	if e == nil {
		e = out.Sync()
	}
	ce := out.Close()
	if e == nil {
		e = ce
	}
	return e
}

func (c Config) activate(ctx context.Context, selected *Selection, o Options) (err error) {
	p, e := Verify(selected.Release)
	if e != nil {
		return e
	}
	if p.ManifestSHA256 != selected.Package.ManifestSHA256 || p.ContentSHA256 != selected.Package.ContentSHA256 {
		return fmt.Errorf("release changed before activation")
	}
	if e = probe(ctx, selected.Release, p); e != nil {
		return e
	}
	before, e := loadState(c.Root)
	if e != nil {
		return e
	}

	if e = c.ensureLaunchers(); e != nil {
		return e
	}
	if before.Current != nil && before.Current.Release == selected.Release && before.Current.Package.ContentSHA256 == p.ContentSHA256 && c.bindingsMatch(*before.Current, o) {
		*selected = *before.Current
		return nil
	}
	tx := transaction{Schema: "pptxgengo.activation/v1", Before: before, After: State{Schema: before.Schema, Current: selected, Previous: before.Current, OriginalSkills: before.OriginalSkills}}
	if before.Current != nil && before.Current.Release == selected.Release {
		tx.After.Previous = before.Previous
	}
	id := token()
	if !o.SkipSkill && p.Kind == "full" {
		tx.SkillPath = c.SkillDir
		tx.SkillStage = c.SkillDir + ".stage-" + id
		tx.SkillBackup = c.SkillDir + ".backup-" + id
		tx.AfterSkillHash = p.SkillSHA256
		if e = os.MkdirAll(filepath.Dir(c.SkillDir), 0700); e != nil {
			return e
		}
		if _, e = os.Lstat(c.SkillDir); e == nil {
			tx.BeforeSkillHash, e = optionalSkillHash(c.SkillDir)
			if e != nil {
				return fmt.Errorf("existing skill cannot be safely backed up: %w", e)
			}
		} else if !os.IsNotExist(e) {
			return e
		}
		defer func() {
			if _, e := os.Lstat(filepath.Join(c.Root, "pending.json")); os.IsNotExist(e) {
				os.RemoveAll(tx.SkillStage)
			}
		}()
		tx.After.OriginalSkills = map[string]OriginalSkill{}
		for path, original := range before.OriginalSkills {
			tx.After.OriginalSkills[path] = original
		}
		original, known := tx.After.OriginalSkills[c.SkillDir]
		if !known {
			original = OriginalSkill{Hash: tx.BeforeSkillHash}
			if original.Hash != "" {
				original.Backup = tx.SkillBackup
			}
		}
		original.ManagedHash = tx.AfterSkillHash
		tx.After.OriginalSkills[c.SkillDir] = original
		if e = copyTree(filepath.Join(selected.Release, "skills", "west-monroe-presentations"), tx.SkillStage); e != nil {
			return e
		}
		if copied, e := treeHash(tx.SkillStage); e != nil || copied != tx.AfterSkillHash {
			return fmt.Errorf("skill changed while staging: %v", e)
		}
		selected.SkillDir = c.SkillDir
		selected.SkillSHA256 = p.SkillSHA256
	}
	if !o.NoPath {
		if runtime.GOOS == "windows" {
			tx.ChangePath = true
			tx.BeforePath, e = userPath()
			if e != nil {
				return e
			}
			bin := filepath.Join(c.Root, "bin")
			parts := []string{bin}
			for _, part := range strings.Split(tx.BeforePath.Value, ";") {
				if part != "" && !strings.EqualFold(filepath.Clean(part), bin) {
					parts = append(parts, part)
				}
			}
			tx.AfterPath = pathSnapshot{Value: strings.Join(parts, ";"), Kind: tx.BeforePath.Kind, Exists: true}
			if !tx.BeforePath.Exists {
				tx.AfterPath.Kind = 2
			}
		} else {
			if e = os.MkdirAll(c.BinDir, 0700); e != nil {
				return e
			}
			for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
				link := LinkChange{Path: filepath.Join(c.BinDir, tool), After: filepath.Join(c.Root, "bin", tool)}
				if _, e = os.Lstat(link.Path); e == nil {
					old, e := os.Readlink(link.Path)
					if e != nil {
						return fmt.Errorf("refusing to replace an unowned executable: %s", link.Path)
					}
					absolute := old
					if !filepath.IsAbs(old) {
						absolute = filepath.Join(c.BinDir, old)
					}
					if !within(c.Root, filepath.Clean(absolute)) {
						return fmt.Errorf("refusing to replace an unowned launcher: %s", link.Path)
					}
					link.Before = &old
				} else if !os.IsNotExist(e) {
					return e
				}
				tx.Links = append(tx.Links, link)
			}
		}
	}
	return c.commit(tx)
}
func (c Config) commit(tx transaction) (err error) {
	var e error
	if e = c.validateReceipt(tx); e != nil {
		return e
	}
	state, e := loadState(c.Root)
	if e != nil {
		return e
	}
	if !equalJSON(state, tx.Before) {
		return fmt.Errorf("selection changed before activation")
	}
	if tx.SkillPath != "" {
		current, e := optionalSkillHash(tx.SkillPath)
		if e != nil {
			return e
		}
		if current != tx.BeforeSkillHash {
			return fmt.Errorf("user skill changed before activation; preserved")
		}
		staged, e := optionalSkillHash(tx.SkillStage)
		if e != nil {
			return e
		}
		if staged != tx.AfterSkillHash {
			return fmt.Errorf("staged skill changed before activation; preserved")
		}
		if _, e = os.Lstat(tx.SkillBackup); !os.IsNotExist(e) {
			return fmt.Errorf("backup path already exists; preserved")
		}
	}
	for _, link := range tx.Links {
		target, e := os.Readlink(link.Path)
		if link.Before == nil {
			if !os.IsNotExist(e) {
				return fmt.Errorf("launcher changed before activation")
			}
		} else if e != nil || target != *link.Before {
			return fmt.Errorf("launcher changed before activation; preserved")
		}
	}
	if tx.ChangePath {
		current, e := userPath()
		if e != nil {
			return e
		}
		if current != tx.BeforePath {
			return fmt.Errorf("user PATH changed before activation; preserved")
		}
	}
	if e = atomicJSON(filepath.Join(c.Root, "pending.json"), tx); e != nil {
		return e
	}
	defer func() {
		if err != nil {
			if recovery := c.recover(); recovery != nil {
				err = errors.Join(err, fmt.Errorf("activation recovery pending: %w", recovery))
			}
		}
	}()
	if tx.SkillPath != "" {
		if tx.BeforeSkillHash != "" {
			if e = os.Rename(tx.SkillPath, tx.SkillBackup); e != nil {
				return e
			}
		}
		if tx.AfterSkillHash != "" {
			if e = os.Rename(tx.SkillStage, tx.SkillPath); e != nil {
				return e
			}
		}
	}
	for _, link := range tx.Links {
		if link.After == "" {
			e = os.Remove(link.Path)
			if os.IsNotExist(e) {
				e = nil
			}
		} else {
			e = writeLink(link.Path, link.After)
		}
		if e != nil {
			return e
		}
	}
	if tx.ChangePath {
		if e = setUserPath(tx.AfterPath); e != nil {
			return e
		}
	}
	if e = atomicJSON(filepath.Join(c.Root, "active.json"), tx.After); e != nil {
		return e
	}
	return os.Remove(filepath.Join(c.Root, "pending.json"))
}
func (c Config) bindingsMatch(selected Selection, o Options) bool {
	if !o.SkipSkill && selected.Package.Kind == "full" {
		if selected.SkillDir != c.SkillDir {
			return false
		}
		h, e := optionalSkillHash(c.SkillDir)
		if e != nil || h != selected.Package.SkillSHA256 {
			return false
		}
	}
	if o.NoPath {
		return true
	}
	if runtime.GOOS == "windows" {
		path, e := userPath()
		if e != nil {
			return false
		}
		first := strings.Split(path.Value, ";")[0]
		return strings.EqualFold(filepath.Clean(first), filepath.Join(c.Root, "bin"))
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		target, e := os.Readlink(filepath.Join(c.BinDir, tool))
		if e != nil || target != filepath.Join(c.Root, "bin", tool) {
			return false
		}
	}
	return true
}
func writeLink(path, target string) error {
	tmp := path + ".tmp-" + token()
	if e := os.Symlink(target, tmp); e != nil {
		return e
	}
	defer os.Remove(tmp)
	return replaceFile(tmp, path)
}
func equalJSON(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

// Recover restores the recorded predecessor, refusing to discard any settings or
// skill content modified after the interrupted transaction.
func (c Config) Recover() error { return c.mutate(func(c Config) error { return nil }) }
func (c Config) validateReceipt(tx transaction) error {
	if tx.Schema != "pptxgengo.activation/v1" || tx.Before.Schema != "pptxgengo.installation/v1" || tx.After.Schema != tx.Before.Schema {
		return fmt.Errorf("unsupported activation receipt")
	}
	for _, state := range []State{tx.Before, tx.After} {
		for _, selection := range []*Selection{state.Current, state.Previous} {
			if selection != nil && (!versionPattern.MatchString(selection.Package.Version) || selection.Release != filepath.Join(c.Root, "releases", selection.Package.Version)) {
				return fmt.Errorf("receipt release escapes managed storage")
			}
		}
	}
	if tx.SkillPath != "" {
		if tx.SkillPath != c.SkillDir || !strings.HasPrefix(tx.SkillStage, tx.SkillPath+".stage-") || !strings.HasPrefix(tx.SkillBackup, tx.SkillPath+".backup-") {
			return fmt.Errorf("receipt skill path differs; use the original --skill-dir")
		}
		stageID := strings.TrimPrefix(tx.SkillStage, tx.SkillPath+".stage-")
		backupID := strings.TrimPrefix(tx.SkillBackup, tx.SkillPath+".backup-")
		if stageID != backupID || len(stageID) != 32 {
			return fmt.Errorf("invalid skill receipt identity")
		}
		if _, e := hex.DecodeString(stageID); e != nil {
			return e
		}
		if tx.AfterSkillHash != "" && !digestPattern.MatchString(tx.AfterSkillHash) || tx.BeforeSkillHash != "" && !digestPattern.MatchString(tx.BeforeSkillHash) {
			return fmt.Errorf("invalid skill receipt digest")
		}
	}
	seen := map[string]bool{}
	for _, link := range tx.Links {
		name := filepath.Base(link.Path)
		if name != "pptxgengo" && name != "pptxdesign" && name != "wmdsdocs" {
			return fmt.Errorf("unknown receipt launcher")
		}
		if seen[name] || link.Path != filepath.Join(c.BinDir, name) || (link.After != "" && link.After != filepath.Join(c.Root, "bin", name)) {
			return fmt.Errorf("receipt launcher escapes configured paths")
		}
		seen[name] = true
		if link.Before != nil {
			target := *link.Before
			if !filepath.IsAbs(target) {
				target = filepath.Join(c.BinDir, target)
			}
			if !within(c.Root, filepath.Clean(target)) {
				return fmt.Errorf("receipt predecessor launcher is not owned")
			}
		}
	}
	return nil
}
func (c Config) recover() error {
	var tx transaction
	pending := filepath.Join(c.Root, "pending.json")
	if e := readJSON(pending, &tx); e != nil {
		return e
	}
	if e := c.validateReceipt(tx); e != nil {
		return e
	}
	state, e := loadState(c.Root)
	if e != nil {
		return e
	}
	if !equalJSON(state, tx.Before) && !equalJSON(state, tx.After) {
		return fmt.Errorf("active selection changed outside the transaction")
	}
	// Validate all predecessors before attempting any restoration.
	for _, link := range tx.Links {
		target, e := os.Readlink(link.Path)
		if e == nil {
			if target != link.After && (link.Before == nil || target != *link.Before) {
				return fmt.Errorf("launcher changed outside transaction: %s", link.Path)
			}
		} else if !os.IsNotExist(e) {
			return e
		}
	}
	if tx.ChangePath {
		current, e := userPath()
		if e != nil {
			return e
		}
		if current != tx.BeforePath && current != tx.AfterPath {
			return fmt.Errorf("user PATH changed outside transaction; preserved")
		}
	}
	if tx.SkillPath != "" {
		staged, e := optionalSkillHash(tx.SkillStage)
		if e != nil {
			return e
		}
		if staged != "" && staged != tx.AfterSkillHash {
			return fmt.Errorf("staged skill changed; preserved")
		}
		current, e := optionalSkillHash(tx.SkillPath)
		if e != nil {
			return e
		}
		backup, e := optionalSkillHash(tx.SkillBackup)
		if e != nil {
			return e
		}
		if current != "" && current != tx.BeforeSkillHash && current != tx.AfterSkillHash {
			return fmt.Errorf("skill changed outside transaction; preserved")
		}
		if backup != "" && backup != tx.BeforeSkillHash {
			return fmt.Errorf("skill backup changed; preserved")
		}
		if tx.BeforeSkillHash != "" && current != tx.BeforeSkillHash && backup != tx.BeforeSkillHash {
			return fmt.Errorf("previous skill backup is missing")
		}
		if current != tx.BeforeSkillHash {
			if current != "" {
				if e = os.RemoveAll(tx.SkillPath); e != nil {
					return e
				}
			}
			if tx.BeforeSkillHash != "" {
				if e = os.Rename(tx.SkillBackup, tx.SkillPath); e != nil {
					return e
				}
			}
		}
	}
	for _, link := range tx.Links {
		if link.Before == nil {
			if e = os.Remove(link.Path); e != nil && !os.IsNotExist(e) {
				return e
			}
		} else if e = writeLink(link.Path, *link.Before); e != nil {
			return e
		}
	}
	if tx.ChangePath {
		if e = setUserPath(tx.BeforePath); e != nil {
			return e
		}
	}
	if e = atomicJSON(filepath.Join(c.Root, "active.json"), tx.Before); e != nil {
		return e
	}
	if tx.SkillStage != "" {
		if e = os.RemoveAll(tx.SkillStage); e != nil {
			return e
		}
	}
	return os.Remove(pending)
}
func optionalSkillHash(path string) (string, error) {
	if _, e := os.Lstat(path); os.IsNotExist(e) {
		return "", nil
	} else if e != nil {
		return "", e
	}
	info, e := os.Lstat(path)
	if e != nil {
		return "", e
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, e := os.Readlink(path)
		if e != nil {
			return "", e
		}
		h := sha256.Sum256([]byte("symlink\x00" + target))
		return hex.EncodeToString(h[:]), nil
	}
	return treeHash(path)
}

// ActiveRelease is used by the stable dispatcher. No state means this is a normal
// packaged executable. Corrupt or out-of-root state always fails closed.
func ActiveRelease(root string) (string, error) {
	s, e := loadState(root)
	if e != nil {
		return "", e
	}
	if s.Current == nil {
		return "", nil
	}
	return s.Current.Release, nil
}

// ActiveManagerAPI lets the immutable bootstrap delegate administration to the
// upgraded package while retaining recovery for older packages without this API.
func ActiveManagerAPI(root string) (int, error) {
	s, e := loadState(root)
	if e != nil {
		return 0, e
	}
	if s.Current == nil {
		return 0, nil
	}
	return s.Current.Package.ManagerAPI, nil
}
