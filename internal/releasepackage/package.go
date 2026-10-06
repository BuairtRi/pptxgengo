// Package releasepackage creates portable Windows tester ZIPs without modifying
// the installed release. Replacements are explicit and hashes are regenerated.
package releasepackage

import (
	"archive/zip"
	"compress/flate"
	"crypto/sha256"
	"debug/pe"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed install-windows.ps1
var installer []byte

//go:embed smoke-test-windows.ps1
var smokeTest []byte

//go:embed WINDOWS.md
var readme []byte

//go:embed THIRD-PARTY-NOTICES.txt
var thirdPartyNotices []byte

//go:embed guides/*.html
var colleagueGuides embed.FS

type Options struct {
	Release, Binaries, Skill, Out, Architecture, Version string
	WithoutPhotos                                        bool
}
type Report struct {
	Path                    string `json:"path"`
	Files                   int    `json:"files"`
	Bytes                   int64  `json:"bytes"`
	Target                  string `json:"target"`
	OriginalPhotosIncluded  bool   `json:"original_photos_included"`
	WindowsRuntimeQualified bool   `json:"windows_runtime_qualified"`
}
type inputFile struct {
	path     string
	data     []byte
	expected string
	mode     os.FileMode
}

func validateArchivePath(path string) error {
	if path == "" || strings.HasPrefix(path, "/") || strings.ContainsAny(path, `\:*?"<>|`) {
		return fmt.Errorf("unsafe Windows package path %q", path)
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("unsafe Windows package path %q", path)
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || (len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9') {
			return fmt.Errorf("reserved Windows package path %q", path)
		}
	}
	return nil
}
func addTree(files map[string]inputFile, root, prefix string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("package requires regular files: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[prefix+filepath.ToSlash(rel)] = inputFile{path: path, mode: info.Mode()}
		return nil
	})
}
func Create(o Options) (_ Report, err error) {
	report := Report{Path: o.Out, Target: "windows-" + o.Architecture, OriginalPhotosIncluded: !o.WithoutPhotos}
	if o.Architecture != "amd64" && o.Architecture != "arm64" {
		return report, fmt.Errorf("Windows architecture must be amd64 or arm64")
	}
	if o.Release == "" || o.Binaries == "" || o.Skill == "" || o.Out == "" || o.Version == "" {
		return report, fmt.Errorf("release, binaries, skill, out and version are required")
	}
	if !strings.HasSuffix(strings.ToLower(o.Out), ".zip") {
		return report, fmt.Errorf("--out must be a new .zip file")
	}
	var manifest map[string]any
	data, err := os.ReadFile(filepath.Join(o.Release, "release-manifest.json"))
	if err != nil {
		return report, err
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return report, err
	}
	if manifest["schema"] != "pptxgengo.local-release-manifest.v1" {
		return report, fmt.Errorf("unsupported release manifest")
	}
	oldHashes, ok := manifest["files_sha256"].(map[string]any)
	if !ok {
		return report, fmt.Errorf("release manifest has no file hashes")
	}
	files := map[string]inputFile{}
	for rel, value := range oldHashes {
		if err = validateArchivePath(rel); err != nil {
			return report, err
		}
		if strings.HasPrefix(rel, "bin/") || strings.HasPrefix(rel, "skills/west-monroe-presentations/") || (o.WithoutPhotos && strings.HasPrefix(rel, "branding/West Monroe Photos/")) {
			continue
		}
		expected, ok := value.(string)
		if !ok || len(expected) != 64 {
			return report, fmt.Errorf("invalid release file hash: %s", rel)
		}
		path := filepath.Join(o.Release, filepath.FromSlash(rel))
		info, err := os.Lstat(path)
		if err != nil {
			return report, err
		}
		if !info.Mode().IsRegular() {
			return report, fmt.Errorf("release file is not regular: %s", rel)
		}
		files[rel] = inputFile{path: path, expected: expected, mode: info.Mode()}
	}
	for _, tool := range []string{"pptxgengo", "pptxdesign", "wmdsdocs"} {
		path := filepath.Join(o.Binaries, tool+".exe")
		binary, err := pe.Open(path)
		if err != nil {
			return report, fmt.Errorf("not a Windows executable: %s: %w", path, err)
		}
		machine := uint16(pe.IMAGE_FILE_MACHINE_AMD64)
		if o.Architecture == "arm64" {
			machine = pe.IMAGE_FILE_MACHINE_ARM64
		}
		actual := binary.Machine
		binary.Close()
		if actual != machine {
			return report, fmt.Errorf("wrong Windows executable architecture: %s", path)
		}
		files["bin/"+tool+".exe"] = inputFile{path: path, mode: 0755}
	}
	if err = addTree(files, o.Skill, "skills/west-monroe-presentations/"); err != nil {
		return report, err
	}
	files["VERSION"] = inputFile{data: []byte(o.Version + "\n"), mode: 0644}
	files["release/VERSION"] = files["VERSION"]
	files["install-windows.ps1"] = inputFile{data: installer, mode: 0644}
	files["smoke-test-windows.ps1"] = inputFile{data: smokeTest, mode: 0644}
	files["WINDOWS.md"] = inputFile{data: readme, mode: 0644}
	files["THIRD-PARTY-NOTICES.txt"] = inputFile{data: thirdPartyNotices, mode: 0644}
	guides, err := colleagueGuides.ReadDir("guides")
	if err != nil {
		return report, err
	}
	for _, guide := range guides {
		data, err := colleagueGuides.ReadFile("guides/" + guide.Name())
		if err != nil {
			return report, err
		}
		files["guides/"+guide.Name()] = inputFile{data: data, mode: 0644}
	}
	keys := make([]string, 0, len(files))
	casePaths := map[string]string{}
	for rel := range files {
		if err = validateArchivePath(rel); err != nil {
			return report, err
		}
		folded := strings.ToLower(rel)
		if prior, exists := casePaths[folded]; exists {
			return report, fmt.Errorf("Windows case-insensitive path collision: %s and %s", prior, rel)
		}
		casePaths[folded] = rel
		keys = append(keys, rel)
	}
	sort.Strings(keys)
	if err = os.MkdirAll(filepath.Dir(o.Out), 0755); err != nil {
		return report, err
	}
	output, err := os.OpenFile(o.Out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return report, fmt.Errorf("package output must be new: %w", err)
	}
	defer func() {
		output.Close()
		if err != nil {
			os.Remove(o.Out)
		}
	}()
	writer := zip.NewWriter(output)
	defer writer.Close()
	writer.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) { return flate.NewWriter(w, flate.BestSpeed) })
	rootName := "pptxgengo-windows-" + o.Architecture
	hashes := map[string]string{}
	for _, rel := range keys {
		file := files[rel]
		header := &zip.FileHeader{Name: rootName + "/" + rel, Method: zip.Deflate}
		header.SetMode(file.mode)
		switch strings.ToLower(filepath.Ext(rel)) {
		case ".png", ".jpg", ".jpeg", ".pptx", ".zip":
			header.Method = zip.Store
		}
		destination, err := writer.CreateHeader(header)
		if err != nil {
			return report, err
		}
		sum := sha256.New()
		if file.path == "" {
			_, err = io.MultiWriter(destination, sum).Write(file.data)
		} else {
			input, openErr := os.Open(file.path)
			if openErr != nil {
				return report, openErr
			}
			_, err = io.Copy(io.MultiWriter(destination, sum), input)
			input.Close()
		}
		if err != nil {
			return report, err
		}
		digest := hex.EncodeToString(sum.Sum(nil))
		if file.expected != "" && file.expected != digest {
			return report, fmt.Errorf("installed release file drift: %s", rel)
		}
		hashes[rel] = digest
	}
	manifest["version"] = o.Version
	manifest["target_os"], manifest["target_arch"] = "windows", o.Architecture
	manifest["windows_runtime_qualified"] = false
	manifest["windows_native_render_status"] = "experimental; requires desktop PowerPoint smoke test"
	manifest["original_photographs_included"] = !o.WithoutPhotos
	manifest["files_sha256"], manifest["file_count"] = hashes, len(hashes)
	data, err = json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return report, err
	}
	entry, err := writer.Create(rootName + "/release-manifest.json")
	if err != nil {
		return report, err
	}
	if _, err = entry.Write(append(data, '\n')); err != nil {
		return report, err
	}
	if err = writer.Close(); err != nil {
		return report, err
	}
	if err = output.Sync(); err != nil {
		return report, err
	}
	info, err := output.Stat()
	if err != nil {
		return report, err
	}
	report.Files, report.Bytes = len(hashes), info.Size()
	if err = output.Close(); err != nil {
		return report, err
	}
	return report, nil
}
