package library

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Store struct {
	Root         string
	ContractsDir string
	IndexPath    string
	CatalogPath  string
}

func NewStore(root string) (Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Store{}, err
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return Store{}, err
	}
	return Store{Root: abs, ContractsDir: filepath.Join(abs, "library/contracts"), IndexPath: filepath.Join(abs, "library/catalog-library.sqlite"), CatalogPath: filepath.Join(abs, "samples/showcase/catalog-v3.sqlite")}, nil
}

var hex64 = regexp.MustCompile(`^[a-f0-9]{64}$`)
var hexColor = regexp.MustCompile(`^#[A-Fa-f0-9]{6}$`)

func hashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (s Store) SafePath(rel string) (string, error) {
	if rel == "" || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path must be repository-relative: %q", rel)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes repository: %q", rel)
	}
	p := filepath.Join(s.Root, clean)
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	relcheck, err := filepath.Rel(s.Root, resolved)
	if err != nil || relcheck == ".." || strings.HasPrefix(relcheck, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("symlink escapes repository: %q", rel)
	}
	return resolved, nil
}
func (s Store) VerifyArtifact(a Artifact) error {
	if !hex64.MatchString(a.SHA256) {
		return fmt.Errorf("artifact %s missing SHA-256", a.Path)
	}
	p, err := s.SafePath(a.Path)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if hashBytes(b) != a.SHA256 {
		return fmt.Errorf("stale artifact %s SHA-256", a.Path)
	}
	return nil
}
func (s Store) ContractFiles() ([]string, error) {
	ents, err := os.ReadDir(s.ContractsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && e.Name() != "schema.json" && !strings.HasSuffix(e.Name(), "-schema.json") {
			paths = append(paths, filepath.Join(s.ContractsDir, e.Name()))
		}
	}
	sort.Strings(paths)
	return paths, nil
}
func (s Store) LoadContract(path string) (Contract, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, "", err
	}
	var c Contract
	if err := json.Unmarshal(b, &c); err != nil {
		return c, "", err
	}
	if err := s.ValidateContract(c); err != nil {
		return c, "", fmt.Errorf("%s: %w", path, err)
	}
	return c, hashBytes(b), nil
}
func (s Store) Contracts() ([]ContractRecord, error) {
	paths, err := s.ContractFiles()
	if err != nil {
		return nil, err
	}
	var out []ContractRecord
	ids := map[string]bool{}
	for _, p := range paths {
		c, h, e := s.LoadContract(p)
		if e != nil {
			return nil, e
		}
		if ids[c.ID] {
			return nil, fmt.Errorf("duplicate contract ID %s", c.ID)
		}
		ids[c.ID] = true
		rel, _ := filepath.Rel(s.Root, p)
		out = append(out, ContractRecord{Contract: c, SHA256: h, Path: rel})
	}
	return out, nil
}

type ContractRecord struct {
	Contract    Contract `json:"contract"`
	SHA256      string   `json:"sha256"`
	Path        string   `json:"path"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	CanonicalID string   `json:"canonical_id,omitempty"`
}

func (s Store) ValidateContract(c Contract) error {
	if c.Schema != ContractSchema || c.ID == "" || c.Version == "" || c.Name == "" || c.Purpose == "" {
		return fmt.Errorf("invalid contract identity/schema")
	}
	switch c.Kind {
	case "layout", "component", "recipe", "asset":
	default:
		return fmt.Errorf("invalid kind %q", c.Kind)
	}
	switch c.Qualification.State {
	case "inventory", "reviewed", "measured_fixture", "adaptation_qualified":
	default:
		return fmt.Errorf("invalid qualification state %q", c.Qualification.State)
	}
	switch c.Preference.Value {
	case "preferred", "alternate", "avoid", "unreviewed":
	default:
		return fmt.Errorf("invalid preference %q", c.Preference.Value)
	}
	if c.Source.SourceID == "" || !hex64.MatchString(c.Source.SourceSHA256) || c.Source.Slide < 1 {
		return fmt.Errorf("source identity/hash/slide required")
	}
	if c.Source.Path != "" {
		if err := s.VerifyArtifact(Artifact{Path: c.Source.Path, SHA256: c.Source.SourceSHA256}); err != nil {
			return fmt.Errorf("source: %w", err)
		}
	}
	if c.Composition.SpecPath == "" || !hex64.MatchString(c.Composition.SpecSHA256) || c.Composition.SlideID == "" {
		return fmt.Errorf("composition spec path/hash/slide ID required")
	}
	if err := s.VerifyArtifact(Artifact{Path: c.Composition.SpecPath, SHA256: c.Composition.SpecSHA256}); err != nil {
		return fmt.Errorf("composition: %w", err)
	}
	if c.FitEnvelope.FontPolicy != "no_silent_shrink" {
		return fmt.Errorf("font policy must be no_silent_shrink")
	}
	for _, v := range []string{c.Transforms.Translation, c.Transforms.Resize, c.Transforms.Rotation} {
		if v != "tested" && v != "unsupported" {
			return fmt.Errorf("transforms must be tested or unsupported")
		}
	}
	seen := map[string]bool{}
	for _, sl := range c.Composition.Slots {
		if sl.Name == "" || seen[sl.Name] || !strings.HasPrefix(sl.Pointer, "/") || (sl.ValueType != "string" && sl.ValueType != "string_array") {
			return fmt.Errorf("invalid/duplicate slot %s", sl.Name)
		}
		seen[sl.Name] = true
		if sl.MaxChars <= 0 || sl.MinItems < 0 || sl.MaxItems < 0 || (sl.MaxItems > 0 && sl.MaxItems < sl.MinItems) {
			return fmt.Errorf("invalid slot bounds %s", sl.Name)
		}
		if sl.ValueType == "string_array" && sl.MaxItems == 0 {
			return fmt.Errorf("array slot %s requires max_items", sl.Name)
		}
	}
	if _, err := s.Fingerprint(c); err != nil {
		return fmt.Errorf("composition slots: %w", err)
	}
	for _, a := range c.Assets {
		if err := s.VerifyArtifact(a); err != nil {
			return err
		}
	}
	for _, v := range c.StyleVariants {
		if v.ID == "" || v.SemanticProfile == "" {
			return fmt.Errorf("style variant requires ID and semantic profile")
		}
		for token, color := range v.Tokens {
			if token == "" || !hexColor.MatchString(color) {
				return fmt.Errorf("style variant %s token %s requires concrete #RRGGBB color", v.ID, token)
			}
		}
	}
	for _, p := range c.Previews {
		if err := s.VerifyArtifact(Artifact{Path: p.Path, SHA256: p.SHA256}); err != nil {
			return err
		}
	}
	for _, e := range c.Qualification.Evidence {
		if err := s.VerifyArtifact(e); err != nil {
			return err
		}
	}
	if c.Qualification.State == "adaptation_qualified" {
		if c.Source.Path == "" {
			return fmt.Errorf("qualified contract requires hash-verifiable source path")
		}
		required := map[string]bool{"native_fit": false, "changed_content": false, "stress_negative": false, "visual_review": false}
		for _, e := range c.Qualification.Evidence {
			if _, ok := required[e.Role]; ok {
				required[e.Role] = true
			}
		}
		for role, ok := range required {
			if !ok {
				return fmt.Errorf("qualified contract missing %s proof", role)
			}
		}
	}
	return nil
}
func (s Store) ContractByID(id string) (ContractRecord, error) {
	all, err := s.Contracts()
	if err != nil {
		return ContractRecord{}, err
	}
	for _, r := range all {
		if r.Contract.ID == id {
			return r, nil
		}
	}
	return ContractRecord{}, fmt.Errorf("contract %s not found", id)
}
