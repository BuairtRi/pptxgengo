package library

import "encoding/json"

const ContractSchema = "pptxgengo.library-component.v1"
const NarrativeSchema = "pptxgengo.proposal-narrative.v1"

// Contract is the portable source of truth; SQLite only indexes a projection.
type Contract struct {
	Schema        string           `json:"schema"`
	ID            string           `json:"id"`
	Version       string           `json:"version"`
	Kind          string           `json:"kind"`
	Name          string           `json:"name"`
	Purpose       string           `json:"purpose"`
	ContentRoles  []string         `json:"content_roles"`
	Source        Source           `json:"source"`
	Composition   Composition      `json:"composition"`
	Assets        []Artifact       `json:"assets,omitempty"`
	Cardinality   map[string]Range `json:"cardinality,omitempty"`
	FitEnvelope   FitEnvelope      `json:"fit_envelope"`
	Transforms    Transforms       `json:"transforms"`
	StyleVariants []StyleVariant   `json:"style_variants,omitempty"`
	Preference    Preference       `json:"preference"`
	Qualification Qualification    `json:"qualification"`
	Previews      []Preview        `json:"previews,omitempty"`
	Provenance    Provenance       `json:"provenance"`
	Aliases       []string         `json:"aliases,omitempty"`
}
type Source struct {
	SourceID        string   `json:"source_id"`
	ComponentID     string   `json:"component_id,omitempty"`
	Path            string   `json:"path,omitempty"`
	SourceSHA256    string   `json:"source_sha256"`
	Slide           int      `json:"slide"`
	SourcePart      string   `json:"source_part,omitempty"`
	SourceObjectIDs []string `json:"source_object_ids,omitempty"`
	SceneSHA256     string   `json:"scene_sha256,omitempty"`
}
type Composition struct {
	SpecPath          string            `json:"spec_path"`
	SpecSHA256        string            `json:"spec_sha256"`
	SlideID           string            `json:"slide_id"`
	Children          []string          `json:"children,omitempty"`
	Slots             []Slot            `json:"slots"`
	NarrativeBindings map[string]string `json:"narrative_bindings,omitempty"`
	PaginationBinding string            `json:"pagination_binding,omitempty"`
}
type Slot struct {
	Name             string          `json:"name"`
	Role             string          `json:"role"`
	Pointer          string          `json:"pointer"`
	ValueType        string          `json:"value_type"`
	Required         bool            `json:"required"`
	MaxChars         int             `json:"max_chars,omitempty"`
	MinItems         int             `json:"min_items,omitempty"`
	MaxItems         int             `json:"max_items,omitempty"`
	ItemTemplate     json.RawMessage `json:"item_template,omitempty"`
	ItemValuePointer string          `json:"item_value_pointer,omitempty"`
	ItemIDPointer    string          `json:"item_id_pointer,omitempty"`
	ItemIDPrefix     string          `json:"item_id_prefix,omitempty"`
}
type Artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role,omitempty"`
}
type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}
type FitEnvelope struct {
	Measured    []string `json:"measured,omitempty"`
	Unsupported []string `json:"unsupported,omitempty"`
	FontPolicy  string   `json:"font_policy"`
}
type Transforms struct {
	Translation string `json:"translation"`
	Resize      string `json:"resize"`
	Rotation    string `json:"rotation"`
}
type StyleVariant struct {
	ID              string            `json:"id"`
	SemanticProfile string            `json:"semantic_profile"`
	Tokens          map[string]string `json:"tokens,omitempty"`
	Qualified       bool              `json:"qualified"`
}
type Preference struct {
	Value  string `json:"value"`
	Source string `json:"source,omitempty"`
	Note   string `json:"note,omitempty"`
}
type Qualification struct {
	State      string     `json:"state"`
	Evidence   []Artifact `json:"evidence,omitempty"`
	Failures   []string   `json:"failures,omitempty"`
	ReviewedAt string     `json:"reviewed_at,omitempty"`
}
type Preview struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Caption    string `json:"caption,omitempty"`
	SourceLink string `json:"source_link,omitempty"`
	Variant    string `json:"variant,omitempty"`
}
type Provenance struct {
	CreatedFrom   []string `json:"created_from,omitempty"`
	ReviewHistory []string `json:"review_history,omitempty"`
}

type Narrative struct {
	Schema   string           `json:"schema"`
	Brief    NarrativeBrief   `json:"brief"`
	Slides   []NarrativeSlide `json:"slides"`
	Claims   []Claim          `json:"claims,omitempty"`
	Evidence []Evidence       `json:"evidence,omitempty"`
}
type NarrativeBrief struct {
	Name         string     `json:"name"`
	Synthetic    bool       `json:"synthetic"`
	Date         string     `json:"date"`
	Audience     string     `json:"audience"`
	Decision     string     `json:"decision"`
	SourcePacket []Artifact `json:"source_packet,omitempty"`
}
type NarrativeSlide struct {
	ID                 string   `json:"id"`
	Audience           string   `json:"audience"`
	Role               string   `json:"role"`
	Takeaway           string   `json:"takeaway"`
	AssertionTitle     string   `json:"assertion_title"`
	RequiredDetail     []string `json:"required_detail"`
	EmphasisTargets    []string `json:"emphasis_targets,omitempty"`
	VisualRelationship string   `json:"visual_relationship"`
	EvidenceRefs       []string `json:"evidence_refs,omitempty"`
	Qualifications     []string `json:"qualifications,omitempty"`
}
type Claim struct {
	ID            string   `json:"id"`
	Assertion     string   `json:"assertion"`
	Status        string   `json:"status"`
	EvidenceRefs  []string `json:"evidence_refs,omitempty"`
	Qualification string   `json:"qualification,omitempty"`
}
type Evidence struct {
	ID           string   `json:"id"`
	Source       Artifact `json:"source"`
	Locator      string   `json:"locator"`
	Quote        string   `json:"quote,omitempty"`
	TypedExtract string   `json:"typed_extract,omitempty"`
	Caveat       string   `json:"caveat,omitempty"`
}
