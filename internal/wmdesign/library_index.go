package wmdesign

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

const UnifiedLibrarySchema = "pptxgengo.unified-library-index.v1"

type LibraryIndexOptions struct {
	Bundle       string `json:"bundle"`
	Source       string `json:"source,omitempty"`
	LegacyIndex  string `json:"legacy_index,omitempty"`
	LegacyRoot   string `json:"legacy_root,omitempty"`
	Gallery      string `json:"gallery,omitempty"`
	SlideLibrary string `json:"slide_library,omitempty"`
}

type LibraryIndexPin struct {
	Scope  string `json:"scope"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type LibraryArtifactLink struct {
	Role   string `json:"role"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type LibraryDependency struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	Basis string `json:"basis"`
}

type LibraryEntity struct {
	ID                   string                `json:"id"`
	Namespace            string                `json:"namespace"`
	Kind                 string                `json:"kind"`
	Key                  string                `json:"key"`
	Name                 string                `json:"name"`
	Purpose              string                `json:"purpose,omitempty"`
	Family               string                `json:"family,omitempty"`
	Lifecycle            string                `json:"lifecycle"`
	Revision             int                   `json:"revision,omitempty"`
	SourceRevision       string                `json:"source_revision,omitempty"`
	SourceFile           string                `json:"source_file,omitempty"`
	SourcePointer        string                `json:"source_pointer,omitempty"`
	SourceSHA256         string                `json:"source_sha256,omitempty"`
	Definition           json.RawMessage       `json:"definition"`
	Discovery            LibraryDiscovery      `json:"discovery"`
	Capacity             json.RawMessage       `json:"capacity,omitempty"`
	SupportedAdaptations []string              `json:"supported_adaptations"`
	Dependencies         []LibraryDependency   `json:"dependencies,omitempty"`
	Artifacts            []LibraryArtifactLink `json:"artifacts,omitempty"`
	Template             *LibraryTemplate      `json:"template,omitempty"`
}

type LibraryIndexReport struct {
	Schema                  string              `json:"schema"`
	FinishedSlideTreeSHA256 string              `json:"finished_slide_tree_sha256,omitempty"`
	Path                    string              `json:"path"`
	SourceRevision          string              `json:"source_revision"`
	Counts                  map[string]int      `json:"counts"`
	ProjectionSHA256        string              `json:"projection_sha256"`
	RetrievalText           string              `json:"retrieval_text,omitempty"`
	RetrievalSHA256         string              `json:"retrieval_sha256,omitempty"`
	AssetRegistrySHA256     string              `json:"asset_registry_sha256"`
	Pins                    []LibraryIndexPin   `json:"pins"`
	Options                 LibraryIndexOptions `json:"options"`
	Warnings                []string            `json:"warnings"`
}

func indexJSON(v any) []byte      { b, _ := json.Marshal(v); return b }
func indexDigest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func indexFileDigest(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return indexDigest(b), nil
}
func indexDB(path string, readonly bool) (*sql.DB, error) {
	abs, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	uriPath := filepath.ToSlash(abs)
	// A drive-letter path must be a URI path, not the file URI authority.
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath}
	q := url.Values{}
	if readonly {
		q.Set("mode", "ro")
		q.Add("_pragma", "query_only(1)")
	} else {
		q.Set("mode", "rw")
	}
	u.RawQuery = q.Encode()
	db, e := sql.Open("sqlite", u.String())
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	if e = db.Ping(); e != nil {
		db.Close()
		return nil, e
	}
	return db, nil
}

func projectionHash(entities []LibraryEntity) string {
	h := sha256.New()
	for _, entity := range entities {
		// Normalize typed discovery map numbers exactly as the stored projection is decoded.
		var normalized LibraryEntity
		json.Unmarshal(indexJSON(entity), &normalized)
		h.Write(indexJSON(normalized))
		h.Write([]byte{0})
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func BuildLibraryIndex(path string, options LibraryIndexOptions) (LibraryIndexReport, error) {
	report := LibraryIndexReport{Schema: UnifiedLibrarySchema, Path: path, Counts: map[string]int{}, Options: options, Warnings: []string{}}
	if path == "" {
		return report, fmt.Errorf("index.output_required")
	}
	for _, target := range []*string{&options.Bundle, &options.Source, &options.LegacyIndex, &options.LegacyRoot, &options.Gallery, &options.SlideLibrary} {
		if *target != "" {
			absolute, e := filepath.Abs(*target)
			if e != nil {
				return report, e
			}
			*target = absolute
		}
	}
	if options.LegacyIndex != "" && options.LegacyRoot == "" {
		options.LegacyRoot = filepath.Dir(filepath.Dir(options.LegacyIndex))
	}
	report.Options = options
	s, e := Load(options.Bundle, options.Source)
	if e != nil {
		return report, e
	}
	entities, e := collectLibraryEntities(s, options)
	if e != nil {
		return report, e
	}
	if options.SlideLibrary != "" {
		extra, pins, fingerprint, err := collectFinishedSlides(s, options)
		if err != nil {
			return report, err
		}
		entities = append(entities, extra...)
		report.Pins = append(report.Pins, pins...)
		report.FinishedSlideTreeSHA256 = fingerprint
	}
	if options.LegacyIndex != "" {
		legacy, e := collectLegacyEntities(options.LegacyIndex, options.LegacyRoot)
		if e != nil {
			return report, e
		}
		entities = append(entities, legacy...)
	} else {
		report.Warnings = append(report.Warnings, "Legacy projection was not supplied; this index contains modern entities only.")
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID })
	seen := map[string]bool{}
	for _, entity := range entities {
		if seen[entity.ID] {
			return report, fmt.Errorf("index.duplicate_entity: %s", entity.ID)
		}
		seen[entity.ID] = true
		report.Counts[entity.Kind]++
	}
	report.SourceRevision = s.Revision
	report.ProjectionSHA256 = projectionHash(entities)
	report.RetrievalText = LibraryRetrievalTextVersion
	report.RetrievalSHA256 = retrievalProjectionHash(entities)
	report.AssetRegistrySHA256 = AssetRegistryFingerprint()
	for _, file := range s.Files {
		report.Pins = append(report.Pins, LibraryIndexPin{Scope: "source", Path: file.Path, SHA256: file.SHA256})
	}
	for _, file := range []string{"bundle.json", "inventory.json"} {
		h, e := indexFileDigest(filepath.Join(options.Bundle, file))
		if e != nil {
			return report, e
		}
		report.Pins = append(report.Pins, LibraryIndexPin{Scope: "bundle", Path: file, SHA256: h})
	}
	if options.LegacyIndex != "" {
		h, e := indexFileDigest(options.LegacyIndex)
		if e != nil {
			return report, e
		}
		report.Pins = append(report.Pins, LibraryIndexPin{Scope: "legacy", Path: "index.sqlite", SHA256: h})
	}
	if options.LegacyIndex != "" {
		for _, entity := range entities {
			if entity.Namespace == "legacy" && entity.SourceSHA256 != "" {
				file, e := indexRelative(options.LegacyRoot, entity.SourceFile)
				if e != nil {
					return report, e
				}
				if _, e = os.Stat(file); e == nil {
					report.Pins = append(report.Pins, LibraryIndexPin{Scope: "legacy-resource", Path: entity.SourceFile, SHA256: entity.SourceSHA256})
				} else if !os.IsNotExist(e) {
					return report, e
				}
			}
		}
	}
	if options.Gallery != "" {
		h, e := indexFileDigest(filepath.Join(options.Gallery, "design-system", "index.json"))
		if e != nil {
			return report, e
		}
		report.Pins = append(report.Pins, LibraryIndexPin{Scope: "gallery", Path: "design-system/index.json", SHA256: h})
	}
	if e = os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return report, e
	}
	placeholder, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return report, fmt.Errorf("index.new_output_required: %w", e)
	}
	placeholder.Close()
	success := false
	defer func() {
		if !success {
			os.Remove(path)
		}
	}()
	db, e := indexDB(path, false)
	if e != nil {
		return report, e
	}
	defer db.Close()
	tx, e := db.BeginTx(context.Background(), nil)
	if e != nil {
		return report, e
	}
	defer tx.Rollback()
	_, e = tx.Exec(`CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT NOT NULL);
CREATE TABLE entities(id TEXT PRIMARY KEY,namespace TEXT NOT NULL,kind TEXT NOT NULL,key TEXT NOT NULL,name TEXT NOT NULL,purpose TEXT NOT NULL,family TEXT NOT NULL,lifecycle TEXT NOT NULL,json TEXT NOT NULL);
CREATE INDEX entity_kind ON entities(kind,namespace,lifecycle);
CREATE TABLE dimensions(entity_id TEXT NOT NULL,dimension TEXT NOT NULL,value TEXT NOT NULL,PRIMARY KEY(entity_id,dimension,value));
CREATE INDEX dimension_value ON dimensions(dimension,value,entity_id);
CREATE TABLE content_groups(entity_id TEXT NOT NULL,pointer TEXT NOT NULL,role TEXT NOT NULL,item_count INTEGER NOT NULL,scope TEXT NOT NULL,json TEXT NOT NULL);
CREATE TABLE dependencies(entity_id TEXT NOT NULL,target_id TEXT NOT NULL,role TEXT NOT NULL,basis TEXT NOT NULL);
CREATE VIEW artifacts AS SELECT e.id AS entity_id,json_extract(a.value,'$.role') AS role,json_extract(a.value,'$.path') AS path,json_extract(a.value,'$.sha256') AS sha256 FROM entities e,json_each(e.json,'$.artifacts') a;
CREATE VIEW content_zones AS SELECT e.id AS entity_id,json_extract(z.value,'$.source_pointer') AS pointer,json_extract(z.value,'$.component_type') AS component_type,json_extract(z.value,'$.role') AS role,json_extract(z.value,'$.source_bounds') AS bounds_json FROM entities e,json_each(e.json,'$.discovery.zones') z;
CREATE VIEW content_slots AS SELECT e.id AS entity_id,json_extract(s.value,'$.name') AS name,json_extract(s.value,'$.source_pointer') AS pointer,json_extract(s.value,'$.kind') AS kind,coalesce(json_extract(s.value,'$.allow_empty'),0) AS allow_empty FROM entities e,json_each(e.json,'$.template.slots') s;
` + libraryFTSSchema)
	if e != nil {
		return report, e
	}
	insert, e := tx.Prepare("INSERT INTO entities VALUES(?,?,?,?,?,?,?,?,?)")
	if e != nil {
		return report, e
	}
	defer insert.Close()
	for _, entity := range entities {
		if _, e = insert.Exec(entity.ID, entity.Namespace, entity.Kind, entity.Key, entity.Name, entity.Purpose, entity.Family, entity.Lifecycle, string(indexJSON(entity))); e != nil {
			return report, e
		}
		if _, e = tx.Exec("INSERT INTO entity_fts VALUES(?,?,?,?)", entity.ID, entity.Name, entity.Purpose, libraryRetrievalBody(entity)); e != nil {
			return report, e
		}
		for dimension, values := range map[string][]string{"role": entity.Discovery.ContentRoles, "structure": entity.Discovery.Structures, "visual-form": entity.Discovery.VisualForms} {
			for _, value := range values {
				if _, e = tx.Exec("INSERT OR IGNORE INTO dimensions VALUES(?,?,?)", entity.ID, dimension, value); e != nil {
					return report, e
				}
			}
		}
		for _, group := range entity.Discovery.Groups {
			if _, e = tx.Exec("INSERT INTO content_groups VALUES(?,?,?,?,?,?)", entity.ID, group.SourcePointer, group.Role, group.ExactCount, group.Scope, string(indexJSON(group))); e != nil {
				return report, e
			}
		}
		for _, dependency := range entity.Dependencies {
			if _, e = tx.Exec("INSERT INTO dependencies VALUES(?,?,?,?)", entity.ID, dependency.ID, dependency.Role, dependency.Basis); e != nil {
				return report, e
			}
		}
	}
	if _, e = tx.Exec("INSERT INTO meta VALUES('schema',?),('report',?)", UnifiedLibrarySchema, string(indexJSON(report))); e != nil {
		return report, e
	}
	if e = tx.Commit(); e != nil {
		return report, e
	}
	success = true
	return report, nil
}

type LibraryIndex struct {
	db      *sql.DB
	Report  LibraryIndexReport
	Options LibraryIndexOptions
}

func (index *LibraryIndex) Close() error { return index.db.Close() }

func OpenLibraryIndex(path string, overrides LibraryIndexOptions) (*LibraryIndex, error) {
	db, e := indexDB(path, true)
	if e != nil {
		return nil, e
	}
	fail := func(e error) (*LibraryIndex, error) { db.Close(); return nil, e }
	var raw, schema string
	if e = db.QueryRow("SELECT value FROM meta WHERE key='schema'").Scan(&schema); e != nil || schema != UnifiedLibrarySchema {
		return fail(fmt.Errorf("index.unsupported_schema: %s: %v", schema, e))
	}
	if e = db.QueryRow("SELECT value FROM meta WHERE key='report'").Scan(&raw); e != nil {
		return fail(e)
	}
	var report LibraryIndexReport
	if e = json.Unmarshal([]byte(raw), &report); e != nil {
		return fail(e)
	}
	options := report.Options
	if overrides.Bundle != "" {
		options.Bundle = overrides.Bundle
	}
	if overrides.Source != "" {
		options.Source = overrides.Source
	}
	if overrides.LegacyIndex != "" {
		options.LegacyIndex = overrides.LegacyIndex
	}
	if overrides.LegacyRoot != "" {
		options.LegacyRoot = overrides.LegacyRoot
	}
	if options.LegacyIndex != "" && options.LegacyRoot == "" {
		options.LegacyRoot = filepath.Dir(filepath.Dir(options.LegacyIndex))
	}
	if overrides.Gallery != "" {
		options.Gallery = overrides.Gallery
	}
	if overrides.SlideLibrary != "" {
		options.SlideLibrary = overrides.SlideLibrary
	}
	if options.SlideLibrary != "" {
		_, fingerprint, err := readFinishedLibrary(options.SlideLibrary)
		if err != nil {
			return fail(err)
		}
		if report.FinishedSlideTreeSHA256 == "" || report.FinishedSlideTreeSHA256 != fingerprint {
			return fail(fmt.Errorf("index.finished_slide_library_changed: rebuild the library index"))
		}
	} else if report.FinishedSlideTreeSHA256 != "" {
		return fail(fmt.Errorf("index.finished_slide_library_missing"))
	}
	s, e := Load(options.Bundle, options.Source)
	if e != nil {
		return fail(e)
	}
	if s.Revision != report.SourceRevision {
		return fail(fmt.Errorf("index.stale_source_revision"))
	}
	sourceResolver := newIndexSourcePinResolver(s)
	for _, pin := range report.Pins {
		var path string
		switch pin.Scope {
		case "source":
			path, e = sourceResolver.resolve(pin)
		case "bundle":
			path, e = indexRelative(options.Bundle, pin.Path)
		case "legacy":
			path = options.LegacyIndex
		case "legacy-resource":
			path, e = indexRelative(options.LegacyRoot, pin.Path)
		case "gallery":
			path, e = indexRelative(options.Gallery, pin.Path)
		case "finished-slide":
			path, e = indexRelative(options.SlideLibrary, pin.Path)
		default:
			return fail(fmt.Errorf("index.unknown_pin_scope"))
		}
		if e != nil {
			return fail(e)
		}
		h, e := indexFileDigest(path)
		if e != nil {
			return fail(e)
		}
		if h != pin.SHA256 {
			return fail(fmt.Errorf("index.stale_input: %s/%s", pin.Scope, pin.Path))
		}
	}
	if AssetRegistryFingerprint() != report.AssetRegistrySHA256 {
		return fail(fmt.Errorf("index.stale_asset_registry"))
	}
	index := &LibraryIndex{db: db, Report: report, Options: options}
	entities, e := index.entities("", nil)
	if e != nil {
		return fail(e)
	}
	if projectionHash(entities) != report.ProjectionSHA256 {
		return fail(fmt.Errorf("index.projection_integrity_failed"))
	}
	if report.RetrievalText != "" {
		if report.RetrievalText != LibraryRetrievalTextVersion || report.RetrievalSHA256 != retrievalProjectionHash(entities) {
			return fail(fmt.Errorf("index.stale_retrieval_projection: rebuild the library index"))
		}
		if e = index.verifyRetrievalProjection(entities); e != nil {
			return fail(e)
		}
	}
	return index, nil
}

func (index *LibraryIndex) entities(where string, args []any) ([]LibraryEntity, error) {
	query := "SELECT id,namespace,kind,key,name,purpose,family,lifecycle,json FROM entities"
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY id"
	rows, e := index.db.Query(query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []LibraryEntity{}
	for rows.Next() {
		var raw, id, namespace, kind, key, name, purpose, family, lifecycle string
		if e = rows.Scan(&id, &namespace, &kind, &key, &name, &purpose, &family, &lifecycle, &raw); e != nil {
			return nil, e
		}
		var entity LibraryEntity
		if e = json.Unmarshal([]byte(raw), &entity); e != nil {
			return nil, e
		}
		if id != entity.ID || namespace != entity.Namespace || kind != entity.Kind || key != entity.Key || name != entity.Name || purpose != entity.Purpose || family != entity.Family || lifecycle != entity.Lifecycle {
			return nil, fmt.Errorf("index.projection_columns_integrity_failed: %s", id)
		}
		out = append(out, entity)
	}
	return out, rows.Err()
}

// Discovery does not need source examples, binding contracts or text zones.
// Leave those in the verified SQLite projection for Inspect/fit, rather than
// decoding tens of megabytes of source scene objects on every query.
func (index *LibraryIndex) discoveryEntities(where string, args []any) ([]LibraryEntity, error) {
	return index.discoveryProjection(where, args, false)
}

func (index *LibraryIndex) embeddingEntities(where string, args []any) ([]LibraryEntity, error) {
	return index.discoveryProjection(where, args, true)
}

func (index *LibraryIndex) discoveryProjection(where string, args []any, authoring bool) ([]LibraryEntity, error) {
	template := `json_object('uses',json_extract(json,'$.template.uses'))`
	if authoring {
		template = `json_object('uses',json_extract(json,'$.template.uses'),'authoring',json_extract(json,'$.template.authoring'))`
	}
	query := `SELECT json_set(json_remove(json,'$.definition','$.template','$.discovery.zones'),'$.template',` + template + `) FROM entities`
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY id"
	rows, err := index.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	entities := []LibraryEntity{}
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		var entity LibraryEntity
		if err = json.Unmarshal([]byte(raw), &entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, rows.Err()
}

func (index *LibraryIndex) Inspect(id string) (LibraryEntity, error) {
	entities, e := index.entities("id=? OR (namespace='wmds' AND key=?)", []any{id, id})
	if e != nil {
		return LibraryEntity{}, e
	}
	if len(entities) != 1 {
		return LibraryEntity{}, fmt.Errorf("index.entity_not_unique_or_missing: %s", id)
	}
	entity := entities[0]
	if entity.Namespace == "legacy" && strings.HasPrefix(entity.ID, "legacy/inventory/") {
		var reference struct {
			Table  string `json:"table"`
			RowID  string `json:"row_id"`
			SHA256 string `json:"record_sha256"`
		}
		if e = json.Unmarshal(entity.Definition, &reference); e != nil {
			return LibraryEntity{}, e
		}
		if reference.Table != "inventory" || reference.RowID != entity.Key || reference.SHA256 == "" {
			return LibraryEntity{}, fmt.Errorf("index.invalid_legacy_reference")
		}
		db, e := indexDB(index.Options.LegacyIndex, true)
		if e != nil {
			return LibraryEntity{}, e
		}
		defer db.Close()
		var raw string
		if e = db.QueryRow("SELECT COALESCE(json,'{}') FROM inventory WHERE id=?", reference.RowID).Scan(&raw); e != nil {
			return LibraryEntity{}, e
		}
		if indexDigest([]byte(raw)) != reference.SHA256 || !json.Valid([]byte(raw)) {
			return LibraryEntity{}, fmt.Errorf("index.legacy_record_drift: %s", entity.ID)
		}
		entity.Definition = json.RawMessage(raw)
		entity.SourceFile = index.Options.LegacyIndex
	}
	return entity, nil
}
