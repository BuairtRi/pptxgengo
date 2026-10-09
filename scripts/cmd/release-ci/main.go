// release-ci owns archive layout, release identity, checksums and private GitLab
// publication. Model package pins use a dependency-free package; signing jobs
// do not import or initialize the inference runtime.
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var targets = []string{"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64", "windows-amd64", "windows-arm64"}
var tools = []string{"pptxgengo", "pptxdesign", "wmdsdocs"}
var versionRE = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$`)

type Evidence struct {
	Schema       string            `json:"schema"`
	Target       string            `json:"target"`
	Version      string            `json:"version"`
	Commit       string            `json:"commit"`
	Verification string            `json:"verification"`
	Binaries     map[string]string `json:"binaries_sha256"`
	Unsigned     *Evidence         `json:"unsigned,omitempty"`
}
type Manifest struct {
	OfflineModelArchive       string            `json:"offline_model_archive,omitempty"`
	Schema                    string            `json:"schema"`
	Project                   string            `json:"project"`
	Version                   string            `json:"version"`
	Commit                    string            `json:"commit"`
	Pipeline                  string            `json:"pipeline"`
	PackageKind               string            `json:"package_kind"`
	Branding                  string            `json:"branding"`
	BrowsingPolicy            string            `json:"browsing_policy,omitempty"`
	WindowsRuntimeQualified   bool              `json:"windows_runtime_qualified"`
	NativePowerPointQualified bool              `json:"native_powerpoint_qualified"`
	Files                     map[string]string `json:"files_sha256"`
	Targets                   []string          `json:"targets"`
}

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "release-ci:", err)
		os.Exit(1)
	}
}
func digest(path string) (string, error) {
	f, e := os.Open(path)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	_, e = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), e
}
func readJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func writeJSON(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func name(tool, target string) string {
	if strings.HasPrefix(target, "windows-") {
		return tool + ".exe"
	}
	return tool
}
func validTarget(target string) bool {
	for _, t := range targets {
		if t == target {
			return true
		}
	}
	return false
}
func identity(target, version, commit string) error {
	if !validTarget(target) || !versionRE.MatchString(version) || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(commit) {
		return fmt.Errorf("invalid release identity")
	}
	return nil
}
func binaryIdentity(path, target, version string) error {
	info, e := buildinfo.ReadFile(path)
	if e != nil {
		return e
	}
	settings := map[string]string{}
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}
	parts := strings.Split(target, "-")
	if info.GoVersion != "go1.27.2" || settings["GOOS"] != parts[0] || settings["GOARCH"] != parts[1] || settings["CGO_ENABLED"] != "0" {
		return fmt.Errorf("wrong Go build identity for %s", path)
	}
	return nil
}
func binaries(dir, target, version, commit string) (map[string]string, error) {
	result := map[string]string{}
	for _, tool := range tools {
		rel := "bin/" + name(tool, target)
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if e := binaryIdentity(p, target, version); e != nil {
			return nil, e
		}
		payload, e := os.ReadFile(p)
		if e != nil {
			return nil, e
		}
		marker := "pptxgengo-build:" + version + ":" + commit + ":" + target
		if !strings.Contains(string(payload), marker) {
			return nil, fmt.Errorf("release marker mismatch: %s", p)
		}
		d, e := digest(p)
		if e != nil {
			return nil, e
		}
		result[rel] = d
	}
	return result, nil
}
func evidence(dir, target, version, commit, verification, previous string) error {
	if e := identity(target, version, commit); e != nil {
		return e
	}
	b, e := binaries(dir, target, version, commit)
	if e != nil {
		return e
	}
	ev := Evidence{Schema: "pptxgengo.build-evidence/v1", Target: target, Version: version, Commit: commit, Verification: verification, Binaries: b}
	if previous != "" {
		var p Evidence
		if e = readJSON(previous, &p); e != nil {
			return e
		}
		if p.Target != target || p.Version != version || p.Commit != commit || p.Verification != "reproducible" {
			return fmt.Errorf("unsigned evidence identity mismatch")
		}
		ev.Unsigned = &p
	}
	return writeJSON(filepath.Join(dir, "build-evidence.json"), ev)
}
func verifyBinaries(dir, target, version, commit string) error {
	if e := identity(target, version, commit); e != nil {
		return e
	}
	var ev Evidence
	if e := readJSON(filepath.Join(dir, "build-evidence.json"), &ev); e != nil {
		return e
	}
	if ev.Schema != "pptxgengo.build-evidence/v1" || ev.Target != target || ev.Version != version || ev.Commit != commit {
		return fmt.Errorf("evidence identity mismatch")
	}
	b, e := binaries(dir, target, version, commit)
	if e != nil {
		return e
	}
	if len(b) != len(ev.Binaries) {
		return fmt.Errorf("binary count mismatch")
	}
	for p, h := range b {
		if ev.Binaries[p] != h {
			return fmt.Errorf("binary hash mismatch: %s", p)
		}
	}
	return nil
}
func safePath(p string) bool {
	return p != "" && !strings.Contains(p, `\`) && !strings.HasPrefix(p, "/") && !strings.Contains(p, ":") && filepath.ToSlash(filepath.Clean(p)) == p && p != ".." && !strings.HasPrefix(p, "../")
}

type Input struct {
	Data []byte
	Path string
}

func (f Input) size() (int64, error) {
	if f.Path == "" {
		return int64(len(f.Data)), nil
	}
	s, e := os.Stat(f.Path)
	if e != nil {
		return 0, e
	}
	return s.Size(), nil
}
func (f Input) write(w io.Writer) error {
	if f.Path == "" {
		_, e := w.Write(f.Data)
		return e
	}
	r, e := os.Open(f.Path)
	if e != nil {
		return e
	}
	defer r.Close()
	_, e = io.Copy(w, r)
	return e
}
func (f Input) hash() (string, error) {
	if f.Path != "" {
		return digest(f.Path)
	}
	h := sha256.Sum256(f.Data)
	return hex.EncodeToString(h[:]), nil
}
func packageKind() string {
	kind := os.Getenv("PPTXGENGO_PACKAGE_KIND")
	if kind == "" {
		kind = "cli-only"
	}
	return kind
}
func archive(dir, target, version, commit, out string) error {
	if e := verifyBinaries(dir, target, version, commit); e != nil {
		return e
	}
	files := map[string]Input{}
	var local map[string]any
	if packageKind() == "full" {
		if e := readJSON("dist/resources/release-manifest.json", &local); e != nil {
			return fmt.Errorf("full release requires verified resources: %w", e)
		}
		hashes, ok := local["files_sha256"].(map[string]any)
		if !ok {
			return fmt.Errorf("resource hashes missing")
		}
		for rel, hash := range hashes {
			if !safePath(rel) {
				return fmt.Errorf("unsafe resource path")
			}
			if strings.HasPrefix(rel, "bin/") {
				continue
			}
			path := filepath.Join("dist/resources", filepath.FromSlash(rel))
			d, e := digest(path)
			if e != nil {
				return e
			}
			if d != hash {
				return fmt.Errorf("resource drift: %s", rel)
			}
			files[rel] = Input{Path: path}
		}
		for _, p := range []string{"install-windows.ps1", "smoke-test-windows.ps1", "WINDOWS.md"} {
			files[p] = Input{Path: filepath.Join("internal/releasepackage", p)}
		}
		if e := filepath.WalkDir("internal/releasepackage/guides", func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if !d.IsDir() {
				files["guides/"+d.Name()] = Input{Path: path}
			}
			return nil
		}); e != nil {
			return e
		}
	} else if packageKind() != "cli-only" {
		return fmt.Errorf("unknown package kind")
	}
	if e := addScopedBrowsingFiles(files, "dist/resources", version, commit); e != nil {
		return e
	}
	for _, tool := range tools {
		p := "bin/" + name(tool, target)
		files[p] = Input{Path: filepath.Join(dir, filepath.FromSlash(p))}
	}
	files["VERSION"] = Input{Data: []byte(version + "\n")}
	files["README.txt"] = Input{Data: []byte("pptxgengo " + version + " (" + target + ")\nPackage kind: " + packageKind() + "\nExtract and add bin to PATH. Run pptxgengo --version.\n" + browsingContentDescription(packageKind(), browsingPolicy()) + "\nWindows runtime and native PowerPoint validation remain pending.\nBare macOS CLIs cannot be stapled; Apple online ticket lookup is required on first use.\nVerify manifest.sigstore.json before trusting checksums and installing.\n")}
	for _, p := range []string{"LICENSE", "internal/releasepackage/THIRD-PARTY-NOTICES.txt"} {
		files[filepath.Base(p)] = Input{Path: p}
	}
	files["build-evidence.json"] = Input{Path: filepath.Join(dir, "build-evidence.json")}
	if local != nil {
		hashes := map[string]string{}
		for p, f := range files {
			d, e := f.hash()
			if e != nil {
				return e
			}
			hashes[p] = d
		}
		local["version"] = version
		local["target_os"] = strings.Split(target, "-")[0]
		local["target_arch"] = strings.Split(target, "-")[1]
		local["windows_runtime_qualified"] = false
		local["files_sha256"] = hashes
		local["file_count"] = len(hashes)
		data, e := json.MarshalIndent(local, "", "  ")
		if e != nil {
			return e
		}
		files["release-manifest.json"] = Input{Data: append(data, '\n')}
	}
	return writeArchive(files, target, out)
}
func writeArchive(files map[string]Input, target, out string) (err error) {
	if e := os.MkdirAll(filepath.Dir(out), 0755); e != nil {
		return e
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	defer func() {
		ce := f.Close()
		if err == nil {
			err = ce
		}
		if err != nil {
			os.Remove(out)
		}
	}()
	keys := []string{}
	folded := map[string]bool{}
	for p := range files {
		if !safePath(p) || folded[strings.ToLower(p)] {
			return fmt.Errorf("unsafe or duplicate archive path %s", p)
		}
		folded[strings.ToLower(p)] = true
		keys = append(keys, p)
	}
	sort.Strings(keys)
	if strings.HasPrefix(target, "windows-") {
		z := zip.NewWriter(f)
		for _, p := range keys {
			h := &zip.FileHeader{Name: p, Method: zip.Deflate}
			h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
			mode := os.FileMode(0644)
			if strings.HasPrefix(p, "bin/") {
				mode = 0755
			}
			h.SetMode(mode)
			w, e := z.CreateHeader(h)
			if e != nil {
				return e
			}
			if e = files[p].write(w); e != nil {
				return e
			}
		}
		return z.Close()
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, p := range keys {
		mode := int64(0644)
		if strings.HasPrefix(p, "bin/") {
			mode = 0755
		}
		size, e := files[p].size()
		if e != nil {
			return e
		}
		h := &tar.Header{Name: p, Mode: mode, Size: size, ModTime: time.Unix(0, 0), Typeflag: tar.TypeReg}
		if e = tw.WriteHeader(h); e != nil {
			return e
		}
		if e = files[p].write(tw); e != nil {
			return e
		}
	}
	if e = tw.Close(); e != nil {
		return e
	}
	return gz.Close()
}
func extract(archive, out string) error {
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		return fmt.Errorf("extraction destination must be new")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		return e
	}
	seen := map[string]bool{}
	save := func(p string, mode os.FileMode, size int64, r io.Reader) error {
		if !safePath(p) || seen[strings.ToLower(p)] || size < 0 || size > 1<<30 {
			return fmt.Errorf("unsafe or duplicate archive member: %s", p)
		}
		seen[strings.ToLower(p)] = true
		dst := filepath.Join(out, filepath.FromSlash(p))
		if e := os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return e
		}
		f, e := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm()&0755)
		if e != nil {
			return e
		}
		_, e = io.CopyN(f, r, size)
		ce := f.Close()
		if e != nil {
			return e
		}
		return ce
	}
	if strings.HasSuffix(archive, ".zip") {
		z, e := zip.OpenReader(archive)
		if e != nil {
			return e
		}
		defer z.Close()
		for _, f := range z.File {
			if !f.Mode().IsRegular() {
				return fmt.Errorf("nonregular ZIP member")
			}
			r, e := f.Open()
			if e != nil {
				return e
			}
			e = save(f.Name, f.Mode(), int64(f.UncompressedSize64), r)
			r.Close()
			if e != nil {
				return e
			}
		}
		return nil
	}
	f, e := os.Open(archive)
	if e != nil {
		return e
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		return e
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, e := tr.Next()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		if h.Typeflag != tar.TypeReg {
			return fmt.Errorf("nonregular TAR member")
		}
		if e = save(h.Name, os.FileMode(h.Mode), h.Size, tr); e != nil {
			return e
		}
	}
}
func archiveName(target, version string) string {
	ext := ".tar.gz"
	if strings.HasPrefix(target, "windows-") {
		ext = ".zip"
	}
	return "pptxgengo-" + version + "-" + target + ext
}
func assemble(version, commit string) error {
	out := "dist/final"
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		return fmt.Errorf("final directory must be new")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		return e
	}
	resourcePolicy, e := readResourcePolicy("dist/resources", version, commit)
	if e != nil {
		return e
	}
	if resourcePolicy.PackageKind != packageKind() || resourcePolicy.BrowsingPolicy != browsingPolicy() {
		return fmt.Errorf("resource policy differs from assembly intent")
	}
	if e = writeJSON(filepath.Join(out, resourcePolicyName), resourcePolicy); e != nil {
		return e
	}
	for _, target := range targets {
		prefix := "signed"
		verification := "notarized"
		if strings.HasPrefix(target, "linux-") {
			prefix = "unsigned"
			verification = "reproducible"
		}
		if strings.HasPrefix(target, "windows-") {
			verification = "authenticode"
		}
		dir := filepath.Join("dist", prefix, target)
		if e := verifyBinaries(dir, target, version, commit); e != nil {
			return e
		}
		var ev Evidence
		if e := readJSON(filepath.Join(dir, "build-evidence.json"), &ev); e != nil {
			return e
		}
		if ev.Verification != verification {
			return fmt.Errorf("required platform verification missing: %s", target)
		}
		if prefix == "signed" && (ev.Unsigned == nil || ev.Unsigned.Verification != "reproducible") {
			return fmt.Errorf("unsigned reproducibility proof missing")
		}
		if strings.HasPrefix(target, "darwin-") {
			var notary struct {
				Status string `json:"status"`
				ID     string `json:"id"`
			}
			if e := readJSON(filepath.Join(dir, "notarization.json"), &notary); e != nil {
				return e
			}
			if notary.Status != "Accepted" || notary.ID == "" {
				return fmt.Errorf("notarization not Accepted")
			}
			b, e := os.ReadFile(filepath.Join(dir, "notarization.json"))
			if e != nil {
				return e
			}
			if e = os.WriteFile(filepath.Join(out, target+".notarization.json"), b, 0644); e != nil {
				return e
			}
		}
		if e := archive(dir, target, version, commit, filepath.Join(out, archiveName(target, version))); e != nil {
			return e
		}
		b, e := os.ReadFile(filepath.Join(dir, "build-evidence.json"))
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(out, target+".build-evidence.json"), b, 0644); e != nil {
			return e
		}
	}
	if offlineModelEnabled() {
		if e := copyModelRelease("dist/model", out, version, commit); e != nil {
			return e
		}
	}
	return nil
}
func seal(dir, version, commit string) error {
	var policy struct {
		Passed bool `json:"passed"`
	}
	if e := readJSON(filepath.Join(dir, "security-policy.json"), &policy); e != nil {
		return e
	}
	if !policy.Passed {
		return fmt.Errorf("security policy failed")
	}
	resources, e := readResourcePolicy(dir, version, commit)
	if e != nil {
		return e
	}
	if resources.PackageKind != packageKind() || resources.BrowsingPolicy != browsingPolicy() {
		return fmt.Errorf("resource policy differs from sealing intent")
	}
	m := Manifest{Schema: "pptxgengo.release-manifest/v1", Project: "riscott/pptxgengo", Version: version, Commit: commit, Pipeline: os.Getenv("CI_PIPELINE_URL"), PackageKind: resources.PackageKind, BrowsingPolicy: resources.BrowsingPolicy, Branding: "verified-private-browsing-decks", Files: map[string]string{}, Targets: targets}
	if resources.BrowsingPolicy == "deferred" {
		m.Branding = "none-distribution-deferred"
	} else if resources.BrowsingPolicy == "templates-only" {
		m.Branding = "template-catalog-no-private-media"
	}
	if offlineModelEnabled() {
		m.OfflineModelArchive = modelArchiveName(version)
		if e := verifyModelRelease(dir, version, commit); e != nil {
			return e
		}
		for _, suffix := range []string{"", ".evidence.json", ".sbom.json", ".vulnerabilities.json"} {
			name := m.OfflineModelArchive + suffix
			hash, e := digest(filepath.Join(dir, name))
			if e != nil {
				return e
			}
			m.Files[name] = hash
		}
	}
	if packageKind() == "full" {
		m.Branding = "verified-private-originals"
	}
	for _, target := range targets {
		for _, suffix := range []string{"", ".sbom.json", ".vulnerabilities.json"} {
			p := archiveName(target, version) + suffix
			d, e := digest(filepath.Join(dir, p))
			if e != nil {
				return e
			}
			m.Files[p] = d
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	for _, entry := range entries {
		p := entry.Name()
		if entry.IsDir() || p == "manifest.json" || p == "SHA256SUMS" || p == "manifest.sigstore.json" {
			return fmt.Errorf("unexpected final artifact %s", p)
		}
		d, e := digest(filepath.Join(dir, p))
		if e != nil {
			return e
		}
		m.Files[p] = d
	}
	if e = writeJSON(filepath.Join(dir, "manifest.json"), m); e != nil {
		return e
	}
	hashes := map[string]string{}
	for p, d := range m.Files {
		hashes[p] = d
	}
	hashes["manifest.json"], e = digest(filepath.Join(dir, "manifest.json"))
	if e != nil {
		return e
	}
	keys := []string{}
	for p := range hashes {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	var sums strings.Builder
	for _, p := range keys {
		fmt.Fprintf(&sums, "%s  %s\n", hashes[p], p)
	}
	return os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(sums.String()), 0644)
}
func verify(dir, version, commit string) error {
	var m Manifest
	if e := readJSON(filepath.Join(dir, "manifest.json"), &m); e != nil {
		return e
	}
	if m.Schema != "pptxgengo.release-manifest/v1" || m.Project != "riscott/pptxgengo" || m.Version != version || m.Commit != commit || len(m.Targets) != 6 || (m.PackageKind != "cli-only" && m.PackageKind != "full") {
		return fmt.Errorf("manifest identity mismatch")
	}
	_, resourcePolicyErr := os.Lstat(filepath.Join(dir, resourcePolicyName))
	if m.BrowsingPolicy != "" || m.Files[resourcePolicyName] != "" || !os.IsNotExist(resourcePolicyErr) {
		p, err := readResourcePolicy(dir, version, commit)
		if err != nil {
			return err
		}
		if m.Files[resourcePolicyName] == "" || p.BrowsingPolicy != m.BrowsingPolicy || p.PackageKind != m.PackageKind {
			return fmt.Errorf("signed resource policy mismatch")
		}
		branding := "verified-private-browsing-decks"
		if p.PackageKind == "full" {
			branding = "verified-private-originals"
		}
		if p.BrowsingPolicy == "deferred" {
			branding = "none-distribution-deferred"
		} else if p.BrowsingPolicy == "templates-only" {
			branding = "template-catalog-no-private-media"
		}
		if m.Branding != branding {
			return fmt.Errorf("signed branding scope mismatch")
		}
	}
	for i, target := range targets {
		if m.Targets[i] != target {
			return fmt.Errorf("manifest target inventory mismatch")
		}
		for _, name := range []string{archiveName(target, version), archiveName(target, version) + ".sbom.json", archiveName(target, version) + ".vulnerabilities.json", target + ".build-evidence.json"} {
			if m.Files[name] == "" {
				return fmt.Errorf("required release evidence missing: %s", name)
			}
		}
		if strings.HasPrefix(target, "darwin-") && m.Files[target+".notarization.json"] == "" {
			return fmt.Errorf("notarization evidence missing")
		}
	}
	if m.OfflineModelArchive != "" {
		if m.OfflineModelArchive != modelArchiveName(version) {
			return fmt.Errorf("offline model archive identity mismatch")
		}
		for _, suffix := range []string{"", ".sbom.json", ".vulnerabilities.json", ".evidence.json"} {
			if m.Files[m.OfflineModelArchive+suffix] == "" {
				return fmt.Errorf("offline model release evidence missing: %s", suffix)
			}
		}
		if e := verifyModelRelease(dir, version, commit); e != nil {
			return e
		}
	} else {
		for name := range m.Files {
			if strings.Contains(name, "-offline-model.zip") {
				return fmt.Errorf("unidentified optional model artifact")
			}
		}
	}
	if m.Files["security-policy.json"] == "" {
		return fmt.Errorf("security policy missing")
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	for _, entry := range entries {
		name := entry.Name()
		info, e := entry.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular final artifact: %s", name)
		}
		if name != "manifest.json" && name != "SHA256SUMS" && name != "manifest.sigstore.json" && m.Files[name] == "" {
			return fmt.Errorf("unattested final artifact: %s", name)
		}
	}
	for p, h := range m.Files {
		if !safePath(p) {
			return fmt.Errorf("unsafe manifest path")
		}
		d, e := digest(filepath.Join(dir, p))
		if e != nil {
			return e
		}
		if d != h {
			return fmt.Errorf("final artifact hash mismatch: %s", p)
		}
	}
	return nil
}
func request(method, endpoint string, body io.Reader) ([]byte, int, error) {
	return requestWithCredential(method, endpoint, body, "JOB-TOKEN", os.Getenv("CI_JOB_TOKEN"))
}

func requestWithCredential(method, endpoint string, body io.Reader, header, credential string) ([]byte, int, error) {
	req, e := http.NewRequest(method, endpoint, body)
	if e != nil {
		return nil, 0, e
	}
	req.Header.Set(header, credential)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	if file, ok := body.(*os.File); ok {
		info, err := file.Stat()
		if err != nil {
			return nil, 0, err
		}
		req.ContentLength = info.Size()
	}
	client := &http.Client{Timeout: 30 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := client.Do(req)
	if e != nil {
		return nil, 0, e
	}
	defer resp.Body.Close()
	b, e := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	return b, resp.StatusCode, e
}

// Hash package downloads as a stream: full resource archives can be large.
func downloadDigest(endpoint string) (string, int, error) {
	req, e := http.NewRequest("GET", endpoint, nil)
	if e != nil {
		return "", 0, e
	}
	req.Header.Set("JOB-TOKEN", os.Getenv("CI_JOB_TOKEN"))
	client := &http.Client{Timeout: 30 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	resp, e := client.Do(req)
	if e != nil {
		return "", 0, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", resp.StatusCode, nil
	}
	h := sha256.New()
	_, e = io.Copy(h, resp.Body)
	return hex.EncodeToString(h.Sum(nil)), resp.StatusCode, e
}
func publish(dir, version, commit string) error {
	if e := verify(dir, version, commit); e != nil {
		return e
	}
	if os.Getenv("CI_PROJECT_ID") != "17" || os.Getenv("CI_COMMIT_REF_PROTECTED") != "true" || os.Getenv("CI_COMMIT_TAG") != version || os.Getenv("CI_JOB_TOKEN") == "" {
		return fmt.Errorf("publication requires protected project CI")
	}
	base := "https://gitlab.samcott.com/api/v4/projects/17"
	// The publication lane verifies privacy again; changing project visibility
	// cannot accidentally expose a new release. This endpoint does not accept
	// CI_JOB_TOKEN; a protected, environment-scoped Guest/read_api project token
	// is used only for this read. Uploads and release creation use the job token.
	privacyToken := os.Getenv("RELEASE_PRIVACY_READ_TOKEN")
	if privacyToken == "" {
		return fmt.Errorf("publication requires the protected project privacy read token")
	}
	b, status, e := requestWithCredential("GET", base, nil, "PRIVATE-TOKEN", privacyToken)
	if e != nil {
		return e
	}
	var project struct {
		Visibility                 string `json:"visibility"`
		PackageRegistryAccessLevel string `json:"package_registry_access_level"`
	}
	if status != 200 {
		return fmt.Errorf("cannot inspect publication project: HTTP %d", status)
	}
	if e = json.Unmarshal(b, &project); e != nil {
		return e
	}
	if project.Visibility != "private" || project.PackageRegistryAccessLevel != "private" {
		return fmt.Errorf("release project and package registry must remain private")
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return e
	}
	links := []map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("unexpected directory")
		}
		p := entry.Name()
		endpoint := base + "/packages/generic/pptxgengo-release/" + url.PathEscape(version) + "/" + url.PathEscape(p)
		existing, status, e := downloadDigest(endpoint)
		if e != nil {
			return e
		}
		localHash, e := digest(filepath.Join(dir, p))
		if e != nil {
			return e
		}
		if status == 200 {
			if existing != localHash {
				return fmt.Errorf("refusing to replace published artifact %s", p)
			}
		} else if status == 404 {
			payload, err := os.Open(filepath.Join(dir, p))
			if err != nil {
				return err
			}
			_, status, e = request("PUT", endpoint, payload)
			payload.Close()
			if e != nil {
				return e
			}
			if status != 201 {
				return fmt.Errorf("upload %s failed: HTTP %d", p, status)
			}
		} else {
			return fmt.Errorf("artifact check failed: HTTP %d", status)
		}
		downloaded, status, e := downloadDigest(endpoint)
		if e != nil {
			return e
		}
		if status != 200 || downloaded != localHash {
			return fmt.Errorf("published bytes differ for %s", p)
		}
		links = append(links, map[string]string{"name": p, "url": endpoint, "link_type": "package"})
	}
	releaseKind := "release"
	if strings.Contains(version, "-rc.") {
		releaseKind = "prerelease"
	}
	var publishedManifest Manifest
	if e := readJSON(filepath.Join(dir, "manifest.json"), &publishedManifest); e != nil {
		return e
	}
	description := "Signed CLI " + releaseKind + " for macOS, Linux and Windows (amd64 and arm64).\n\nPackage kind: " + publishedManifest.PackageKind + ". Every installation archive contains three executables. " + browsingContentDescription(publishedManifest.PackageKind, publishedManifest.BrowsingPolicy) + "\n\nWindows runtime and native PowerPoint validation are pending the interactive desktop runner. macOS notarization uses online Apple ticket lookup.\n\nFinal archives have CycloneDX SBOMs and vulnerability scans. manifest.json is signed with the private Sigstore service; verify its bundle against release/sigstore-policy.json before trusting checksums.\n\nSource: `" + commit + "`\nPipeline: " + publishedManifest.Pipeline
	description += modelReleaseDescription(publishedManifest)
	body, e := json.Marshal(map[string]any{"name": "pptxgengo " + version, "tag_name": version, "description": description, "assets": map[string]any{"links": links}})
	if e != nil {
		return e
	}
	_, status, e = request("POST", base+"/releases", strings.NewReader(string(body)))
	if e != nil {
		return e
	}
	if status != 201 {
		return fmt.Errorf("release creation failed: HTTP %d", status)
	}
	fmt.Println("Published and downloaded-verified private release:", version)
	return nil
}
func main() {
	a := os.Args[1:]
	if len(a) == 0 {
		fail(fmt.Errorf("expected a release-ci command"))
		return
	}
	var e error
	switch a[0] {
	case "prepare-resource-policy":
		if len(a) != 4 {
			e = fmt.Errorf("prepare-resource-policy DIR VERSION COMMIT")
			break
		}
		e = prepareResourcePolicy(a[1], a[2], a[3])
	case "model-package":
		if len(a) != 5 {
			e = fmt.Errorf("model-package DIR OUT_DIR VERSION COMMIT")
			break
		}
		e = assembleModel(a[1], a[2], a[3], a[4])
	case "verify-model":
		if len(a) != 4 {
			e = fmt.Errorf("verify-model DIR VERSION COMMIT")
			break
		}
		e = verifyModelRelease(a[1], a[2], a[3])
	case "evidence":
		if len(a) != 6 && len(a) != 7 {
			e = fmt.Errorf("evidence DIR TARGET VERSION COMMIT VERIFICATION [UNSIGNED]")
			break
		}
		previous := ""
		if len(a) == 7 {
			previous = a[6]
		}
		e = evidence(a[1], a[2], a[3], a[4], a[5], previous)
	case "verify-binaries":
		if len(a) != 5 {
			e = fmt.Errorf("verify-binaries DIR TARGET VERSION COMMIT")
			break
		}
		e = verifyBinaries(a[1], a[2], a[3], a[4])
	case "extract":
		if len(a) != 3 {
			e = fmt.Errorf("extract ARCHIVE NEW_DIRECTORY")
			break
		}
		e = extract(a[1], a[2])
	case "assemble":
		if len(a) != 3 {
			e = fmt.Errorf("assemble VERSION COMMIT")
			break
		}
		e = assemble(a[1], a[2])
	case "seal", "verify", "publish":
		if len(a) != 4 {
			e = fmt.Errorf("%s DIRECTORY VERSION COMMIT", a[0])
			break
		}
		switch a[0] {
		case "seal":
			e = seal(a[1], a[2], a[3])
		case "verify":
			e = verify(a[1], a[2], a[3])
		case "publish":
			e = publish(a[1], a[2], a[3])
		}
	default:
		e = fmt.Errorf("unknown command %s", a[0])
	}
	fail(e)
}
