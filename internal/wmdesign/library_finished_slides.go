package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/finishedslide"
)

type finishedRevision struct {
	path     string
	manifest finishedslide.Manifest
}

// A library root contains closed revision directories, without loose files or
// links. The fingerprint includes every revision, so additions require reindexing.
func readFinishedLibrary(root string) ([]finishedRevision, string, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, "", fmt.Errorf("index.finished_slide_root_invalid")
	}
	entries := map[string]string{}
	revisions := []finishedRevision{}
	count := 0
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		count++
		if count > 40000 {
			return fmt.Errorf("index.finished_slide_library_too_large")
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("index.finished_slide_symlink: %s", relative)
		}
		if !entry.IsDir() {
			return fmt.Errorf("index.finished_slide_loose_file: %s", relative)
		}
		if _, err := os.Lstat(filepath.Join(path, "manifest.json")); os.IsNotExist(err) {
			entries["directory/"+relative] = ""
			return nil
		} else if err != nil {
			return err
		}
		manifest, err := finishedslide.Read(path)
		if err != nil {
			return err
		}
		revisions = append(revisions, finishedRevision{path: relative, manifest: manifest})
		if len(revisions) > 2048 {
			return fmt.Errorf("index.finished_slide_revision_limit")
		}
		entries["revision/"+relative] = manifest.RevisionSHA256
		return filepath.SkipDir
	})
	if err != nil {
		return nil, "", err
	}
	if len(revisions) == 0 {
		return nil, "", fmt.Errorf("index.finished_slide_library_empty")
	}
	sort.Slice(revisions, func(i, j int) bool { return revisions[i].path < revisions[j].path })
	return revisions, indexDigest(indexJSON(entries)), nil
}

func collectFinishedSlides(source *Source, options LibraryIndexOptions) ([]LibraryEntity, []LibraryIndexPin, string, error) {
	revisions, fingerprint, err := readFinishedLibrary(options.SlideLibrary)
	if err != nil {
		return nil, nil, "", err
	}
	catalog, err := LibraryCatalogFromSource(source)
	if err != nil {
		return nil, nil, "", err
	}
	templates := map[string]LibraryTemplate{}
	for _, def := range catalog {
		templates[def.Key] = def
	}
	latest := map[string]finishedRevision{}
	seen := map[string]bool{}
	for _, revision := range revisions {
		m := revision.manifest
		key := fmt.Sprintf("%s@%d", m.ID, m.Revision)
		if seen[key] {
			return nil, nil, "", fmt.Errorf("index.finished_slide_duplicate_revision: %s", key)
		}
		seen[key] = true
		if err := validateFinishedIndexSource(options.SlideLibrary, revision); err != nil {
			return nil, nil, "", err
		}
		def, ok := templates[m.Pins.TemplateID]
		if !ok || m.Pins.SourceRevision != source.Revision || m.Pins.TemplateRevision != def.Revision || m.Pins.TemplateSourceSHA256 != def.SourceSHA256 {
			return nil, nil, "", fmt.Errorf("index.finished_slide_template_pins_mismatch: %s", key)
		}
		tree, err := libraryObject(def.RawSlide)
		if err != nil {
			return nil, nil, "", err
		}
		if indexDigest(indexJSON(tree)) != m.Pins.TemplateDefinitionSHA256 {
			return nil, nil, "", fmt.Errorf("index.finished_slide_template_definition_mismatch: %s", key)
		}
		if previous, ok := latest[m.ID]; !ok || previous.manifest.Revision < m.Revision {
			latest[m.ID] = revision
		}
	}
	ids := []string{}
	for id := range latest {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	entities := []LibraryEntity{}
	pins := []LibraryIndexPin{}
	for _, id := range ids {
		revision := latest[id]
		m := revision.manifest
		def := templates[m.Pins.TemplateID]
		discovery := def.Discovery
		discovery.Basis = "Content-complete authored revision; structure inherited from its exact pinned shared template."
		discovery.Capability.ContentAdapter = "finished_slide_independent_copy"
		discovery.Capability.SpecimenReview = "review_metadata_in_manifest_not_native_acceptance"
		e := LibraryEntity{ID: m.ID, Namespace: "curated", Kind: finishedslide.Kind, Key: strings.TrimPrefix(m.ID, "curated/slide/"), Name: m.Name, Purpose: m.Purpose, Family: def.Family, Lifecycle: m.Lifecycle, Revision: m.Revision, SourceRevision: fmt.Sprintf("%s@%d", m.ID, m.Revision), SourceFile: revision.path + "/" + m.Source, Definition: indexJSON(m), Discovery: discovery, Capacity: indexJSON(map[string]any{"basis": "authored_content_not_measured_for_this_deck", "owner": m.Owner, "approval": m.Approval, "reviewed_at": m.ReviewedAt, "valid_until": m.ValidUntil, "revision_sha256": m.RevisionSHA256, "package_path": revision.path, "keywords": m.Keywords, "template_pins": m.Pins}), SupportedAdaptations: []string{"Insert an independent copy with fresh slide/item/asset identities and explicit composition rationale; insertion validates toolchain/template pins and freshness.", "Later copy changes require the destination deck's evidence and review process; library approval does not qualify an inserted deck."}, Dependencies: []LibraryDependency{{ID: "wmds/template/" + def.Key, Role: "shared_template", Basis: "exact_revision_source_and_definition_pins"}}}
		pins = append(pins, LibraryIndexPin{Scope: "finished-slide", Path: revision.path + "/manifest.json"})
		pins[len(pins)-1].SHA256, err = indexFileDigest(filepath.Join(options.SlideLibrary, filepath.FromSlash(pins[len(pins)-1].Path)))
		if err != nil {
			return nil, nil, "", err
		}
		for _, file := range m.Files {
			relative := revision.path + "/" + file.Path
			pins = append(pins, LibraryIndexPin{Scope: "finished-slide", Path: relative, SHA256: file.SHA256})
			if file.Path == m.Source {
				e.SourceSHA256 = file.SHA256
			}
			if file.Role == "preview" {
				e.Artifacts = append(e.Artifacts, LibraryArtifactLink{Role: "source_preview", Path: relative, SHA256: file.SHA256})
			}
		}
		entities = append(entities, e)
	}
	return entities, pins, fingerprint, nil
}

func finishedReuseStatus(entity LibraryEntity) string {
	if entity.Kind != finishedslide.Kind {
		return ""
	}
	var dates struct {
		ReviewedAt string `json:"reviewed_at"`
		ValidUntil string `json:"valid_until"`
	}
	if err := json.Unmarshal(entity.Capacity, &dates); err != nil {
		return "invalid_review_metadata"
	}
	m := finishedslide.Manifest{ReviewedAt: dates.ReviewedAt, ValidUntil: dates.ValidUntil}
	return entity.Lifecycle + ";" + m.Freshness(time.Now()) + ";destination_review_required"
}

// Indexing establishes authored source identity and kind. Insertion additionally
// validates the complete values contract, assets and the destination compiler.
func validateFinishedIndexSource(root string, revision finishedRevision) error {
	m := revision.manifest
	var size int64
	for _, f := range m.Files {
		if f.Path == m.Source {
			size = f.Bytes
		}
	}
	if size > 16<<20 {
		return fmt.Errorf("index.finished_slide_source_too_large")
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(revision.path+"/"+m.Source)))
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var node, extra yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return err
	}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("index.finished_slide_source_documents_invalid")
	}
	var check func(*yaml.Node, int) error
	check = func(n *yaml.Node, depth int) error {
		if depth > 100 || n.Kind == yaml.AliasNode || n.Anchor != "" {
			return fmt.Errorf("index.finished_slide_source_alias_or_depth_invalid")
		}
		for _, child := range n.Content {
			if err := check(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := check(&node, 0); err != nil {
		return err
	}
	var source map[string]any
	if err := node.Decode(&source); err != nil {
		return err
	}
	template, _ := source["template"].(map[string]any)
	_, hasValues := source["values"].(map[string]any)
	id, _ := source["id"].(string)
	if id == "" || source["content_kind"] != "supplied_content" || template["scope"] != "shared" || template["id"] != m.Pins.TemplateID || !hasValues {
		return fmt.Errorf("index.finished_slide_source_identity_or_kind_invalid")
	}
	return nil
}
