package deckproject

import (
	"archive/zip"
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ExportOptions struct {
	Mode   string
	Out    string
	Bundle string
}
type ExportReceipt struct {
	Schema                  string            `json:"schema"`
	Mode                    string            `json:"mode"`
	BuildID                 string            `json:"build_id"`
	SHA256                  string            `json:"sha256"`
	Files                   map[string]string `json:"files"`
	ContainsPrivateMaterial bool              `json:"contains_private_material"`
}

func Export(p *Project, opts ExportOptions) (ExportReceipt, error) {
	r := ExportReceipt{Schema: "pptxgengo.deck-export-receipt.v1", Mode: opts.Mode, Files: map[string]string{}}
	if opts.Mode != "maintainer" && opts.Mode != "client" && opts.Mode != "offline" && opts.Mode != "reviewer" {
		return r, fmt.Errorf("export mode must be maintainer, client, offline or reviewer")
	}
	if opts.Out == "" {
		return r, fmt.Errorf("export requires new --out ZIP path")
	}
	s, e := Status(p)
	if e != nil {
		return r, e
	}
	if e = protectBaseline(p, s); e != nil {
		return r, e
	}
	if s.CurrentBuild == "" || s.SemanticSHA256 != digest(p.Canonical) {
		return r, fmt.Errorf("export requires current generated build")
	}
	currentDeps, e := dependencies(p)
	if e != nil {
		return r, e
	}
	if !reflectEqual(currentDeps, s.Dependencies) {
		return r, fmt.Errorf("project dependencies changed since build; rebuild before export")
	}
	r.BuildID = s.CurrentBuild
	builddir, e := SafePath(p.Root, "builds/"+s.CurrentBuild)
	if e != nil {
		return r, e
	}
	files := map[string][]byte{}
	add := func(name, path string) error {
		if _, e := SafePath(p.Root, "deck.yaml"); e != nil {
			return e
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		files[filepath.ToSlash(name)] = b
		return nil
	}
	if opts.Mode == "client" || opts.Mode == "reviewer" {
		if e = add("deck.pptx", filepath.Join(builddir, "deck.pptx")); e != nil {
			return r, e
		}
		if _, e = os.Stat(filepath.Join(builddir, "deck.pdf")); e == nil {
			return r, fmt.Errorf("PDF is not yet a recorded build artifact; import/qualification is required before packaging")
		}
		if opts.Mode == "reviewer" {
			for _, f := range []string{"receipt.json", "layout-report.json", "object-map.json"} {
				if e = add(f, filepath.Join(builddir, f)); e != nil {
					return r, e
				}
			}
			slides := []map[string]any{}
			for _, s := range p.Document.Slides {
				slides = append(slides, map[string]any{"id": s.ID, "template": s.Template, "values": s.Values})
			}
			files["review.json"] = canonical(map[string]any{"schema": "pptxgengo.deck-review-packet.v1", "deck_id": p.Document.ID, "build_id": s.CurrentBuild, "slides": slides, "review_stages": []string{"source_accuracy", "fit", "native_render", "visual_quality"}, "alternatives": "separate authored sources/builds required; no sample substitution"})
		}
	} else {
		r.ContainsPrivateMaterial = true
		e = filepath.WalkDir(p.Root, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			rel, e := filepath.Rel(p.Root, path)
			if e != nil {
				return e
			}
			if rel == "." {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("export refuses symlink %s", rel)
			}
			if d.IsDir() {
				if strings.HasPrefix(rel, ".git") || rel == "builds" {
					if rel == "builds" {
						return nil
					}
					return filepath.SkipDir
				}
				if strings.HasPrefix(rel, "builds/") && rel != "builds/"+s.CurrentBuild && !strings.HasPrefix(rel, "builds/"+s.CurrentBuild+"/") {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("export refuses nonregular %s", rel)
			}
			if strings.HasPrefix(rel, ".") {
				return nil
			}
			if strings.HasPrefix(rel, "builds/") && !strings.HasPrefix(rel, "builds/"+s.CurrentBuild+"/") {
				return nil
			}
			safe, e := SafePath(p.Root, rel)
			if e != nil {
				return e
			}
			return add(rel, safe)
		})
		if e != nil {
			return r, e
		}
		if opts.Mode == "offline" {
			lock, _, e := ReadLock(p)
			if e != nil {
				return r, e
			}
			if _, e = Check(p, opts.Bundle, lock.Engine); e != nil {
				return r, e
			}
			e = filepath.WalkDir(opts.Bundle, func(path string, d fs.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					return nil
				}
				rel, e := filepath.Rel(opts.Bundle, path)
				if e != nil {
					return e
				}
				safe, e := SafePath(opts.Bundle, filepath.ToSlash(rel))
				if e != nil {
					return e
				}
				return add("runtime/library/wm-design-system/pinned/"+rel, safe)
			})
			if e != nil {
				return r, e
			}
			for rel, want := range lock.RuntimeFiles {
				path, e := SafePath(filepath.Dir(opts.Bundle), rel)
				if e != nil {
					return r, e
				}
				bytes, e := os.ReadFile(path)
				if e != nil {
					return r, e
				}
				if digest(bytes) != want {
					return r, fmt.Errorf("offline runtime dependency drift: %s", rel)
				}
				files["runtime/library/wm-design-system/"+rel] = bytes
			}
			exe, e := os.Executable()
			if e != nil {
				return r, e
			}
			if e = add("runtime/pptxdesign", exe); e != nil {
				return r, e
			}
			data, e := os.ReadFile(filepath.Join(builddir, "deck.pptx"))
			if e != nil {
				return r, e
			}
			allowedHashes := map[string]bool{}
			for _, a := range p.Document.Assets {
				if a.Path != "" {
					path, e := SafePath(p.Root, a.Path)
					if e != nil {
						return r, e
					}
					bytes, e := os.ReadFile(path)
					if e != nil {
						return r, e
					}
					allowedHashes[digest(bytes)] = true
				}
			}
			refs, e := wmdesign.UsedPrimitiveAssetsWithProjectHashes(data, allowedHashes)
			if e != nil {
				return r, e
			}
			branding := os.Getenv("WMDS_BRANDING_ROOT")
			if branding == "" && os.Getenv("PPTXGENGO_RELEASE_ROOT") != "" {
				branding = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "branding")
			}
			if branding == "" {
				home, e := os.UserHomeDir()
				if e != nil {
					return r, e
				}
				branding = filepath.Join(home, "Documents/branding")
			}
			for _, ref := range refs {
				path, archivePath, e := registryResource(branding, ref.Path)
				if e != nil {
					return r, e
				}
				b, e := os.ReadFile(path)
				if e != nil {
					return r, e
				}
				if digest(b) != ref.SHA256 {
					return r, fmt.Errorf("offline registry asset drift: %s", ref.Key)
				}
				files[archivePath] = b
			}
			files["OFFLINE.md"] = []byte("# Offline maintainer package\n\nIncludes private source/context/evidence. Extract into a new directory. Runtime is pinned to its original OS/architecture and executable hash; fonts and calibration are in the bundled library.\n\nFrom the extracted package:\n\n```sh\nchmod +x runtime/pptxdesign\nWMDS_BRANDING_ROOT=\"$PWD/runtime/branding\" runtime/pptxdesign project check --project . --bundle runtime/library/wm-design-system/pinned\n```\n\nUse the engine in toolchain.lock.json explicitly if it differs from the project command default. This package does not install fonts into PowerPoint or contain native PowerPoint/PDF conversion.\n")
		}
	}
	keys := []string{}
	for name, b := range files {
		r.Files[name] = digest(b)
		keys = append(keys, name)
	}
	files["export-manifest.json"] = canonical(map[string]any{"schema": r.Schema, "mode": r.Mode, "build_id": r.BuildID, "contains_private_material": r.ContainsPrivateMaterial, "files": r.Files})
	keys = append(keys, "export-manifest.json")
	sort.Strings(keys)
	out, e := filepath.Abs(opts.Out)
	if e != nil {
		return r, e
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(out))
	if e != nil {
		return r, e
	}
	out = filepath.Join(parent, filepath.Base(out))
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return r, e
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(out)
		}
	}()
	z := zip.NewWriter(f)
	for _, name := range keys {
		if strings.HasPrefix(name, "/") || strings.Contains(name, "..") || strings.Contains(name, "\\") {
			z.Close()
			return r, fmt.Errorf("unsafe export archive name %s", name)
		}
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		h.SetModTime(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC))
		h.SetMode(0644)
		if name == "runtime/pptxdesign" {
			h.SetMode(0755)
		}
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
	}
	b, e := os.ReadFile(out)
	if e != nil {
		return r, e
	}
	r.SHA256 = digest(b)
	ok = true
	return r, nil
}
func reflectEqual(a, b any) bool { return string(canonical(a)) == string(canonical(b)) }
func ReviewPacket(p *Project, out string) (ExportReceipt, error) {
	return Export(p, ExportOptions{Mode: "reviewer", Out: out})
}

func registryResource(root, rel string) (string, string, error) {
	if strings.HasPrefix(rel, "../career/") {
		clean := strings.TrimPrefix(rel, "../")
		path, e := SafePath(filepath.Dir(root), clean)
		return path, "runtime/" + clean, e
	}
	path, e := SafePath(root, rel)
	return path, "runtime/branding/" + rel, e
}
