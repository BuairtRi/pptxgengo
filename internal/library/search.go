package library

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type FindOptions struct {
	Query        string
	Kind         string
	State        string
	Preference   string
	Source       string
	Limit        int
	Inventory    bool
	IncludeAvoid bool
}
type FindHit struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Purpose     string `json:"purpose,omitempty"`
	SourceID    string `json:"source_id,omitempty"`
	State       string `json:"qualification_state"`
	Preference  string `json:"preference"`
	CanonicalID string `json:"canonical_id,omitempty"`
	Why         string `json:"why"`
}

func (s Store) indexReady(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	b, err := sqlite(path, "SELECT value FROM meta WHERE key='catalog_sha256';", true)
	if err != nil {
		return err
	}
	var rows []struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return err
	}
	if len(rows) != 1 {
		return fmt.Errorf("library index metadata missing")
	}
	h, err := s.catalogHash()
	if err != nil {
		return err
	}
	if rows[0].Value != h {
		return fmt.Errorf("library index references stale catalog")
	}
	return nil
}
func (s Store) Find(path string, o FindOptions) ([]FindHit, error) {
	if path == "" {
		path = s.IndexPath
	}
	if err := s.indexReady(path); err != nil {
		return nil, err
	}
	if o.Limit == 0 {
		o.Limit = 20
	}
	if o.Limit < 1 || o.Limit > 100 {
		return nil, fmt.Errorf("limit must be 1..100")
	}
	filters := []string{"1=1"}
	if o.Kind != "" {
		filters = append(filters, "kind="+sql(o.Kind))
	}
	if o.Source != "" {
		filters = append(filters, "source_id="+sql(o.Source))
	}
	if o.Preference != "" {
		filters = append(filters, "preference="+sql(o.Preference))
	} else if !o.IncludeAvoid {
		filters = append(filters, "preference!='avoid'")
	}
	query := strings.TrimSpace(o.Query)
	if o.Inventory {
		q := "SELECT id,kind,title AS name,source_id,preference FROM inventory WHERE " + strings.Join(filters, " AND ")
		if query != "" {
			q += " AND id IN (SELECT id FROM inventory_fts WHERE inventory_fts MATCH " + sql(query) + ")"
		}
		q += " ORDER BY CASE preference WHEN 'preferred' THEN 0 WHEN 'alternate' THEN 1 WHEN 'unreviewed' THEN 2 ELSE 3 END, id LIMIT " + fmt.Sprint(o.Limit) + ";"
		b, err := sqlite(path, q, true)
		if err != nil {
			return nil, err
		}
		var rows []struct {
			ID         string `json:"id"`
			Kind       string `json:"kind"`
			Name       string `json:"name"`
			SourceID   string `json:"source_id"`
			Preference string `json:"preference"`
		}
		if err := json.Unmarshal(b, &rows); err != nil {
			return nil, err
		}
		out := make([]FindHit, 0, len(rows))
		for _, r := range rows {
			out = append(out, FindHit{ID: r.ID, Kind: r.Kind, Name: r.Name, SourceID: r.SourceID, State: "inventory", Preference: r.Preference, Why: "catalog inventory match; adaptation unproven"})
		}
		return out, nil
	}
	if o.State != "" {
		filters = append(filters, "state="+sql(o.State))
	} else {
		filters = append(filters, "state='adaptation_qualified'")
	}
	q := "SELECT id,kind,name,purpose,source_id,state,preference,canonical_id FROM contracts WHERE " + strings.Join(filters, " AND ")
	if query != "" {
		q += " AND id IN (SELECT id FROM contracts_fts WHERE contracts_fts MATCH " + sql(query) + ")"
	}
	q += " ORDER BY CASE preference WHEN 'preferred' THEN 0 WHEN 'alternate' THEN 1 WHEN 'unreviewed' THEN 2 ELSE 3 END, CASE state WHEN 'adaptation_qualified' THEN 0 WHEN 'measured_fixture' THEN 1 WHEN 'reviewed' THEN 2 ELSE 3 END, id LIMIT " + fmt.Sprint(o.Limit) + ";"
	b, err := sqlite(path, q, true)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		ID          string `json:"id"`
		Kind        string `json:"kind"`
		Name        string `json:"name"`
		Purpose     string `json:"purpose"`
		SourceID    string `json:"source_id"`
		State       string `json:"state"`
		Preference  string `json:"preference"`
		CanonicalID string `json:"canonical_id"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}
	out := make([]FindHit, 0, len(rows))
	for _, r := range rows {
		out = append(out, FindHit{ID: r.ID, Kind: r.Kind, Name: r.Name, Purpose: r.Purpose, SourceID: r.SourceID, State: r.State, Preference: r.Preference, CanonicalID: r.CanonicalID, Why: "contract purpose/role match; qualification and preference filters applied"})
	}
	return out, nil
}

type Inspection struct {
	ContractRecord
	IndexPath string `json:"index_path"`
	Caveat    string `json:"caveat"`
}

func (s Store) Inspect(indexPath, id string) (Inspection, error) {
	if indexPath == "" {
		indexPath = s.IndexPath
	}
	if err := s.indexReady(indexPath); err != nil {
		return Inspection{}, err
	}
	b, err := sqlite(indexPath, "SELECT path,sha256,fingerprint,canonical_id FROM contracts WHERE id="+sql(id)+";", true)
	if err != nil {
		return Inspection{}, err
	}
	var rows []struct {
		Path        string `json:"path"`
		SHA256      string `json:"sha256"`
		Fingerprint string `json:"fingerprint"`
		CanonicalID string `json:"canonical_id"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		return Inspection{}, err
	}
	if len(rows) != 1 {
		return Inspection{}, fmt.Errorf("contract %s not indexed", id)
	}
	p, err := s.SafePath(rows[0].Path)
	if err != nil {
		return Inspection{}, err
	}
	c, h, err := s.LoadContract(p)
	if err != nil {
		return Inspection{}, err
	}
	if c.ID != id || h != rows[0].SHA256 {
		return Inspection{}, fmt.Errorf("contract/index correspondence is stale for %s", id)
	}
	return Inspection{ContractRecord: ContractRecord{Contract: c, SHA256: h, Path: rows[0].Path, Fingerprint: rows[0].Fingerprint, CanonicalID: rows[0].CanonicalID}, IndexPath: indexPath, Caveat: "Native measurement, fit report and visual review are required for new copy."}, nil
}

type PreviewResult struct {
	ContractID  string   `json:"contract_id"`
	Variant     string   `json:"variant,omitempty"`
	Preview     *Preview `json:"preview,omitempty"`
	FullSlide   *Preview `json:"full_slide,omitempty"`
	Preparation string   `json:"preparation,omitempty"`
}

func (s Store) Preview(indexPath, id, variant string) (PreviewResult, error) {
	ins, err := s.Inspect(indexPath, id)
	if err != nil {
		return PreviewResult{}, err
	}
	out := PreviewResult{ContractID: id, Variant: variant}
	for _, p := range ins.Contract.Previews {
		if p.Variant == "full_slide" {
			cp := p
			out.FullSlide = &cp
		}
		if (variant == "" && p.Variant != "full_slide" && out.Preview == nil) || (variant != "" && p.Variant == variant) {
			cp := p
			out.Preview = &cp
		}
	}
	if out.Preview == nil {
		out.Preparation = "No pinned preview. Prepare a read-only render from the source or contract spec, then record its hash before use."
	}
	return out, nil
}
