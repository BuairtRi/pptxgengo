package wmdesign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func collectLibraryEntities(source *Source, options LibraryIndexOptions) ([]LibraryEntity, error) {
	catalog, e := libraryCatalog(source)
	if e != nil {
		return nil, e
	}
	hashes := map[string]string{}
	for _, file := range source.Files {
		hashes[file.Path] = file.SHA256
	}
	var entities []LibraryEntity
	componentIDs := map[string]string{}
	var components struct {
		Components []json.RawMessage `json:"components"`
	}
	if e = json.Unmarshal(source.Components, &components); e != nil {
		return nil, e
	}
	for i, raw := range components.Components {
		obj, e := libraryObject(raw)
		if e != nil {
			return nil, e
		}
		key, _ := obj["id"].(string)
		name, _ := obj["name"].(string)
		kind, _ := obj["layer"].(string)
		if key == "" || name == "" {
			return nil, fmt.Errorf("index.invalid_component_identity")
		}
		if kind == "" {
			kind = "component"
		}
		if kind != "primitive" && kind != "component" && kind != "composite" {
			return nil, fmt.Errorf("index.invalid_component_layer: %s", kind)
		}
		purpose, _ := obj["summary"].(string)
		family, _ := obj["group"].(string)
		entity := LibraryEntity{ID: "wmds/" + kind + "/" + key, Namespace: "wmds", Kind: kind, Key: key, Name: name, Purpose: purpose, Family: family, Lifecycle: "active", SourceRevision: source.Revision, SourceFile: "components/v0/components.json", SourcePointer: fmt.Sprintf("/components/%d", i), SourceSHA256: hashes["components/v0/components.json"], Definition: raw, SupportedAdaptations: []string{"Compose source-declared examples through supported native scene nodes; source option descriptions are advisory, not unrestricted engine support."}}
		componentIDs[key] = entity.ID
		entity.Discovery = componentExampleDiscovery(source.Revision, obj, entity.SourcePointer)
		entity.Capacity = indexJSON(map[string]any{"basis": "source_advisory", "sizing": obj["sizing"], "variants": obj["variants"], "defaults": obj["defaults"]})
		entities = append(entities, entity)
	}
	frames, e := collectFrameEntities(source, hashes)
	if e != nil {
		return nil, e
	}
	entities = append(entities, frames...)
	for _, template := range catalog {
		copy := template
		entity := LibraryEntity{ID: "wmds/template/" + template.Key, Namespace: "wmds", Kind: "template", Key: template.Key, Name: template.Name, Purpose: template.Purpose, Family: template.Family, Lifecycle: template.Status, Revision: template.Revision, SourceRevision: template.SourceRevision, SourceFile: template.SourceFile, SourceSHA256: template.SourceSHA256, Definition: indexJSON(map[string]any{"contract": template, "source_slide": json.RawMessage(template.RawSlide)}), Discovery: template.Discovery, Capacity: indexJSON(map[string]any{"basis": "source_advisory_not_measured", "budget": json.RawMessage(template.AdvisoryBudget), "slot_guidance": json.RawMessage(template.AdvisorySlots), "exact_arrays": template.Arrays, "typed_values": template.ValueSchema}), SupportedAdaptations: append([]string(nil), template.Policy...), Template: &copy}
		for _, use := range template.Uses {
			if id, ok := componentIDs[use]; ok {
				entity.Dependencies = append(entity.Dependencies, LibraryDependency{ID: id, Role: "declared_component", Basis: "source_uses_annotation"})
			}
		}
		frame := template.Discovery.Frame
		rail, footer := frame.Rail, frame.Footer
		if rail == "" {
			rail = "none"
		}
		if footer == "" {
			footer = "compact"
		}
		frameID := "wmds/frame/" + rail + "-" + footer
		if frame.Split != "" {
			frameID += "/split/" + frame.Split
		}
		entity.Dependencies = append(entity.Dependencies, LibraryDependency{ID: frameID, Role: "frame", Basis: "source_slide_chrome"})
		entities = append(entities, entity)
	}
	for _, asset := range PrimitiveAssetCatalog() {
		meta := assetMetadataFor(asset.Key, asset.Path)
		definition := struct {
			PrimitiveAssetReference
			Kind           string              `json:"kind"`
			Tags           []string            `json:"tags"`
			People         string              `json:"people"`
			Industry       []string            `json:"industry,omitempty"`
			Setting        []string            `json:"setting,omitempty"`
			Orientation    string              `json:"orientation,omitempty"`
			SourceMetadata *PhotoMetadata      `json:"source_metadata,omitempty"`
			OriginalFacts  *PhotoOriginalFacts `json:"original_facts,omitempty"`
		}{asset, meta.kind, meta.tags, meta.people, meta.industry, meta.setting, meta.orientation, meta.photoMetadata, registeredPhotoOriginalFacts(asset.Key, asset.Path)}
		metadataBasis := "curated_asset_metadata"
		if meta.photoMetadata != nil {
			metadataBasis = "registered_curated_metadata_enriched_from_existing_branding_sidecar"
		}
		entity := LibraryEntity{ID: "wmds/asset/" + asset.Key, Namespace: "wmds", Kind: "asset", Key: asset.Key, Name: meta.name, Purpose: strings.TrimSpace(meta.description + " " + meta.searchText), Family: meta.kind, Lifecycle: "active", SourceRevision: source.Revision, SourceFile: asset.Path, SourceSHA256: asset.SHA256, Definition: indexJSON(definition), Capacity: indexJSON(map[string]any{"basis": metadataBasis, "tags": meta.tags, "people": meta.people, "industry": meta.industry, "setting": meta.setting, "orientation": meta.orientation}), SupportedAdaptations: []string{"Use the registered asset ID in supported media/icon/artwork fields; crop metadata describes a registered variant. Original bytes are verified when used."}}
		entity.Discovery = LibraryDiscovery{Schema: LibraryDiscoverySchema, Basis: "pinned_native_asset_registry", ContentRoles: []string{"image"}, VisualForms: []string{"image"}, Capability: LibraryCapability{ContentAdapter: "registered_asset", BuildForQuery: "not_executed", SpecimenReview: "not_loaded_by_discovery", ContentEnvelope: "not_applicable", CapacityBasis: "registered_original_not_layout_capacity"}}
		entities = append(entities, entity)
	}
	if options.Gallery != "" {
		if e = collectGalleryArtifacts(entities, source.Revision, options.Gallery); e != nil {
			return nil, e
		}
	}
	return entities, nil
}

func componentExampleDiscovery(revision string, obj map[string]any, pointer string) LibraryDiscovery {
	d := LibraryDiscovery{Schema: LibraryDiscoverySchema, Basis: "derived_from_source_examples_not_reusable_envelope", Capability: LibraryCapability{ContentAdapter: "source_examples_composition", BuildForQuery: "not_executed", SpecimenReview: "not_loaded_by_discovery", ContentEnvelope: "not_established_by_discovery", CapacityBasis: "source_advisory"}}
	roles, structures, forms, types := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	examples, _ := obj["examples"].([]any)
	for i, raw := range examples {
		example, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		nodes, ok := example["nodes"].([]any)
		if !ok {
			continue
		}
		projection := map[string]any{"body": nodes}
		def := LibraryTemplate{TemplateDefinition: TemplateDefinition{SourceRevision: revision}}
		for j, node := range nodes {
			libraryContentWalk(&def, node, fmt.Sprintf("/body/%d", j), fmt.Sprintf("node%02d", j+1), "", libraryProjectionContext{})
		}
		part := libraryDiscovery(def, projection)
		for _, value := range part.ContentRoles {
			roles[value] = true
		}
		for _, value := range part.Structures {
			structures[value] = true
		}
		for _, value := range part.VisualForms {
			forms[value] = true
		}
		for _, value := range part.ComponentTypes {
			types[value] = true
		}
		prefix := fmt.Sprintf("%s/examples/%d/nodes", pointer, i)
		for _, zone := range part.Zones {
			zone.SourcePointer = strings.Replace(zone.SourcePointer, "/body", prefix, 1)
			d.Zones = append(d.Zones, zone)
		}
		for _, group := range part.Groups {
			group.SourcePointer = strings.Replace(group.SourcePointer, "/body", prefix, 1)
			group.Scope = "example"
			group.ScopeBasis = "illustrated_example_count_not_general_cardinality"
			d.Groups = append(d.Groups, group)
		}
		for _, relationship := range part.Relationships {
			for j := range relationship.SourcePointers {
				relationship.SourcePointers[j] = strings.Replace(relationship.SourcePointers[j], "/body", prefix, 1)
			}
			d.Relationships = append(d.Relationships, relationship)
		}
	}
	d.ContentRoles, d.Structures, d.VisualForms, d.ComponentTypes = discoverySorted(roles), discoverySorted(structures), discoverySorted(forms), discoverySorted(types)
	return d
}

func collectFrameEntities(source *Source, hashes map[string]string) ([]LibraryEntity, error) {
	var entities []LibraryEntity
	rails, footers, splits := []string{}, []string{}, []string{""}
	for key := range source.Frames.Rails {
		rails = append(rails, key)
	}
	for key := range source.Frames.Footers {
		footers = append(footers, key)
	}
	for key, raw := range source.Frames.Splits {
		var split map[string]json.RawMessage
		if json.Unmarshal(raw, &split) == nil && split["short"] != nil && split["tall"] != nil {
			splits = append(splits, key)
		}
	}
	sort.Strings(rails)
	sort.Strings(footers)
	sort.Strings(splits)
	for _, rail := range rails {
		for _, footer := range footers {
			for _, split := range splits {
				if split != "" && rail != "none" && rail != "nav" {
					continue
				}
				q := FrameRequest{Rail: rail, Footer: footer, TitleLines: 1, Split: split}
				if rail == "nav" {
					q.Nav = []NavTab{{ID: "section-a", Label: "Section A"}, {ID: "section-b", Label: "Section B"}}
					q.Active = "section-a"
				}
				resolved, e := source.ResolveFrame(q)
				if e != nil {
					return nil, fmt.Errorf("index.frame_resolution: %s/%s/%s: %w", rail, footer, split, e)
				}
				key := rail + "-" + footer
				if split != "" {
					key += "/split/" + split
				}
				definition := map[string]any{"default_request": resolved.Request, "default_resolved": resolved, "source_rail": source.Frames.Rails[rail], "source_footer": source.Frames.Footers[footer], "source_split": source.Frames.Splits[split], "nav_labels": "illustrative geometry only"}
				entity := LibraryEntity{ID: "wmds/frame/" + key, Namespace: "wmds", Kind: "frame", Key: key, Name: rail + " rail · " + footer + " footer", Purpose: "Resolved source frame zones and chrome; content remains authored separately.", Lifecycle: "active", SourceRevision: source.Revision, SourceFile: "frames/v0/frames.json", SourceSHA256: hashes["frames/v0/frames.json"], Definition: indexJSON(definition), SupportedAdaptations: []string{"Select declared rail/footer/split, surface, header/source allocations and navigation through FrameRequest; unsupported combinations are rejected by ResolveFrame."}}
				entity.Discovery = LibraryDiscovery{Schema: LibraryDiscoverySchema, Basis: "source_frame_and_resolved_default_geometry", ContentRoles: []string{"headline", "content"}, Frame: LibraryFrameMetadata{Rail: rail, Footer: footer, Split: split, Nav: rail == "nav", Basis: "pinned_frame_definition"}, Capability: LibraryCapability{ContentAdapter: "resolved_frame", BuildForQuery: "not_executed", SpecimenReview: "not_loaded_by_discovery", ContentEnvelope: "not_layout_content_qualification", CapacityBasis: "resolved_default_zones_not_text_capacity"}}
				entity.Capacity = indexJSON(map[string]any{"basis": "resolved_default_geometry", "body": resolved.Body, "header": resolved.Header, "rail": resolved.Rail, "source": resolved.Source})
				entities = append(entities, entity)
			}
		}
	}
	return entities, nil
}

func collectLegacyEntities(path, root string) ([]LibraryEntity, error) {
	db, e := indexDB(path, true)
	if e != nil {
		return nil, e
	}
	defer db.Close()
	var entities []LibraryEntity
	rows, e := db.Query("SELECT id,COALESCE(kind,''),COALESCE(title,''),COALESCE(body,''),COALESCE(json,'{}'),COALESCE(preference,'') FROM inventory ORDER BY id")
	if e != nil {
		return nil, e
	}
	for rows.Next() {
		var id, kind, name, body, raw, pref string
		if e = rows.Scan(&id, &kind, &name, &body, &raw, &pref); e != nil {
			rows.Close()
			return nil, e
		}
		if !json.Valid([]byte(raw)) {
			rows.Close()
			return nil, fmt.Errorf("index.invalid_legacy_record: %s", id)
		}
		entity := LibraryEntity{ID: "legacy/inventory/" + id, Namespace: "legacy", Kind: kind, Key: id, Name: name, Purpose: body, Lifecycle: "inventory", Definition: indexJSON(map[string]string{"basis": "hash_pinned_original_legacy_record", "table": "inventory", "row_id": id, "record_sha256": indexDigest([]byte(raw))}), SourceFile: path, SourceSHA256: "", SupportedAdaptations: []string{"Legacy inventory discovery only; inspect original capabilities and evidence before execution."}}
		entity.Discovery.Capability = LibraryCapability{ContentAdapter: "inventory_unproven", BuildForQuery: "not_executed", SpecimenReview: "preserved_in_definition", ContentEnvelope: "not_inferred"}
		if pref == "avoid" {
			entity.Lifecycle = "avoid"
		}
		entities = append(entities, entity)
	}
	if e = rows.Err(); e != nil {
		rows.Close()
		return nil, e
	}
	rows.Close()
	rows, e = db.Query("SELECT id,COALESCE(kind,''),COALESCE(name,''),COALESCE(purpose,''),COALESCE(state,''),COALESCE(path,''),COALESCE(sha256,'') FROM contracts ORDER BY id")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var id, kind, name, purpose, state, contractPath, hash string
		if e = rows.Scan(&id, &kind, &name, &purpose, &state, &contractPath, &hash); e != nil {
			return nil, e
		}
		entity := LibraryEntity{ID: "legacy/contract/" + id, Namespace: "legacy", Kind: kind, Key: id, Name: name, Purpose: purpose, Lifecycle: state, SourceFile: contractPath, SourceSHA256: hash, Definition: indexJSON(map[string]string{"contract_path": contractPath, "contract_sha256": hash, "qualification_state": state}), SupportedAdaptations: []string{"Use the legacy lib inspect/instantiate route and its declared qualification rules; this projection does not broaden them."}}
		if contractPath != "" {
			file, e := indexRelative(root, contractPath)
			if e != nil {
				return nil, e
			}
			if raw, e := os.ReadFile(file); e == nil {
				if indexDigest(raw) != hash || !json.Valid(raw) {
					return nil, fmt.Errorf("index.legacy_contract_drift: %s", id)
				}
				entity.Definition = indexJSON(map[string]any{"contract_path": contractPath, "contract_sha256": hash, "qualification_state": state, "contract": json.RawMessage(raw)})
			} else if !os.IsNotExist(e) {
				return nil, e
			}
		}
		entity.Discovery.Capability = LibraryCapability{ContentAdapter: "legacy_contract", BuildForQuery: "not_executed", SpecimenReview: state, ContentEnvelope: "inspect_original_contract"}
		entities = append(entities, entity)
	}
	return entities, rows.Err()
}

func collectGalleryArtifacts(entities []LibraryEntity, revision, root string) error {
	raw, e := os.ReadFile(filepath.Join(root, "design-system", "index.json"))
	if e != nil {
		return e
	}
	var gallery struct {
		SourceRevision string                       `json:"source_revision"`
		Designs        []map[string]json.RawMessage `json:"designs"`
	}
	if e = json.Unmarshal(raw, &gallery); e != nil {
		return e
	}
	if gallery.SourceRevision != revision {
		return fmt.Errorf("index.gallery_source_revision_mismatch")
	}
	byKey := map[string]int{}
	for i, entity := range entities {
		if entity.Kind == "template" {
			byKey[entity.Key] = i
		}
	}
	for _, record := range gallery.Designs {
		var key string
		json.Unmarshal(record["template"], &key)
		i, ok := byKey[key]
		if !ok {
			continue
		}
		var contractPath string
		json.Unmarshal(record["contract"], &contractPath)
		contractFile, e := indexRelative(root, contractPath)
		if e != nil {
			return e
		}
		contractRaw, e := os.ReadFile(contractFile)
		if e != nil {
			return e
		}
		var contract LibraryTemplate
		if e = json.Unmarshal(contractRaw, &contract); e != nil {
			return e
		}
		if contract.Key != key || contract.Revision != entities[i].Revision || contract.SourceSHA256 != entities[i].SourceSHA256 || contract.SourceRevision != revision {
			return fmt.Errorf("index.gallery_contract_drift: %s", key)
		}
		for _, role := range []string{"source_preview", "alternate_preview", "source_values", "alternate_values", "source_foundation", "alternate_foundation"} {
			var rel string
			json.Unmarshal(record[role], &rel)
			if rel == "" {
				continue
			}
			file, e := indexRelative(root, rel)
			if e != nil {
				return e
			}
			hash, e := indexFileDigest(file)
			if e != nil {
				return e
			}
			var declared string
			json.Unmarshal(record[role+"_sha256"], &declared)
			if declared != "" && declared != hash {
				return fmt.Errorf("index.gallery_artifact_drift: %s", rel)
			}
			entities[i].Artifacts = append(entities[i].Artifacts, LibraryArtifactLink{Role: role, Path: rel, SHA256: hash})
		}
		var native string
		json.Unmarshal(record["native_review"], &native)
		if native != "" {
			entities[i].Discovery.Capability.SpecimenReview = native
		}
	}
	return nil
}

func indexRelative(root, rel string) (string, error) {
	clean := filepath.Clean(rel)
	if rel == "" || filepath.IsAbs(rel) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("index.invalid_relative_resource: %s", rel)
	}
	base, e := filepath.Abs(root)
	if e != nil {
		return "", e
	}
	resolved, e := filepath.EvalSymlinks(base)
	if e != nil {
		return "", e
	}
	path := resolved
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		path = filepath.Join(path, part)
		st, e := os.Lstat(path)
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
		if e == nil && st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("index.symlink_resource: %s", rel)
		}
	}
	return path, nil
}
