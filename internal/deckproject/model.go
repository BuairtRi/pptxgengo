// Package deckproject compiles human-edited deck sources into immutable,
// source-pinned WMDS build artifacts. Reviewed receipt-bound native changes
// can be adopted; it does not infer arbitrary edited PPTX as authored source.
package deckproject

import (
	"encoding/json"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"github.com/buairtri/pptxgengo/pptx"
)

const Schema = "pptxgengo.deck-document.v1"
const RuntimeVersion = "deckproject.v1"

type Reference struct {
	Scope    string `json:"scope"`
	ID       string `json:"id"`
	Revision string `json:"revision,omitempty"`
}
type Asset struct {
	RegistryID        string      `json:"registry_id,omitempty"`
	Path              string      `json:"path,omitempty"`
	SHA256            string      `json:"sha256,omitempty"`
	Description       string      `json:"description,omitempty"`
	DerivedFrom       string      `json:"derived_from,omitempty"`
	DerivationReceipt string      `json:"derivation_receipt,omitempty"`
	Focus             *AssetFocus `json:"focus,omitempty"`
}

// Focus is normalized source-image metadata for composition decisions. Rendering
// still follows the slide's explicit fit/crop contract.
type AssetFocus struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type Slide struct {
	DiagramContainment     map[string]wmdesign.DiagramContainment `json:"diagram_containment,omitempty"`
	NativeGeometryTemplate *Reference                             `json:"native_geometry_template,omitempty"`
	NativeGeometry         map[string]NativeGeometry              `json:"native_geometry,omitempty"`
	NativeOrder            map[string][]string                    `json:"native_order,omitempty"`
	ID                     string                                 `json:"id"`
	Hidden                 bool                                   `json:"hidden,omitempty"`
	Notes                  string                                 `json:"notes,omitempty"`
	DraftReview            *wmdesign.DraftReviewNote              `json:"draft_review,omitempty"`
	ContentKind            string                                 `json:"content_kind"`
	Template               Reference                              `json:"template"`
	Values                 map[string]any                         `json:"values"`
	Density                string                                 `json:"density,omitempty"`
	HeaderDensity          string                                 `json:"header_density,omitempty"`
	AutoDensity            *bool                                  `json:"auto_density,omitempty"`
	Brief                  string                                 `json:"brief,omitempty"`
	EvidenceRefs           []string                               `json:"evidence_refs,omitempty"`
}
type Zone struct {
	Role           string         `json:"role"`
	Required       bool           `json:"required"`
	Schema         map[string]any `json:"schema"`
	Description    string         `json:"description,omitempty"`
	CapacityNote   string         `json:"capacity_note,omitempty"`
	AuthoringAlias string         `json:"authoring_alias,omitempty"`
}
type Provenance struct {
	Operation          string    `json:"operation"`
	Parent             Reference `json:"parent"`
	DefinitionSnapshot string    `json:"definition_snapshot,omitempty"`
	SourceFile         string    `json:"source_file,omitempty"`
	SourceFileSHA256   string    `json:"source_file_sha256,omitempty"`
	DefinitionSHA256   string    `json:"definition_sha256"`
	Reason             string    `json:"reason,omitempty"`
}
type FrameChrome struct {
	Emphasis         string                       `json:"emphasis,omitempty"`
	Whiteboard       []wmdesign.LibraryWhiteboard `json:"whiteboard,omitempty"`
	CustomWhiteboard bool                         `json:"custom_whiteboard,omitempty"`
}
type LocalTemplate struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Provenance   *Provenance            `json:"provenance,omitempty"`
	Frame        Reference              `json:"frame"`
	FrameOptions *wmdesign.FrameRequest `json:"frame_options,omitempty"`
	FrameChrome  *FrameChrome           `json:"frame_chrome,omitempty"`
	Grid         Reference              `json:"grid"`
	Zones        map[string]Zone        `json:"zones"`
	Nodes        []Node                 `json:"nodes"`
}
type Placement struct {
	Zone string         `json:"zone"`
	Rect *wmdesign.Rect `json:"rect,omitempty"`
	Span *Span          `json:"span,omitempty"`
}
type Span struct {
	Start int     `json:"start"`
	Count int     `json:"count"`
	Y     float64 `json:"y_pt"`
	H     float64 `json:"height_pt"`
}
type Node struct {
	ID         string              `json:"id"`
	Kind       string              `json:"kind"`
	Placement  *Placement          `json:"placement,omitempty"`
	Style      string              `json:"style,omitempty"`
	Ink        string              `json:"ink,omitempty"`
	Align      string              `json:"align,omitempty"`
	Text       any                 `json:"text,omitempty"`
	Surface    string              `json:"surface,omitempty"`
	Border     string              `json:"border,omitempty"`
	Asset      any                 `json:"asset,omitempty"`
	Fit        string              `json:"fit,omitempty"`
	Rotation   float64             `json:"rotation_deg,omitempty"`
	Weight     float64             `json:"weight_pt,omitempty"`
	Nodes      []Node              `json:"nodes,omitempty"`
	Definition *Reference          `json:"definition,omitempty"`
	Arguments  map[string]any      `json:"arguments,omitempty"`
	Keys       map[string][]string `json:"keys,omitempty"`
}
type Document struct {
	EditingProfile string `json:"editing_profile,omitempty"`
	Schema         string `json:"schema"`
	ID             string `json:"id"`
	Title          string `json:"title"`
	Year           int    `json:"year"`
	Toolchain      struct {
		Lockfile string `json:"lockfile"`
	} `json:"toolchain"`
	Context           map[string]string              `json:"context,omitempty"`
	Assets            map[string]Asset               `json:"assets,omitempty"`
	LocalTemplates    map[string]LocalTemplate       `json:"local_templates,omitempty"`
	Slides            []Slide                        `json:"slides"`
	Sections          []wmdesign.SectionSpec         `json:"sections,omitempty"`
	MediaOptimization *pptx.MediaOptimizationOptions `json:"media_optimization,omitempty"`
}
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
type Project struct {
	Root, SourcePath string
	Raw              []byte
	Canonical        []byte
	Document         Document
	Positions        map[string]Position
	SourceFiles      map[string][]byte
	SlideFiles       map[string]string
	TemplateFiles    map[string]string
	NotesFiles       map[string]string
	positionFiles    map[string]string
	activeSource     string
	sourceOverrides  map[string][]byte
	tree             map[string]any
}

func canonical(v any) []byte { b, _ := json.Marshal(v); return b }
