// Package installstate manages immutable releases and recoverable user activation.
// Package hashes prove consistency, not publisher authenticity: verify the signed
// release manifest before running an installer from a downloaded package.
package installstate

import (
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Package struct {
	ManagerAPI     int               `json:"installation_api,omitempty"`
	Version        string            `json:"version"`
	OS             string            `json:"target_os"`
	Arch           string            `json:"target_arch"`
	Bundle         string            `json:"selected_bundle,omitempty"`
	SourceRevision string            `json:"source_revision,omitempty"`
	SourceCommit   string            `json:"source_commit,omitempty"`
	Kind           string            `json:"package_kind"`
	ContentSHA256  string            `json:"content_sha256"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	SkillSHA256    string            `json:"skill_sha256,omitempty"`
	Files          map[string]string `json:"-"`
}

var versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
var digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func toolName(tool string) string {
	if runtime.GOOS == "windows" {
		return tool + ".exe"
	}
	return tool
}
func hashFile(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
		return "", e
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func readJSON(path string, value any) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, 16*1024*1024+1))
	if e != nil {
		return e
	}
	if len(data) > 16*1024*1024 {
		return fmt.Errorf("metadata too large: %s", path)
	}
	return json.Unmarshal(data, value)
}
func safeRelative(path string) bool {
	if path == "" || strings.ContainsAny(path, `\:*?"<>|`) || strings.HasPrefix(path, "/") || filepath.ToSlash(filepath.Clean(path)) != path {
		return false
	}
	for _, p := range strings.Split(path, "/") {
		if p == ".." || p == "." || strings.HasSuffix(p, ".") || strings.HasSuffix(p, " ") {
			return false
		}
		base := strings.ToUpper(strings.SplitN(p, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || regexp.MustCompile(`^(COM|LPT)[1-9]$`).MatchString(base) {
			return false
		}
	}
	return true
}
func inventory(root string) (map[string]string, error) {
	info, e := os.Lstat(root)
	if e != nil {
		return nil, e
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("package root must be a directory, not a link")
	}
	files := map[string]string{}
	folded := map[string]bool{}
	e = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		if !safeRelative(rel) {
			return fmt.Errorf("unsafe package path: %s", rel)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("package links are unsupported: %s", rel)
		}
		if entry.IsDir() {
			return nil
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("package file is not regular: %s", rel)
		}
		key := strings.ToLower(rel)
		if folded[key] {
			return fmt.Errorf("case-insensitive path collision: %s", rel)
		}
		folded[key] = true
		h, e := hashFile(path)
		if e != nil {
			return e
		}
		files[rel] = h
		return nil
	})
	return files, e
}
func treeHash(root string) (string, error) {
	files, e := inventory(root)
	if e != nil {
		return "", e
	}
	return inventoryHash(files), nil
}
func inventoryHash(files map[string]string) string {
	keys := make([]string, 0, len(files))
	for p := range files {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, p := range keys {
		fmt.Fprintf(h, "%s\x00%s\n", p, files[p])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Verify checks native target, complete full-package inventory or the CLI binary
// evidence, required resources and version agreement. It never executes code.
func Verify(root string) (Package, error) {
	var p Package
	files, e := inventory(root)
	if e != nil {
		return p, e
	}
	manifest := "release-manifest.json"
	var raw struct {
		Schema string `json:"schema"`
		Package
		Count  int               `json:"file_count"`
		Hashes map[string]string `json:"files_sha256"`
	}
	if _, ok := files[manifest]; ok {
		if e = readJSON(filepath.Join(root, manifest), &raw); e != nil {
			return p, e
		}
		p = raw.Package
		p.Kind = "full"
		if raw.Schema != "pptxgengo.local-release-manifest.v1" || raw.Count != len(raw.Hashes) || len(files) != len(raw.Hashes)+1 {
			return p, fmt.Errorf("full package inventory/schema mismatch")
		}
		for name, h := range raw.Hashes {
			if !safeRelative(name) || !digestPattern.MatchString(h) || name == manifest || files[name] != h {
				return p, fmt.Errorf("missing, unsafe or changed package file: %s", name)
			}
		}
		p.Files = raw.Hashes
		if !regexp.MustCompile(`^v[1-9][0-9]*$`).MatchString(p.Bundle) {
			return p, fmt.Errorf("invalid selected bundle")
		}
		for _, name := range []string{"release/VERSION", "skills/west-monroe-presentations/SKILL.md", "wmds-docs/site/SOURCE.json", "library/wm-design-system/" + p.Bundle + "/bundle.json", "library/wm-design-system/" + p.Bundle + "/library.sqlite", "library/wm-design-system/" + p.Bundle + "/catalog/design-system.html"} {
			if _, ok := files[name]; !ok {
				return p, fmt.Errorf("required resource missing: %s", name)
			}
		}
		p.SkillSHA256, e = treeHash(filepath.Join(root, "skills", "west-monroe-presentations"))
		if e != nil {
			return p, e
		}
	} else {
		manifest = "build-evidence.json"
		var ev struct {
			Schema  string            `json:"schema"`
			Version string            `json:"version"`
			Target  string            `json:"target"`
			Hashes  map[string]string `json:"binaries_sha256"`
		}
		if e = readJSON(filepath.Join(root, manifest), &ev); e != nil {
			return p, fmt.Errorf("package manifest or build evidence required: %w", e)
		}
		parts := strings.Split(ev.Target, "-")
		if ev.Schema != "pptxgengo.build-evidence/v1" || len(parts) != 2 || len(ev.Hashes) != 3 {
			return p, fmt.Errorf("invalid CLI evidence")
		}
		p.Version, p.OS, p.Arch, p.Kind = ev.Version, parts[0], parts[1], "cli-only"
		for name, h := range ev.Hashes {
			if !safeRelative(name) || !digestPattern.MatchString(h) || files[name] != h {
				return p, fmt.Errorf("CLI binary hash mismatch: %s", name)
			}
		}
		allowed := map[string]bool{"VERSION": true, "README.txt": true, "LICENSE": true, "THIRD-PARTY-NOTICES.txt": true, manifest: true}
		for name := range ev.Hashes {
			allowed[name] = true
		}
		for name := range files {
			if !allowed[name] {
				return p, fmt.Errorf("unlisted CLI file: %s", name)
			}
		}
		p.Files = files
	}
	if !versionPattern.MatchString(p.Version) || p.Version == "." || p.Version == ".." {
		return p, fmt.Errorf("invalid release version")
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		name := "bin/" + toolName(tool)
		if _, ok := files[name]; !ok {
			return p, fmt.Errorf("missing executable: %s", name)
		}
		info, e := buildinfo.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil {
			return p, fmt.Errorf("invalid Go executable %s: %w", name, e)
		}
		settings := map[string]string{}
		for _, s := range info.Settings {
			settings[s.Key] = s.Value
		}
		if p.OS == "" {
			p.OS = settings["GOOS"]
		}
		if p.Arch == "" {
			p.Arch = settings["GOARCH"]
		}
		if settings["GOOS"] != p.OS || settings["GOARCH"] != p.Arch {
			return p, fmt.Errorf("executable architecture disagrees with manifest: %s", name)
		}
	}
	native, e := nativeArchitecture()
	if e != nil {
		return p, fmt.Errorf("native architecture check failed: %w", e)
	}
	if p.OS != runtime.GOOS || p.Arch != runtime.GOARCH || p.Arch != native {
		return p, fmt.Errorf("package targets %s/%s; this executable is %s/%s on native %s", p.OS, p.Arch, runtime.GOOS, runtime.GOARCH, native)
	}
	for _, name := range []string{"VERSION", "release/VERSION"} {
		if _, ok := files[name]; ok {
			b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
			if e != nil {
				return p, e
			}
			if strings.TrimSpace(string(b)) != p.Version {
				return p, fmt.Errorf("package version mismatch: %s", name)
			}
		} else if name == "VERSION" {
			return p, fmt.Errorf("VERSION is required")
		}
	}
	binaryBytes, e := os.ReadFile(filepath.Join(root, "bin", toolName("pptxgengo")))
	if e != nil {
		return p, e
	}
	if strings.Contains(string(binaryBytes), "pptxgengo.installation-api/v1") {
		p.ManagerAPI = 1
	}
	p.ManifestSHA256 = files[manifest]
	p.ContentSHA256 = inventoryHash(files)
	return p, nil
}
func probe(ctx context.Context, root string, p Package) error {
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		bounded, cancel := context.WithTimeout(ctx, 30*time.Second)
		cmd := exec.CommandContext(bounded, filepath.Join(root, "bin", toolName(tool)), "--version")
		cmd.WaitDelay = time.Second
		out, e := cmd.Output()
		cancel()
		if e != nil {
			return fmt.Errorf("%s startup check failed: %w", tool, e)
		}
		if strings.TrimSpace(string(out)) != p.Version {
			return fmt.Errorf("%s reports an unexpected version", tool)
		}
	}
	return nil
}
func copyTree(from, to string) error {
	if e := os.Mkdir(to, 0700); e != nil {
		return e
	}
	return filepath.WalkDir(from, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == from {
			return nil
		}
		rel, e := filepath.Rel(from, path)
		if e != nil {
			return e
		}
		dst := filepath.Join(to, rel)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source link: %s", rel)
		}
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if info.IsDir() {
			return os.Mkdir(dst, 0700)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source is not regular: %s", rel)
		}
		in, e := os.Open(path)
		if e != nil {
			return e
		}
		defer in.Close()
		out, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if e != nil {
			return e
		}
		_, e = io.Copy(out, in)
		if e == nil {
			e = out.Sync()
		}
		closeErr := out.Close()
		if e == nil {
			e = closeErr
		}
		return e
	})
}
