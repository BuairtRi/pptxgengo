package wmdesign

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AssetVariant retains the canonical registry key for each color treatment.
// ThumbnailPath points at the original registered image bytes, not a derived
// or unverified preview file.
type AssetVariant struct {
	ID                  string   `json:"id"`
	Color               string   `json:"color,omitempty"`
	RecommendedSurfaces []string `json:"recommended_surfaces,omitempty"`
	Path                string   `json:"path"`
	SHA256              string   `json:"sha256"`
	ThumbnailPath       string   `json:"thumbnail_path,omitempty"`
	ThumbnailMIME       string   `json:"thumbnail_mime,omitempty"`
	ThumbnailSHA256     string   `json:"thumbnail_sha256,omitempty"`
	ThumbnailState      string   `json:"thumbnail_state"`
}

// AssetSelection represents one searchable asset. Icon color variants are
// grouped under one stable concept ID while their registered IDs remain intact.
type AssetSelection struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tags        []string       `json:"tags"`
	People      string         `json:"people"`
	Industry    []string       `json:"industry,omitempty"`
	Setting     []string       `json:"setting,omitempty"`
	Orientation string         `json:"orientation,omitempty"`
	Variants    []AssetVariant `json:"variants"`
}

type assetMetadata struct {
	name, description, kind, people, orientation string
	tags, industry, setting                      []string
}

var curatedAssetMetadata = map[string]assetMetadata{
	"photo-abstract-blocks":             {name: "Abstract digital blocks", description: "Abstract blue digital blocks with a layered geometric pattern; no people.", kind: "photo", people: "no", orientation: "landscape", tags: []string{"abstract", "technology", "digital", "data", "blocks", "blue"}, setting: []string{"abstract"}},
	"photo-abstract-cubes":              {name: "Abstract blue cubes", description: "Abstract blue cubic architecture with repeating geometric forms; no people.", kind: "photo", people: "no", orientation: "landscape", tags: []string{"abstract", "technology", "digital", "cubes", "architecture", "blue"}, setting: []string{"abstract"}},
	"photo-abstract-grid":               {name: "Abstract illuminated data grid", description: "Abstract blue illuminated grid suggesting connected data systems; no people.", kind: "photo", people: "no", orientation: "landscape", tags: []string{"abstract", "technology", "digital", "data", "grid", "network", "blue"}, setting: []string{"abstract"}},
	"photo-abstract-led":                {name: "Abstract blue LED data pattern", description: "Close view of a blue LED display pattern; no people.", kind: "photo", people: "no", orientation: "landscape", tags: []string{"abstract", "technology", "digital", "data", "screen", "blue"}, setting: []string{"abstract"}},
	"photo-business-team-report-review": {name: "Business team reviewing performance reports", description: "Several colleagues discuss printed performance charts and reports around a meeting table.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "business team", "performance reports", "charts", "analytics", "strategy", "finance transformation", "collaboration", "decision making", "workshop"}, industry: []string{"financial services"}, setting: []string{"office meeting room"}},
	"photo-clinical-leaders":            {name: "Healthcare leaders reviewing a tablet", description: "Two healthcare professionals review a tablet together in a bright clinical setting.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "healthcare", "clinical", "leaders", "leadership", "tablet", "review", "collaboration", "collaborating", "meeting"}, industry: []string{"healthcare"}, setting: []string{"clinical", "bright interior"}},
	"photo-clinical-team":               {name: "Clinical team reviewing care data", description: "Two clinical professionals review care information on a tablet together.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "healthcare", "clinical", "care", "data", "tablet", "review", "collaboration", "collaborating", "team"}, industry: []string{"healthcare"}, setting: []string{"clinical interior"}},
	"photo-clinician-data":              {name: "Clinician reviewing patient data", description: "A clinician reviews patient data at dual monitors in a clinical workspace.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "healthcare", "clinical", "clinician", "patient", "data", "analytics", "computer", "monitor", "screen", "review"}, industry: []string{"healthcare"}, setting: []string{"clinical workspace"}},
	"photo-corridor":                    {name: "Clinical staff collaborating in a corridor", description: "Clinical staff talk together in a bright hospital corridor.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "healthcare", "clinical", "staff", "team", "collaboration", "collaborating", "conversation", "hospital", "corridor"}, industry: []string{"healthcare"}, setting: []string{"hospital corridor"}},
	"photo-executive":                   {name: "Executive team reviewing documents", description: "An executive group reviews documents around a conference table, seen through glass.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "executive", "leadership", "team", "documents", "review", "meeting", "conference", "collaboration", "collaborating", "office"}, industry: []string{"organization people change"}, setting: []string{"glass conference room", "office"}},
	"photo-financial-adviser":           {name: "Financial adviser meeting with clients", description: "A financial adviser meets with two clients across a desk in a bright office.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "financial adviser", "financial planning", "wealth management", "client consultation", "relationship banking", "personalized service", "advice", "trust", "documents"}, industry: []string{"financial services"}, setting: []string{"bright office", "client meeting"}},
	"photo-financial-analyst":           {name: "Financial analyst reviewing dashboards", description: "A financial professional reviews data visualizations across dual monitors while taking notes.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "financial analyst", "dashboard", "dual monitors", "portfolio", "banking", "investment analysis", "financial data", "risk review", "data visualization", "business intelligence"}, industry: []string{"financial services"}, setting: []string{"office workspace"}},
	"photo-headshot":                    {name: "Executive portrait", description: "Registered executive headshot portrait.", kind: "photo", people: "yes", orientation: "portrait", tags: []string{"people", "portrait", "headshot", "executive", "leadership"}, industry: []string{"organization people change"}, setting: []string{"portrait"}},
	"photo-headshot-face":               {name: "Executive portrait crop", description: "Registered face crop of the executive headshot portrait.", kind: "photo", people: "yes", orientation: "portrait", tags: []string{"people", "portrait", "headshot", "face", "executive", "leadership"}, industry: []string{"organization people change"}, setting: []string{"portrait"}},
	"photo-healthcare-leadership":       {name: "Healthcare leaders reviewing a tablet", description: "Two clinicians confer with a suited colleague reviewing information on a tablet in a bright healthcare facility.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "healthcare", "clinicians", "tablet", "clinical operations", "digital health", "healthcare transformation", "leadership", "care delivery", "technology adoption", "cross functional team"}, industry: []string{"healthcare"}, setting: []string{"healthcare facility"}},
	"photo-stethoscope":                 {name: "Stethoscope with copy space", description: "Stethoscope on a light surface with open copy space; no people.", kind: "photo", people: "no", orientation: "landscape", tags: []string{"healthcare", "medical", "medicine", "stethoscope", "care", "copy space"}, industry: []string{"healthcare"}, setting: []string{"clinical still life"}},
	"photo-software-developer-pair":     {name: "Software developers collaborating at multiple monitors", description: "Two software developers collaborate at a desk with multiple monitors, a laptop, and code notes.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "software developers", "programming", "pair programming", "code review", "software engineering", "digital product", "technology team", "agile", "developer collaboration", "digital delivery", "product development"}, setting: []string{"office technology workspace"}},
	"photo-software-engineering-team":   {name: "Software engineering team working on code", description: "Three software professionals work across laptops and code-filled desktop monitors in a modern office.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "software engineering", "developers", "code", "programming", "software delivery", "product development", "technology talent", "agile team", "digital products", "coding", "application development"}, setting: []string{"modern office technology workspace"}},
	"photo-team-meeting":                {name: "Team meeting in a glass conference room", description: "A group meets around a conference table in a glass office room.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "team", "meeting", "conference", "office", "collaboration", "collaborating", "discussion", "colleagues"}, industry: []string{"organization people change"}, setting: []string{"glass conference room", "office"}},
	"photo-technician":                  {name: "Industrial technician reviewing equipment data", description: "An industrial technician in safety equipment checks data on a laptop beside machinery.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "industrial", "technician", "manufacturing", "operations", "equipment", "data", "laptop", "safety", "review"}, industry: []string{"consumer industrial products"}, setting: []string{"industrial facility"}},
	"photo-warehouse":                   {name: "Warehouse operations manager with tablet", description: "A warehouse operations worker in safety equipment checks a tablet among inventory racks.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "warehouse", "operations", "inventory", "logistics", "tablet", "supply chain", "safety"}, industry: []string{"consumer industrial products"}, setting: []string{"warehouse"}},
	"photo-working-session":             {name: "Colleagues in a focused working session", description: "Colleagues collaborate around a table in a working session; the image shows a small group in discussion with devices on the table.", kind: "photo", people: "yes", orientation: "landscape", tags: []string{"people", "colleagues", "team", "collaboration", "collaborating", "working session", "workshop", "meeting", "discussion", "office", "devices"}, industry: []string{"organization people change"}, setting: []string{"office working session"}},
	"logo-pos":                          {name: "West Monroe horizontal positive logo", description: "West Monroe horizontal logo in positive color treatment.", kind: "logo", tags: []string{"brand", "identity", "horizontal", "positive", "color"}},
	"logo-rev":                          {name: "West Monroe horizontal reverse logo", description: "West Monroe horizontal white reverse logo for dark surfaces.", kind: "logo", tags: []string{"brand", "identity", "horizontal", "reverse", "white", "dark surface"}},
	"tagline-rev":                       {name: "West Monroe reversed tagline", description: "West Monroe reversed tagline lockup for dark surfaces.", kind: "logo", tags: []string{"brand", "identity", "tagline", "reverse", "white", "dark surface"}},
	"arrow-straight":                    {name: "Straight arrow", description: "Hand-drawn arrow graphic pointing left.", kind: "graphic", people: "no", tags: []string{"arrow", "direction", "movement", "left"}},
	"arrow-connecting":                  {name: "Connecting arrow", description: "Hand-drawn connecting arrow graphic.", kind: "graphic", people: "no", tags: []string{"arrow", "connection", "direction", "movement"}},
	"arrow-dashed":                      {name: "Dashed arrow", description: "Hand-drawn dashed arrow graphic.", kind: "graphic", people: "no", tags: []string{"arrow", "direction", "movement", "dashed"}},
	"arrow-double":                      {name: "Double arrow", description: "Hand-drawn double-headed arrow graphic.", kind: "graphic", people: "no", tags: []string{"arrow", "direction", "movement", "two way"}},
	"arrow-right-angle":                 {name: "Right-angle arrow", description: "Hand-drawn right-angle arrow graphic.", kind: "graphic", people: "no", tags: []string{"arrow", "direction", "movement", "right angle"}},
	"circle":                            {name: "Hand-drawn circle", description: "Hand-drawn circle accent graphic.", kind: "graphic", people: "no", tags: []string{"circle", "outline", "emphasis", "accent"}},
	"spark":                             {name: "Hand-drawn spark", description: "Hand-drawn spark accent graphic.", kind: "graphic", people: "no", tags: []string{"spark", "emphasis", "accent", "star"}},
	"underscore":                        {name: "Hand-drawn underscore", description: "Hand-drawn underscore accent graphic.", kind: "graphic", people: "no", tags: []string{"underscore", "underline", "emphasis", "accent"}},
	"icon/alert":                        {name: "Alert", description: "Icon representing an alert or warning.", kind: "icon", tags: []string{"warning", "risk", "issue", "attention", "notification"}},
	"icon/collaboration-high-five":      {name: "Collaboration high-five", description: "Icon representing teamwork and collaboration.", kind: "icon", tags: []string{"collaboration", "collaborating", "teamwork", "people", "partnership", "celebration"}},
	"icon/handshake":                    {name: "Handshake", description: "Icon representing an agreement or partnership.", kind: "icon", tags: []string{"partnership", "agreement", "trust", "collaboration", "collaborating", "people"}},
	"icon/hazard-warning":               {name: "Hazard warning", description: "Icon representing a hazard, warning or safety risk.", kind: "icon", tags: []string{"hazard", "warning", "safety", "risk", "alert", "uncertainty"}},
	"icon/people-group":                 {name: "People group", description: "Icon representing a group of people or workforce.", kind: "icon", tags: []string{"people", "group", "team", "workforce", "colleagues"}},
	"icon/people-network":               {name: "People network", description: "Icon representing a connected people network.", kind: "icon", tags: []string{"people", "network", "connections", "collaboration", "collaborating", "team"}},
	"icon/people-network-2":             {name: "People network", description: "Icon representing a connected people network.", kind: "icon", tags: []string{"people", "network", "connections", "collaboration", "collaborating", "team"}},
	"icon/risk-alert-arrow":             {name: "Risk alert", description: "Icon representing risk or an emerging alert.", kind: "icon", tags: []string{"risk", "warning", "alert", "issue", "trend"}},
	"icon/security-shield-gear":         {name: "Security shield with gear", description: "Icon representing security and protection of systems.", kind: "icon", tags: []string{"security", "protection", "risk", "technology", "systems"}},
	"icon/security-shield-lock":         {name: "Security shield with lock", description: "Icon representing security, privacy and protection.", kind: "icon", tags: []string{"security", "protection", "privacy", "risk", "lock"}},
	"icon/team-huddle":                  {name: "Team huddle", description: "Icon representing a team gathered for discussion.", kind: "icon", tags: []string{"team", "people", "group", "collaboration", "collaborating", "workshop", "huddle", "meeting"}},
	"icon/team-meeting":                 {name: "Team meeting", description: "Icon representing a team meeting or group discussion.", kind: "icon", tags: []string{"team", "people", "meeting", "collaboration", "collaborating", "workshop", "discussion"}},
	"icon/teamwork":                     {name: "Teamwork", description: "Icon representing people working together.", kind: "icon", tags: []string{"team", "people", "collaboration", "collaborating", "cooperation", "workshop"}},
	"icon/umbrella":                     {name: "Umbrella", description: "Icon representing protection or coverage.", kind: "icon", tags: []string{"protection", "risk", "insurance", "coverage", "weather"}},
}

// PrimitiveAssetData reads one registered original and verifies its SHA256.
// The returned path is the exact file shown to consumers as the preview source.
func PrimitiveAssetData(key string) ([]byte, PrimitiveAssetReference, string, error) {
	data, asset, err := primitiveAssetBytes(key)
	if err != nil {
		return nil, PrimitiveAssetReference{}, "", err
	}
	root := os.Getenv("WMDS_BRANDING_ROOT")
	if root == "" && os.Getenv("PPTXGENGO_RELEASE_ROOT") != "" {
		root = filepath.Join(os.Getenv("PPTXGENGO_RELEASE_ROOT"), "branding")
	}
	if root == "" {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return nil, PrimitiveAssetReference{}, "", homeErr
		}
		root = filepath.Join(home, "Documents", "branding")
	}
	path := asset.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, filepath.Clean(path))
	}
	ref := PrimitiveAssetReference{Key: key, Path: asset.Path, SHA256: asset.SHA256, Crop: asset.Crop}
	return data, ref, path, nil
}

// AssetSelections returns curated asset metadata in stable key order. Every
// result includes only registered assets whose original bytes pass hash checks.
// An empty query lists assets of the requested kind; limit <= 0 means no cap.
func AssetSelections(query, kind string, limit int) ([]AssetSelection, error) {
	kind = normalizeAssetKind(kind)
	if kind == "" {
		return nil, fmt.Errorf("asset kind must be icon, photo, graphic, logo or all")
	}
	if limit < 0 {
		return nil, fmt.Errorf("asset limit must be nonnegative")
	}
	queryTokens := queryWords(query)
	groups := map[string]*AssetSelection{}
	scores := map[string]int{}
	for _, ref := range PrimitiveAssetCatalog() {
		meta := assetMetadataFor(ref.Key, ref.Path)
		if kind != "all" && meta.kind != kind {
			continue
		}
		score := assetQueryScore(queryTokens, meta, ref.Key, ref.Path)
		if len(queryTokens) > 0 && score == 0 {
			continue
		}
		id := ref.Key
		color := ""
		if meta.kind == "logo" && strings.HasSuffix(ref.Key, "-rev") {
			color = "white"
		}
		if meta.kind == "icon" {
			parts := strings.Split(ref.Key, "/")
			if len(parts) != 3 {
				continue
			}
			id, color = "icon/"+parts[1], parts[2]
		}
		selection := groups[id]
		if selection == nil {
			selection = &AssetSelection{ID: id, Kind: meta.kind, Name: meta.name, Description: meta.description, Tags: append([]string(nil), meta.tags...), People: meta.people, Industry: append([]string(nil), meta.industry...), Setting: append([]string(nil), meta.setting...), Orientation: meta.orientation, Variants: []AssetVariant{}}
			groups[id] = selection
		}
		selection.Variants = append(selection.Variants, AssetVariant{ID: ref.Key, Color: color, RecommendedSurfaces: surfaceGuidance(color), Path: ref.Path, SHA256: ref.SHA256})
		scores[id] = max(scores[id], score)
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		if scores[ids[i]] != scores[ids[j]] {
			return scores[ids[i]] > scores[ids[j]]
		}
		return ids[i] < ids[j]
	})
	out := make([]AssetSelection, 0, len(ids))
	for _, id := range ids {
		selection := *groups[id]
		if limit > 0 && len(out) >= limit {
			break
		}
		for i := range selection.Variants {
			variant := &selection.Variants[i]
			data, ref, path, err := PrimitiveAssetData(variant.ID)
			if err != nil {
				return nil, err
			}
			if ref.Path != variant.Path || ref.SHA256 != variant.SHA256 {
				return nil, fmt.Errorf("asset registry changed during discovery: %s", variant.ID)
			}
			mimeType := mime.TypeByExtension(filepath.Ext(ref.Path))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			variant.ThumbnailPath = path
			variant.ThumbnailSHA256 = ref.SHA256
			variant.ThumbnailMIME = mimeType
			variant.ThumbnailState = "verified_registered_original"
			if selection.Kind == "photo" {
				selection.Orientation, err = originalImageDimensions(data, ref.Path)
				if err != nil {
					return nil, fmt.Errorf("asset image dimensions unavailable: %s: %w", variant.ID, err)
				}
			}
		}
		out = append(out, selection)
	}
	return out, nil
}

func surfaceGuidance(color string) []string {
	switch color {
	case "navy":
		return []string{"light surface"}
	case "white":
		return []string{"dark surface"}
	case "magenta":
		return []string{"check contrast on selected surface"}
	default:
		return nil
	}
}

func normalizeAssetKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "all", "asset", "assets":
		if strings.TrimSpace(kind) == "" {
			return "all"
		}
		return "all"
	case "icon", "icons":
		return "icon"
	case "photo", "photos", "image", "images":
		return "photo"
	case "graphic", "graphics":
		return "graphic"
	case "logo", "logos":
		return "logo"
	default:
		return ""
	}
}

func assetMetadataFor(key, path string) assetMetadata {
	base := key
	if strings.HasPrefix(key, "icon/") {
		parts := strings.Split(key, "/")
		if len(parts) == 3 {
			base = "icon/" + parts[1]
		}
	}
	if curated, ok := curatedAssetMetadata[base]; ok {
		if curated.people == "" {
			curated.people = "no"
		}
		return curated
	}
	if strings.HasPrefix(key, "highlight-") {
		return assetMetadata{name: titleWords(strings.ReplaceAll(key, "-", " ")) + " brush", description: "Registered West Monroe highlight brush graphic for emphasizing content.", kind: "graphic", people: "no", tags: []string{"highlight", "brush", "emphasis", "accent", "marker"}}
	}
	name := strings.ReplaceAll(strings.TrimPrefix(base, "icon/"), "-", " ")
	if strings.HasPrefix(key, "icon/") {
		name = titleWords(name)
		tags := append([]string{"icon"}, strings.Fields(strings.ToLower(name))...)
		return assetMetadata{name: name, description: "Icon representing " + strings.ToLower(name) + ".", kind: "icon", people: "no", tags: cleanAssetTags(tags)}
	}
	kind := "graphic"
	if strings.HasPrefix(key, "logo-") || strings.Contains(path, "/logos/") {
		kind = "logo"
	}
	if strings.HasPrefix(key, "photo-") || strings.Contains(strings.ToLower(path), "photo") {
		kind = "photo"
	}
	name = titleWords(strings.ReplaceAll(base, "-", " "))
	return assetMetadata{name: name, description: "Registered West Monroe " + kind + ": " + name + ".", kind: kind, people: "unknown", tags: cleanAssetTags(append([]string{kind}, queryWords(base)...))}
}

func titleWords(value string) string {
	fields := strings.Fields(value)
	for i, field := range fields {
		if field != "" {
			fields[i] = strings.ToUpper(field[:1]) + field[1:]
		}
	}
	return strings.Join(fields, " ")
}

func cleanAssetTags(tags []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" && !seen[tag] {
			seen[tag] = true
			out = append(out, tag)
		}
	}
	sort.Strings(out)
	return out
}

func queryWords(query string) []string {
	parts := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') })
	terms := []string{}
	for _, part := range parts {
		switch part {
		case "a", "an", "the", "in", "on", "at", "of", "for", "with", "and", "to":
			continue
		}
		terms = append(terms, part)
	}
	return cleanAssetTags(terms)
}

func assetQueryScore(terms []string, meta assetMetadata, key, path string) int {
	if len(terms) == 0 {
		return 0
	}
	corpus := strings.ToLower(strings.Join(append(append(append(append([]string{key, path, meta.name, meta.description}, meta.tags...), meta.industry...), meta.setting...), meta.kind), " "))
	words := queryWords(corpus)
	score := 0
	for _, term := range terms {
		matched := false
		for _, candidate := range assetQueryExpansions(term) {
			for _, word := range words {
				if assetWordMatches(candidate, word) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if matched {
			score += 10
			for _, word := range words {
				if word == term {
					score += 5
					break
				}
			}
		}
	}
	return score
}

var assetSynonymMap = map[string][]string{
	"people": {"person", "team", "workforce", "colleagues", "staff"}, "person": {"people", "team", "workforce"},
	"collaboration": {"collaborating", "teamwork", "cooperation", "partnership", "meeting", "workshop"}, "collaborating": {"collaboration", "teamwork", "cooperation", "partnership"},
	"risk": {"warning", "hazard", "alert", "safety", "security", "protection"}, "warning": {"risk", "hazard", "alert", "safety"},
	"security": {"protection", "privacy", "risk"}, "protection": {"security", "coverage", "safety"},
	"finance": {"money", "payment", "banking", "cost"}, "healthcare": {"clinical", "medical", "care", "patient"},
	"technology": {"digital", "data", "automation", "systems"}, "growth": {"increase", "improvement", "progress"},
	"industry": {"industrial", "manufacturing", "operations"}, "industrial": {"industry", "manufacturing", "operations"},
	"strategy": {"roadmap", "direction", "planning"}, "planning": {"strategy", "roadmap", "priorities"},
	"alert": {"risk", "warning", "issue"}, "team": {"people", "colleagues", "workforce", "collaboration"},
}

func assetQueryExpansions(term string) []string {
	values := append([]string{term}, assetSynonymMap[term]...)
	return values
}

func assetWordMatches(query, candidate string) bool {
	if query == candidate {
		return true
	}
	stem := func(word string) string {
		for _, suffix := range []string{"ing", "ers", "ies", "es", "s"} {
			if len(word) > len(suffix)+2 && strings.HasSuffix(word, suffix) {
				if suffix == "ies" {
					return word[:len(word)-3] + "y"
				}
				return strings.TrimSuffix(word, suffix)
			}
		}
		return word
	}
	if stem(query) == stem(candidate) {
		return true
	}
	return len(query) >= 4 && strings.HasPrefix(candidate, query) || len(candidate) >= 4 && strings.HasPrefix(query, candidate)
}

// originalImageDimensions is retained for future thumbnail consumers that need
// layout metadata without decoding pixels. It intentionally does not rasterize SVG.
func originalImageDimensions(data []byte, path string) (string, error) {
	if strings.EqualFold(filepath.Ext(path), ".svg") {
		return "vector", nil
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	if config.Width > config.Height {
		return "landscape", nil
	}
	if config.Height > config.Width {
		return "portrait", nil
	}
	return "square", nil
}
