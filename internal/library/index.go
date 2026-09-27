package library

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type IndexReport struct {
	Path           string `json:"path"`
	CatalogSHA256  string `json:"catalog_sha256"`
	ContractCount  int    `json:"contract_count"`
	InventoryCount int    `json:"inventory_count"`
	AliasCount     int    `json:"alias_count"`
}

func sql(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func sqlite(path, script string, readonly bool) ([]byte, error) {
	args := []string{}
	if readonly {
		args = append(args, "-readonly", "-json")
	}
	args = append(args, path)
	cmd := exec.Command("/usr/bin/sqlite3", args...)
	cmd.Stdin = strings.NewReader(script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("sqlite3: %w: %s", err, stderr.String())
	}
	if readonly && len(bytes.TrimSpace(b)) == 0 {
		return []byte("[]"), nil
	}
	return b, nil
}
func (s Store) catalogHash() (string, error) {
	b, err := os.ReadFile(filepath.Join(s.Root, "library/catalog-manifest.json"))
	if err != nil {
		return "", err
	}
	var m struct {
		Artifacts map[string]Artifact `json:"artifacts"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	a := m.Artifacts["sqlite"]
	if a.Path == "" {
		return "", fmt.Errorf("catalog manifest lacks sqlite artifact")
	}
	if err := s.VerifyArtifact(a); err != nil {
		return "", err
	}
	return a.SHA256, nil
}

type prefChoice struct {
	ComponentID string `json:"component_id"`
	Preference  string `json:"preference"`
}

func (s Store) preferences() (map[string]string, error) {
	b, err := os.ReadFile(filepath.Join(s.Root, "library/reference-preferences.json"))
	if err != nil {
		return nil, err
	}
	var p struct {
		ShortlistSHA256 string       `json:"shortlist_sha256"`
		Choices         []prefChoice `json:"choices"`
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	sb, err := os.ReadFile(filepath.Join(s.Root, "library/reference-shortlist.json"))
	if err != nil {
		return nil, err
	}
	if hashBytes(sb) != p.ShortlistSHA256 {
		return nil, fmt.Errorf("reference preference overlay has stale shortlist hash")
	}
	out := map[string]string{}
	for _, c := range p.Choices {
		out[c.ComponentID] = c.Preference
	}
	return out, nil
}
func (s Store) BuildIndex(out string) (IndexReport, error) {
	if out == "" {
		out = s.IndexPath
	}
	if _, err := os.Stat(out); err == nil {
		return IndexReport{}, fmt.Errorf("index output exists: %s", out)
	} else if !os.IsNotExist(err) {
		return IndexReport{}, err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return IndexReport{}, err
	}
	catalogHash, err := s.catalogHash()
	if err != nil {
		return IndexReport{}, err
	}
	contracts, err := s.Contracts()
	if err != nil {
		return IndexReport{}, err
	}
	prefs, err := s.preferences()
	if err != nil {
		return IndexReport{}, err
	}
	templateSQL, templateHash, err := s.templateInventorySQL()
	if err != nil {
		return IndexReport{}, err
	}
	// Canonicalize duplicate geometry, slot and behavior signatures while keeping
	// every alias addressable by exact contract ID.
	canonical := map[string]string{}
	aliasCount := 0
	for i := range contracts {
		fp, err := s.Fingerprint(contracts[i].Contract)
		if err != nil {
			return IndexReport{}, err
		}
		contracts[i].Fingerprint = fp
		if id := canonical[fp]; id != "" {
			contracts[i].CanonicalID = id
			aliasCount++
		} else {
			canonical[fp] = contracts[i].Contract.ID
			contracts[i].CanonicalID = contracts[i].Contract.ID
		}
	}
	var b strings.Builder
	b.WriteString("PRAGMA journal_mode=DELETE; BEGIN IMMEDIATE; ")
	b.WriteString("CREATE TABLE meta(key TEXT PRIMARY KEY,value TEXT NOT NULL); CREATE TABLE inventory(id TEXT PRIMARY KEY,kind TEXT,source_id TEXT,category TEXT,title TEXT,body TEXT,json TEXT,preference TEXT NOT NULL DEFAULT 'unreviewed'); CREATE VIRTUAL TABLE inventory_fts USING fts5(id UNINDEXED,title,body,tokenize='unicode61'); CREATE TABLE contracts(id TEXT PRIMARY KEY,version TEXT,kind TEXT,name TEXT,purpose TEXT,roles TEXT,state TEXT,preference TEXT,source_id TEXT,path TEXT,sha256 TEXT,fingerprint TEXT,canonical_id TEXT,preview_count INTEGER); CREATE VIRTUAL TABLE contracts_fts USING fts5(id UNINDEXED,name,purpose,roles,tokenize='unicode61'); ")
	b.WriteString("ATTACH DATABASE " + sql(s.CatalogPath) + " AS src; INSERT INTO inventory(id,kind,source_id,category,title,body,json) SELECT id,kind,source_id,category,title,body,json FROM src.items; INSERT INTO inventory_fts(id,title,body) SELECT id,title,body FROM inventory; ")
	b.WriteString("INSERT INTO meta VALUES('schema','pptxgengo.library-index.v1'); INSERT INTO meta VALUES('catalog_sha256'," + sql(catalogHash) + "); ")
	b.WriteString("INSERT INTO meta VALUES('template_catalog_sha256'," + sql(templateHash) + "); ")
	b.WriteString(templateSQL)
	prefIDs := make([]string, 0, len(prefs))
	for id := range prefs {
		prefIDs = append(prefIDs, id)
	}
	sort.Strings(prefIDs)
	for _, id := range prefIDs {
		pref := prefs[id]
		b.WriteString("UPDATE inventory SET preference=" + sql(pref) + " WHERE id=" + sql(id) + "; ")
	}
	for _, r := range contracts {
		c := r.Contract
		pref := c.Preference.Value
		if overlay := prefs[c.Source.ComponentID]; overlay != "" {
			pref = overlay
		}
		roles := strings.Join(c.ContentRoles, " ")
		b.WriteString("INSERT INTO contracts VALUES(" + strings.Join([]string{sql(c.ID), sql(c.Version), sql(c.Kind), sql(c.Name), sql(c.Purpose), sql(roles), sql(c.Qualification.State), sql(pref), sql(c.Source.SourceID), sql(r.Path), sql(r.SHA256), sql(r.Fingerprint), sql(r.CanonicalID), fmt.Sprint(len(c.Previews))}, ",") + "); ")
		b.WriteString("INSERT INTO contracts_fts VALUES(" + sql(c.ID) + "," + sql(c.Name) + "," + sql(c.Purpose) + "," + sql(roles) + "); ")
	}
	b.WriteString("COMMIT; PRAGMA user_version=1;")
	if _, err := sqlite(out, b.String(), false); err != nil {
		os.Remove(out)
		return IndexReport{}, err
	}
	countBytes, err := sqlite(out, "SELECT count(*) AS n FROM inventory;", true)
	if err != nil {
		return IndexReport{}, err
	}
	var counts []struct {
		N int `json:"n"`
	}
	if err := json.Unmarshal(countBytes, &counts); err != nil {
		return IndexReport{}, err
	}
	return IndexReport{Path: out, CatalogSHA256: catalogHash, ContractCount: len(contracts), InventoryCount: counts[0].N, AliasCount: aliasCount}, nil
}
