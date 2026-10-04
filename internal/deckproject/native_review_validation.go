package deckproject

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/nativeexport"
)

func nativeAttachmentID(a NativeAttachment) string {
	return digest(canonical(map[string]any{"schema": a.Schema, "manifest": a.RenderManifestSHA256, "decisions": a.Decisions, "build": a.BuildID, "source": a.SourceSemanticSHA256, "pptx": a.PPTXSHA256, "slides": a.SlideIDs, "files": a.Files}))
}

func validateNativeAttachment(directory string, a NativeAttachment) (nativeexport.Receipt, error) {
	var manifest nativeexport.Receipt
	if a.Schema != "pptxgengo.native-review-attachment.v1" || a.ID != nativeAttachmentID(a) || filepath.Base(directory) != a.ID {
		return manifest, fmt.Errorf("native attachment identity drift")
	}
	if _, err := time.Parse(time.RFC3339Nano, a.Created); err != nil {
		return manifest, fmt.Errorf("invalid native attachment creation time")
	}
	raw, err := os.ReadFile(filepath.Join(directory, "render-manifest.json"))
	if err != nil {
		return manifest, err
	}
	if digest(raw) != a.RenderManifestSHA256 {
		return manifest, fmt.Errorf("native attachment render manifest drift")
	}
	manifest, err = nativeexport.VerifyReceipt(raw)
	if err != nil {
		return manifest, err
	}
	if !strings.HasPrefix(manifest.Renderer, "Microsoft PowerPoint (local native PDF)") || manifest.Source.SHA256 != a.PPTXSHA256 || manifest.Pages < 1 || manifest.Slides < manifest.Pages || len(a.SlideIDs) != manifest.Pages {
		return manifest, fmt.Errorf("native attachment receipt identity or page coverage drift")
	}
	ids := map[string]bool{}
	for _, id := range a.SlideIDs {
		if !stableID.MatchString(id) || ids[id] {
			return manifest, fmt.Errorf("invalid native attachment slide identity")
		}
		ids[id] = true
	}
	for id, d := range a.Decisions {
		if !ids[id] || (d.Status != "reviewed" && d.Status != "accepted" && d.Status != "issues_found") || strings.TrimSpace(d.Reviewer) == "" {
			return manifest, fmt.Errorf("invalid native visual decision for %s", id)
		}
	}
	expected := map[string]string{"render-manifest.json": digest(raw)}
	artifacts := append([]nativeexport.Artifact(nil), manifest.PNGs...)
	if manifest.PDF != nil {
		artifacts = append(artifacts, *manifest.PDF)
	}
	if manifest.ContactSheet != nil {
		artifacts = append(artifacts, *manifest.ContactSheet)
	}
	if manifest.PDF == nil && len(manifest.PNGs) == 0 {
		return manifest, fmt.Errorf("native attachment has no page artifacts")
	}
	for _, artifact := range artifacts {
		if _, duplicate := expected[artifact.Path]; duplicate {
			return manifest, fmt.Errorf("duplicate native artifact path")
		}
		expected[artifact.Path] = artifact.SHA256
	}
	if !reflect.DeepEqual(a.Files, expected) {
		return manifest, fmt.Errorf("native attachment artifact inventory drift")
	}
	for relative, hash := range expected {
		path, e := SafePath(directory, relative)
		if e != nil {
			return manifest, e
		}
		data, e := os.ReadFile(path)
		if e != nil || digest(data) != hash {
			return manifest, fmt.Errorf("native review artifact drift: %s", relative)
		}
	}
	if len(manifest.PNGs) > 0 && len(manifest.PNGs) != manifest.Pages {
		return manifest, fmt.Errorf("native PNG coverage count is inconsistent")
	}
	if len(manifest.PageMappings) > 0 {
		if len(manifest.PageMappings) != manifest.Pages {
			return manifest, fmt.Errorf("native page mapping count is inconsistent")
		}
		previous := 0
		for index, mapping := range manifest.PageMappings {
			if mapping.Page != index+1 || mapping.SourceSlide <= previous || mapping.SourceSlide > manifest.Slides {
				return manifest, fmt.Errorf("native page mapping order is inconsistent")
			}
			previous = mapping.SourceSlide
			if mapping.PNG != "" {
				if index >= len(manifest.PNGs) || manifest.PNGs[index].Path != mapping.PNG {
					return manifest, fmt.Errorf("native mapped image is absent from verified page artifacts")
				}
			}
		}
	}
	return manifest, nil
}
