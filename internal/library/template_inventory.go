package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Template inventory is discovery metadata. It never enters the contracts table
// and cannot bypass Instantiate's native qualification requirements.
type templateCatalog struct {
	Schema   string            `json:"schema"`
	Inputs   []Artifact        `json:"inputs"`
	Families []json.RawMessage `json:"families"`
}

type templateSource struct {
	SourceID     string `json:"source_id"`
	SourceSHA256 string `json:"source_sha256"`
	SlideCount   int    `json:"slide_count"`
}

func (s Store) loadTemplateInventory() ([]json.RawMessage, string, error) {
	path := filepath.Join(s.Root, "library/templates/catalog.json")
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	var cat templateCatalog
	if err = json.Unmarshal(b, &cat); err != nil {
		return nil, "", err
	}
	if cat.Schema != "pptxgengo.template-catalog.v1" {
		return nil, "", fmt.Errorf("unsupported template catalog schema")
	}
	if len(cat.Families) == 0 || len(cat.Inputs) == 0 {
		return nil, "", fmt.Errorf("template catalog is missing families or source inputs")
	}
	inputPaths := map[string]bool{}
	for _, a := range cat.Inputs {
		inputPaths[a.Path] = true
		if err = s.VerifyArtifact(a); err != nil {
			return nil, "", fmt.Errorf("stale template catalog input: %w", err)
		}
	}
	for _, required := range []string{"planning/source-registry.json", "library/layout-decisions.json", "library/templates/taxonomy.json"} {
		if !inputPaths[required] {
			return nil, "", fmt.Errorf("template catalog missing source input %s", required)
		}
	}
	// Sequence patterns have their own identity and are not counted as templates.
	seqPath := filepath.Join(s.Root, "library/templates/sequence-patterns.json")
	seqBytes, err := os.ReadFile(seqPath)
	if err != nil {
		return nil, "", err
	}
	var seq struct {
		Schema   string            `json:"schema"`
		Patterns []json.RawMessage `json:"patterns"`
	}
	if err = json.Unmarshal(seqBytes, &seq); err != nil {
		return nil, "", err
	}
	if seq.Schema != "pptxgengo.sequence-patterns.v1" {
		return nil, "", fmt.Errorf("unsupported sequence schema")
	}
	if len(seq.Patterns) == 0 {
		return nil, "", fmt.Errorf("sequence catalog has no patterns")
	}
	registryBytes, err := os.ReadFile(filepath.Join(s.Root, "planning/source-registry.json"))
	if err != nil {
		return nil, "", err
	}
	var registry struct {
		Sources []templateSource `json:"sources"`
	}
	if err = json.Unmarshal(registryBytes, &registry); err != nil {
		return nil, "", err
	}
	sources := make(map[string]templateSource, len(registry.Sources))
	for _, source := range registry.Sources {
		sources[source.SourceID] = source
	}
	records := append([]json.RawMessage{}, cat.Families...)
	for _, raw := range seq.Patterns {
		var m map[string]any
		if err = json.Unmarshal(raw, &m); err != nil {
			return nil, "", err
		}
		if m == nil {
			return nil, "", fmt.Errorf("sequence pattern must be a JSON object")
		}
		if supported, ok := m["supported_automatic_expansion"].(bool); !ok || supported {
			return nil, "", fmt.Errorf("sequence %v must declare automatic expansion unsupported", m["id"])
		}
		m["kind"] = "narrative_sequence"
		m["qualification_state"] = "inventory"
		m["executable"] = false
		m["adaptation_qualified"] = false
		m["design_preference"] = "unreviewed"
		var sourceIDs []string
		seenSources := map[string]bool{}
		if _, present := m["source_examples"]; present {
			examples, ok := m["source_examples"].([]any)
			if !ok {
				return nil, "", fmt.Errorf("sequence %v has invalid source examples", m["id"])
			}
			for _, example := range examples {
				if record, ok := example.(map[string]any); ok {
					id, _ := record["source_id"].(string)
					source, exists := sources[id]
					if !exists || record["source_sha256"] != source.SourceSHA256 {
						return nil, "", fmt.Errorf("sequence %v has stale/unknown source %s", m["id"], id)
					}
					for _, field := range []string{"overview", "detail"} {
						if slides, ok := record[field].([]any); ok {
							for _, value := range slides {
								n, ok := value.(float64)
								if !ok || n < 1 || n > float64(source.SlideCount) || n != float64(int(n)) {
									return nil, "", fmt.Errorf("sequence %v has invalid %s source slide", m["id"], field)
								}
							}
						}
					}
					if !seenSources[id] {
						seenSources[id] = true
						sourceIDs = append(sourceIDs, id)
					}
				} else {
					return nil, "", fmt.Errorf("sequence %v has invalid source example", m["id"])
				}
			}
		}
		m["source_ids"] = sourceIDs
		bs, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		records = append(records, bs)
	}
	return records, hashBytes(append(append([]byte{}, b...), seqBytes...)), nil
}

func (s Store) templateInventorySQL() (string, string, error) {
	records, h, err := s.loadTemplateInventory()
	if err != nil {
		return "", "", err
	}
	var b strings.Builder
	seen := map[string]bool{}
	for _, raw := range records {
		var r struct {
			ID              string   `json:"id"`
			Kind            string   `json:"kind"`
			Name            string   `json:"name"`
			PrimaryCategory string   `json:"primary_category"`
			SourceIDs       []string `json:"source_ids"`
			Disposition     string   `json:"disposition"`
			State           string   `json:"qualification_state"`
			Executable      bool     `json:"executable"`
			Qualified       bool     `json:"adaptation_qualified"`
		}
		if err = json.Unmarshal(raw, &r); err != nil {
			return "", "", err
		}
		if r.ID == "" || r.Name == "" || seen[r.ID] {
			return "", "", fmt.Errorf("missing/duplicate template inventory identity %s", r.ID)
		}
		if (r.Kind != "template_family" && r.Kind != "narrative_sequence") || r.State != "inventory" || r.Executable || r.Qualified {
			return "", "", fmt.Errorf("template discovery record cannot claim qualification: %s", r.ID)
		}
		seen[r.ID] = true
		source := ""
		if len(r.SourceIDs) > 0 {
			source = r.SourceIDs[0]
		}
		pref := "unreviewed"
		if r.Disposition == "avoid" {
			pref = "avoid"
		}
		body := string(raw)
		b.WriteString("INSERT INTO inventory(id,kind,source_id,category,title,body,json,preference) VALUES(" + strings.Join([]string{sql(r.ID), sql(r.Kind), sql(source), sql(r.PrimaryCategory), sql(r.Name), sql(body), sql(body), sql(pref)}, ",") + "); ")
		b.WriteString("INSERT INTO inventory_fts(id,title,body) VALUES(" + sql(r.ID) + "," + sql(r.Name) + "," + sql(body) + "); ")
	}
	return b.String(), h, nil
}

func (s Store) templateIndexReady(path string) error {
	_, current, err := s.loadTemplateInventory()
	if err != nil {
		return err
	}
	b, err := sqlite(path, "SELECT value FROM meta WHERE key='template_catalog_sha256';", true)
	if err != nil {
		return err
	}
	var rows []struct {
		Value string `json:"value"`
	}
	if err = json.Unmarshal(b, &rows); err != nil {
		return err
	}
	indexed := ""
	if len(rows) == 1 {
		indexed = rows[0].Value
	}
	if indexed != current {
		return fmt.Errorf("library index references stale template catalog; rebuild the index")
	}
	return nil
}

func (s Store) InspectInventory(path, id string) (json.RawMessage, error) {
	if path == "" {
		path = s.IndexPath
	}
	if err := s.indexReady(path); err != nil {
		return nil, err
	}
	b, err := sqlite(path, "SELECT json FROM inventory WHERE id="+sql(id)+";", true)
	if err != nil {
		return nil, err
	}
	// SQLite returns the stored document as a JSON string; decode that string first.
	var stored []struct {
		JSON string `json:"json"`
	}
	if err = json.Unmarshal(b, &stored); err != nil {
		return nil, err
	}
	if len(stored) != 1 {
		return nil, fmt.Errorf("inventory item %s not indexed", id)
	}
	raw := json.RawMessage(stored[0].JSON)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid inventory payload")
	}
	return raw, nil
}

func (s Store) PreviewInventory(path, id string) (Artifact, error) {
	raw, err := s.InspectInventory(path, id)
	if err != nil {
		return Artifact{}, err
	}
	var item struct {
		Preview Artifact `json:"preview"`
	}
	if err = json.Unmarshal(raw, &item); err != nil {
		return Artifact{}, err
	}
	if item.Preview.Path == "" {
		return Artifact{}, fmt.Errorf("inventory item %s has no pinned preview", id)
	}
	if err = s.VerifyArtifact(item.Preview); err != nil {
		return Artifact{}, err
	}
	return item.Preview, nil
}
