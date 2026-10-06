package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode"

	"github.com/buairtri/pptxgengo/pptx"
)

// TemplateBindingsContract is an executable, typed v2 adapter. The historical
// declarative v1 binding schema and its source-pointer examples remain distinct.
const TemplateBindingsContract = "pptxgengo.wmds-template-bindings.v2"
const BoundDocumentSchema = "pptxgengo.wmds-template-document.v1"

type TemplateIdentity struct {
	ID            string `json:"id"`
	SourcePointer string `json:"source_pointer"`
	ExpectedType  string `json:"expected_type"`
}
type TemplateGuidance struct {
	Slot          string `json:"slot"`
	MaxCharacters int    `json:"max_characters"`
	Advisory      bool   `json:"advisory"`
}
type TemplateDefinition struct {
	Key            string             `json:"key"`
	SourceFile     string             `json:"source_file"`
	SourceSHA256   string             `json:"source_sha256"`
	SourceRevision string             `json:"source_revision,omitempty"`
	Identities     []TemplateIdentity `json:"identities"`
	Guidance       []TemplateGuidance `json:"guidance,omitempty"`
}
type BoundDocument struct {
	Schema            string                         `json:"schema"`
	Year              int                            `json:"year"`
	Slides            []BoundSlide                   `json:"slides"`
	MediaOptimization *pptx.MediaOptimizationOptions `json:"media_optimization,omitempty"`
}
type BoundSlide struct {
	ID            string          `json:"id"`
	Template      string          `json:"template"`
	ContentKind   string          `json:"content_kind"`
	Density       string          `json:"density,omitempty"`
	HeaderDensity string          `json:"header_density,omitempty"`
	AutoDensity   *bool           `json:"auto_density,omitempty"`
	Values        json.RawMessage `json:"values"`
}

// TextContent is one closed content union: a visible plain string, or the direct
// RichTextSpec object {paragraphs:[...]}. No presentation style/geometry fields
// can be supplied through a text slot.
type TextContent struct {
	Text string
	Rich *RichTextSpec
}

func (v *TextContent) UnmarshalJSON(raw []byte) error {
	*v = TextContent{}
	data := bytes.TrimSpace(raw)
	if len(data) == 0 {
		return fmt.Errorf("binding.empty_text_content")
	}
	switch data[0] {
	case '"':
		if err := bindingStrictDecode(data, &v.Text); err != nil {
			return err
		}
	case '{':
		var rich RichTextSpec
		if err := bindingStrictDecode(data, &rich); err != nil {
			return err
		}
		v.Rich = &rich
	default:
		return fmt.Errorf("binding.text_requires_string_or_paragraphs")
	}
	return validateBoundText(*v)
}
func (v TextContent) MarshalJSON() ([]byte, error) {
	if err := validateBoundText(v); err != nil {
		return nil, err
	}
	if v.Rich != nil {
		return json.Marshal(v.Rich)
	}
	return json.Marshal(v.Text)
}

// BoundMetric accepts an authored display value or an explicit NumberFormatSpec.
// The renderer receives DataMetricSpec; strings never become inferred numbers.
type BoundMetric struct {
	Value  string            `json:"value,omitempty"`
	Format *NumberFormatSpec `json:"format,omitempty"`
	Label  string            `json:"label"`
}

func (m *BoundMetric) UnmarshalJSON(raw []byte) error {
	// Presence, including null, matters for the exactly-one union. A pointer
	// alone would conflate a missing property with an explicitly null property.
	var wire struct {
		Value  json.RawMessage `json:"value"`
		Format json.RawMessage `json:"format"`
		Label  string          `json:"label"`
	}
	if err := bindingStrictDecode(raw, &wire); err != nil {
		return err
	}
	if (len(wire.Value) > 0) == (len(wire.Format) > 0) {
		return fmt.Errorf("binding.metric_exactly_one_value_or_format_required")
	}
	*m = BoundMetric{Label: wire.Label}
	if len(wire.Value) > 0 {
		if bytes.Equal(bytes.TrimSpace(wire.Value), []byte("null")) {
			return fmt.Errorf("binding.metric_value_requires_string")
		}
		if err := bindingStrictDecode(wire.Value, &m.Value); err != nil {
			return err
		}
	} else {
		if bytes.Equal(bytes.TrimSpace(wire.Format), []byte("null")) {
			return fmt.Errorf("binding.metric_format_requires_object")
		}
		var sp NumberFormatSpec
		if err := bindingStrictDecode(wire.Format, &sp); err != nil {
			return err
		}
		m.Format = &sp
	}
	return validateBoundMetric(*m)
}

type BoundCardContent struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Body  string `json:"body"`
}
type BoundCardsContent struct {
	Eyebrow string             `json:"eyebrow"`
	Title   string             `json:"title"`
	Cards   []BoundCardContent `json:"cards"`
}
type BoundStatsContent struct {
	Eyebrow string      `json:"eyebrow"`
	Title   string      `json:"title"`
	Metric1 BoundMetric `json:"metric1"`
	Metric2 BoundMetric `json:"metric2"`
	Metric3 BoundMetric `json:"metric3"`
	Metric4 BoundMetric `json:"metric4"`
	Support TextContent `json:"support"`
	Source  string      `json:"source"`
}
type BoundRailContent struct {
	Eyebrow     string      `json:"eyebrow"`
	Title       string      `json:"title"`
	Metric1     BoundMetric `json:"metric1"`
	Metric2     BoundMetric `json:"metric2"`
	Support     TextContent `json:"support"`
	RailEyebrow string      `json:"railEyebrow"`
	RailHeading TextContent `json:"railHeading"`
	RailBody    TextContent `json:"railBody"`
	Source      string      `json:"source"`
}

type TemplateSlotAssignment struct {
	Slot          string          `json:"slot"`
	TargetID      string          `json:"target_id"`
	Property      string          `json:"property"`
	ValueKind     string          `json:"value_kind"`
	SourcePointer string          `json:"source_pointer"`
	Value         json.RawMessage `json:"value"`
}
type TemplateCardKey struct {
	Key           string `json:"key"`
	Ordinal       int    `json:"ordinal"`
	TargetID      string `json:"target_id"`
	SourcePointer string `json:"source_pointer"`
}
type TemplateSlideRecord struct {
	SlideID         string                   `json:"slide_id"`
	Template        string                   `json:"template"`
	Contract        string                   `json:"contract"`
	SourceFile      string                   `json:"source_file"`
	SourceSHA256    string                   `json:"source_sha256"`
	SourceRevision  string                   `json:"source_revision,omitempty"`
	ContentKind     string                   `json:"content_kind"`
	Identities      []TemplateIdentity       `json:"identities"`
	Assignments     []TemplateSlotAssignment `json:"assignments"`
	CardKeys        []TemplateCardKey        `json:"card_keys,omitempty"`
	Guidance        []TemplateGuidance       `json:"guidance,omitempty"`
	NativeQualified bool                     `json:"native_qualified"`
}
type BindingReport struct {
	Schema          string                `json:"schema"`
	Contract        string                `json:"contract"`
	Profile         string                `json:"profile"`
	Engine          string                `json:"engine"`
	SourceFiles     []SourceFile          `json:"source_files"`
	Slides          []TemplateSlideRecord `json:"slides"`
	Policy          map[string]string     `json:"policy"`
	NativeQualified bool                  `json:"native_qualified"`
}

var executableTemplateKeys = []string{"cards/3", "cards/4", "stats/four-metrics", "takeaway-rail/metrics-rail"}

// DecodeBoundDocument is the strict CLI/API boundary, including duplicate
// properties at every nesting level, unknown fields and trailing JSON values.
func DecodeBoundDocument(raw []byte) (BoundDocument, error) {
	type plain BoundDocument
	var wire plain
	if err := bindingStrictDecode(raw, &wire); err != nil {
		return BoundDocument{}, err
	}
	d := BoundDocument(wire)
	if err := validateBoundDocument(d); err != nil {
		return BoundDocument{}, err
	}
	return d, nil
}
func (d *BoundDocument) UnmarshalJSON(raw []byte) error {
	parsed, err := DecodeBoundDocument(raw)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// bindingStrictDecode scans tokens before typed decoding because encoding/json's
// ordinary duplicate-key behavior silently lets the final property win.
func bindingStrictDecode(raw []byte, out any) error {
	scan := json.NewDecoder(bytes.NewReader(raw))
	scan.UseNumber()
	if err := bindingJSONValue(scan, 0); err != nil {
		return fmt.Errorf("binding.invalid_json: %w", err)
	}
	if _, err := scan.Token(); err != io.EOF {
		return fmt.Errorf("binding.trailing_json_value")
	}
	// encoding/json accepts case-insensitive aliases of tagged property names.
	// A closed binding API accepts only its declared spelling, so Title/title
	// cannot silently assign the same field twice under different JSON names.
	if err := bindingExactFields(raw, reflect.TypeOf(out)); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return fmt.Errorf("binding.unsupported_field_or_invalid_value: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("binding.trailing_json_value")
	}
	return nil
}

func bindingExactFields(raw []byte, typ reflect.Type) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(json.RawMessage{}) {
		return nil // A closed typed adapter decodes the raw content separately.
	}
	data := bytes.TrimSpace(raw)
	if bytes.Equal(data, []byte("null")) {
		return nil // Required-content and explicit union validation rejects null.
	}
	if typ == reflect.TypeOf(sequenceBand{}) && len(data) > 0 && data[0] == '[' {
		var pair []string
		if err := json.Unmarshal(data, &pair); err != nil || len(pair) != 2 {
			return fmt.Errorf("scene.pyramid_band_shape")
		}
		return nil
	}
	if typ == reflect.TypeOf(TextContent{}) {
		if len(data) > 0 && data[0] == '{' {
			return bindingExactFields(data, reflect.TypeOf(RichTextSpec{}))
		}
		return nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return fmt.Errorf("binding.invalid_object: %w", err)
		}
		allowed := map[string]reflect.Type{}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			allowed[name] = field.Type
		}
		for name, value := range fields {
			field, ok := allowed[name]
			if !ok {
				return fmt.Errorf("binding.unsupported_field: %s", name)
			}
			if err := bindingExactFields(value, field); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return fmt.Errorf("binding.invalid_array: %w", err)
		}
		for _, value := range values {
			if err := bindingExactFields(value, typ.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
func bindingJSONValue(d *json.Decoder, depth int) error {
	if depth > 128 {
		return fmt.Errorf("binding.json_depth_exceeded")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := t.(string)
			if !ok || seen[key] {
				return fmt.Errorf("binding.duplicate_json_property: %q", key)
			}
			seen[key] = true
			if err := bindingJSONValue(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("binding.invalid_json_object")
		}
	case '[':
		for d.More() {
			if err := bindingJSONValue(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("binding.invalid_json_array")
		}
	default:
		return fmt.Errorf("binding.unexpected_json_delimiter")
	}
	return nil
}

func validateBoundDocument(d BoundDocument) error {
	if d.Schema != BoundDocumentSchema || d.Year < 2000 || d.Year > 9999 || len(d.Slides) == 0 {
		return fmt.Errorf("binding.invalid_schema_year_or_slides")
	}
	seen := map[string]bool{}
	for _, s := range d.Slides {
		if !validPartKey(s.ID) || seen[s.ID] {
			return fmt.Errorf("binding.invalid_or_duplicate_slide_id: %s", s.ID)
		}
		seen[s.ID] = true
		if s.ContentKind != "synthetic_example" && s.ContentKind != "supplied_content" {
			return fmt.Errorf("binding.invalid_content_kind: %s", s.ContentKind)
		}
		if s.Density != "" && !validBoundTypographyDensity(s.Density) {
			return fmt.Errorf("binding.invalid_density: %s", s.Density)
		}
		if s.HeaderDensity != "" && !validBoundTypographyDensity(s.HeaderDensity) {
			return fmt.Errorf("binding.invalid_header_density: %s", s.HeaderDensity)
		}
		if len(s.Values) == 0 || bytes.Equal(bytes.TrimSpace(s.Values), []byte("null")) {
			return fmt.Errorf("binding.missing_values: %s", s.ID)
		}
	}
	return nil
}

func validBoundTypographyDensity(value string) bool {
	switch value {
	case "comfortable", "compact", "dense":
		return true
	default:
		return false
	}
}

func boundPlain(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("binding.missing_visible_content: %s", field)
	}
	if strings.Contains(value, "[[") || strings.Contains(value, "]]") || strings.Contains(value, "[^") {
		return fmt.Errorf("binding.unsupported_markup: %s", field)
	}
	return nil
}
func validateBoundMetric(m BoundMetric) error {
	if (m.Value != "") == (m.Format != nil) {
		return fmt.Errorf("binding.metric_exactly_one_value_or_format_required")
	}
	if err := boundPlain("metric label", m.Label); err != nil {
		return err
	}
	if m.Format != nil {
		_, err := FormatNumber(*m.Format)
		return err
	}
	return boundPlain("metric value", m.Value)
}
func validateBoundText(v TextContent) error {
	if (v.Text != "") == (v.Rich != nil) {
		return fmt.Errorf("binding.text_exactly_one_string_or_rich_required")
	}
	if v.Rich == nil {
		return boundPlain("text", v.Text)
	}
	if len(v.Rich.Paragraphs) == 0 {
		return fmt.Errorf("binding.rich_requires_paragraphs")
	}
	seen, visible := map[string]bool{}, false
	for _, p := range v.Rich.Paragraphs {
		if !validPartKey(p.Key) || seen[p.Key] {
			return fmt.Errorf("rich.invalid_or_duplicate_paragraph_key: %s", p.Key)
		}
		seen[p.Key] = true
		runs, text := map[string]bool{}, ""
		for i, r := range p.Runs {
			if !validPartKey(r.Key) || runs[r.Key] || r.Text == "" {
				return fmt.Errorf("rich.invalid_or_duplicate_run_key_or_empty_text: %s/%s", p.Key, r.Key)
			}
			runs[r.Key] = true
			if strings.ContainsAny(r.Text, "\r\n") {
				return fmt.Errorf("rich.run_break_or_markup: %s/%s", p.Key, r.Key)
			}
			if r.Weight != 0 && r.Weight != 400 && r.Weight != 500 && r.Weight != 600 && r.Weight != 700 {
				return fmt.Errorf("font.unsupported_weight: %d", r.Weight)
			}
			chars := []rune(r.Text)
			if i > 0 && (unicode.Is(unicode.Mn, chars[0]) || unicode.Is(unicode.Mc, chars[0]) || unicode.Is(unicode.Me, chars[0])) {
				return fmt.Errorf("rich.run_splits_combining_cluster: %s/%s", p.Key, r.Key)
			}
			for _, ch := range chars {
				if unicode.IsControl(ch) || unicode.Is(unicode.Cf, ch) || !unicode.In(ch, unicode.Latin, unicode.Common, unicode.Inherited) {
					return fmt.Errorf("text.unsupported_character: U+%04X", ch)
				}
			}
			text += r.Text
		}
		if strings.Contains(text, "[[") || strings.Contains(text, "]]") || strings.Contains(text, "[^") {
			return fmt.Errorf("binding.unsupported_markup: rich paragraph %s", p.Key)
		}
		visible = visible || strings.TrimSpace(text) != ""
	}
	if !visible {
		return fmt.Errorf("binding.missing_visible_content: rich text")
	}
	return nil
}

// TemplateCatalog verifies the entire pinned bundle/source inventory each call.
// The returned definitions are only the four implemented adapters, not the full
// source inventory's templates or an arbitrary source-pointer execution engine.
func TemplateCatalog(bundle, sourceOverride string) ([]TemplateDefinition, error) {
	s, err := Load(bundle, sourceOverride)
	if err != nil {
		return nil, err
	}
	var catalog []TemplateDefinition
	for _, key := range executableTemplateKeys {
		if !usesLegacyTemplate(s.Revision, key) {
			continue
		}
		_, definition, err := compileTemplateSource(s, key)
		if err != nil {
			return nil, err
		}
		catalog = append(catalog, definition)
	}
	return catalog, nil
}

// BindTemplates compiles pinned geometry, then assigns every required content
// slot through a closed adapter. No sample source copy survives as a fallback.
func BindTemplates(bundle, sourceOverride string, input BoundDocument) (Document, BindingReport, error) {
	var doc Document
	var report BindingReport
	if err := validateBoundDocument(input); err != nil {
		return doc, report, err
	}
	s, err := Load(bundle, sourceOverride)
	if err != nil {
		return doc, report, err
	}
	doc = Document{Schema: "pptxgengo.wmds-foundation.v1", Year: input.Year, MediaOptimization: input.MediaOptimization}
	report = BindingReport{Schema: "pptxgengo.wmds-binding-report.v1", Contract: TemplateBindingsContract, Profile: CandidateProfile, Engine: CandidateEngine, SourceFiles: s.Files, Policy: map[string]string{
		"execution":     "closed typed v2 adapters; frozen geometry/styles; source pointers identify provenance and are not mutation commands",
		"content":       "all required values supplied explicitly; no source example fallback; strings never inferred as numeric data",
		"guidance":      "source character guidance is advisory; final measurement enforces line counts, font coverage, contrast and geometric capacity",
		"qualification": "binding and generation do not establish native typography or exact installed font-file identity",
	}}
	var library []LibraryTemplate
	for _, bound := range input.Slides {
		if !usesLegacyTemplate(s.Revision, bound.Template) {
			if library == nil {
				library, err = libraryCatalog(s)
				if err != nil {
					return Document{}, BindingReport{}, err
				}
			}
			def, e := libraryTemplate(library, bound.Template)
			if e != nil {
				return Document{}, BindingReport{}, e
			}
			slide, record, e := bindLibraryTemplate(def, bound)
			if e != nil {
				return Document{}, BindingReport{}, fmt.Errorf("binding.slide %s: %w", bound.ID, e)
			}
			applyBoundDensity(&slide, bound)
			doc.Slides = append(doc.Slides, slide)
			report.Slides = append(report.Slides, record)
			continue
		}
		slide, def, err := compileTemplateSource(s, bound.Template)
		if err != nil {
			return Document{}, BindingReport{}, err
		}
		slide.ID, slide.ContentKind = bound.ID, bound.ContentKind
		record := TemplateSlideRecord{SlideID: bound.ID, Template: bound.Template, Contract: TemplateBindingsContract, SourceFile: def.SourceFile, SourceSHA256: def.SourceSHA256, SourceRevision: s.Revision, ContentKind: bound.ContentKind, Identities: def.Identities, Guidance: def.Guidance}
		if err := assignBoundTemplate(&slide, def, bound.Values, &record); err != nil {
			return Document{}, BindingReport{}, fmt.Errorf("binding.slide %s: %w", bound.ID, err)
		}
		applyBoundDensity(&slide, bound)
		slide.TemplateBinding = &record
		doc.Slides = append(doc.Slides, slide)
		report.Slides = append(report.Slides, record)
	}
	return doc, report, nil
}

func applyBoundDensity(slide *SlideSpec, bound BoundSlide) {
	if bound.Density != "" {
		slide.Density = bound.Density
	}
	if bound.HeaderDensity != "" {
		slide.Frame.HeaderDensity = bound.HeaderDensity
	}
	if bound.AutoDensity != nil {
		auto := *bound.AutoDensity
		slide.AutoDensity = &auto
	}
}

func templateIdentity(def TemplateDefinition, id, expected string) (TemplateIdentity, error) {
	var result TemplateIdentity
	found := false
	for _, item := range def.Identities {
		if item.ID == id {
			if found || item.ExpectedType != expected || item.SourcePointer == "" {
				return result, fmt.Errorf("binding.compiler_identity_conflict: %s", id)
			}
			result, found = item, true
		}
	}
	if !found {
		return result, fmt.Errorf("binding.compiler_identity_missing: %s", id)
	}
	return result, nil
}
func templateNode(slide *SlideSpec, id, kind string) (*Node, error) {
	var result *Node
	for i := range slide.Nodes {
		if slide.Nodes[i].ID == id {
			if result != nil || slide.Nodes[i].Kind != kind {
				return nil, fmt.Errorf("binding.compiler_node_conflict: %s", id)
			}
			result = &slide.Nodes[i]
		}
	}
	if result == nil {
		return nil, fmt.Errorf("binding.compiler_node_missing: %s", id)
	}
	return result, nil
}
func recordTemplateAssignment(record *TemplateSlideRecord, def TemplateDefinition, slot, id, expected, property, kind string, value any) error {
	identity, err := templateIdentity(def, id, expected)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	record.Assignments = append(record.Assignments, TemplateSlotAssignment{Slot: slot, TargetID: id, Property: property, ValueKind: kind, SourcePointer: identity.SourcePointer, Value: raw})
	return nil
}
func assignTemplateChrome(slide *SlideSpec, def TemplateDefinition, record *TemplateSlideRecord, eyebrow, title string) error {
	if err := boundPlain("eyebrow", eyebrow); err != nil {
		return err
	}
	if err := boundPlain("title", title); err != nil {
		return err
	}
	if err := recordTemplateAssignment(record, def, "eyebrow", "eyebrow", "string", "slide.eyebrow", "string", eyebrow); err != nil {
		return err
	}
	if err := recordTemplateAssignment(record, def, "title", "title", "string", "slide.title", "string", title); err != nil {
		return err
	}
	slide.Eyebrow, slide.Title = eyebrow, title
	return nil
}
func assignTemplateSource(slide *SlideSpec, def TemplateDefinition, record *TemplateSlideRecord, source string) error {
	if err := boundPlain("source", source); err != nil {
		return err
	}
	if err := recordTemplateAssignment(record, def, "source", "source", "string", "slide.source", "string", source); err != nil {
		return err
	}
	slide.Source = source
	return nil
}
func assignTemplateMetric(slide *SlideSpec, def TemplateDefinition, record *TemplateSlideRecord, id string, metric BoundMetric) error {
	if err := validateBoundMetric(metric); err != nil {
		return err
	}
	node, err := templateNode(slide, id, "metric")
	if err != nil {
		return err
	}
	if node.DataMetric == nil {
		return fmt.Errorf("binding.compiler_missing_metric_payload: %s", id)
	}
	if err := recordTemplateAssignment(record, def, id, id, "metric", "node.metric", "metric", metric); err != nil {
		return err
	}
	node.DataMetric = &DataMetricSpec{Value: metric.Value, Format: metric.Format, Label: metric.Label}
	return nil
}
func assignTemplateText(slide *SlideSpec, def TemplateDefinition, record *TemplateSlideRecord, id, slot string, content TextContent) error {
	if err := validateBoundText(content); err != nil {
		return err
	}
	node, err := templateNode(slide, id, "text")
	if err != nil {
		return err
	}
	kind := "string"
	if content.Rich != nil {
		kind = "rich_text"
	}
	if err := recordTemplateAssignment(record, def, slot, id, "text", "node.text", kind, content); err != nil {
		return err
	}
	if content.Rich != nil {
		node.Kind, node.Text, node.RichText = "richtext", "", content.Rich
	} else {
		node.Text = content.Text
	}
	return nil
}

func assignBoundTemplate(slide *SlideSpec, def TemplateDefinition, raw json.RawMessage, record *TemplateSlideRecord) error {
	switch def.Key {
	case "cards/3", "cards/4":
		var values BoundCardsContent
		if err := bindingStrictDecode(raw, &values); err != nil {
			return err
		}
		if err := assignTemplateChrome(slide, def, record, values.Eyebrow, values.Title); err != nil {
			return err
		}
		expected := 3
		if def.Key == "cards/4" {
			expected = 4
		}
		if len(values.Cards) != expected {
			return fmt.Errorf("binding.card_count: %s requires %d", def.Key, expected)
		}
		node, err := templateNode(slide, "cards", "cardrow")
		if err != nil {
			return err
		}
		if node.CardRow == nil || len(node.CardRow.Items) != expected {
			return fmt.Errorf("binding.compiler_card_count_conflict")
		}
		identity, err := templateIdentity(def, "cards", "cardrow")
		if err != nil {
			return err
		}
		seen := map[string]bool{}
		for i, card := range values.Cards {
			if !validPartKey(card.Key) || seen[card.Key] {
				return fmt.Errorf("binding.invalid_or_duplicate_card_key: %s", card.Key)
			}
			seen[card.Key] = true
			if err := boundPlain("card title", card.Title); err != nil {
				return err
			}
			if err := boundPlain("card body", card.Body); err != nil {
				return err
			}
			// Keep the compiler's surface, padding, title ink and all style
			// options. Only the caller's key/title/paragraph copy are assigned.
			item := &node.CardRow.Items[i]
			item.Key, item.Card.Title = card.Key, card.Title
			item.Card.Body = []BodyBlock{{Key: "copy", Paragraph: card.Body}}
			targetID := node.ID + ".items." + card.Key
			record.CardKeys = append(record.CardKeys, TemplateCardKey{Key: card.Key, Ordinal: i + 1, TargetID: targetID, SourcePointer: identity.SourcePointer})
			if err := recordTemplateAssignment(record, def, "cards."+card.Key+".title", "cards", "cardrow", "items."+card.Key+".card.title", "string", card.Title); err != nil {
				return err
			}
			if err := recordTemplateAssignment(record, def, "cards."+card.Key+".body", "cards", "cardrow", "items."+card.Key+".card.body.copy", "string", card.Body); err != nil {
				return err
			}
		}
	case "stats/four-metrics":
		var values BoundStatsContent
		if err := bindingStrictDecode(raw, &values); err != nil {
			return err
		}
		if err := assignTemplateChrome(slide, def, record, values.Eyebrow, values.Title); err != nil {
			return err
		}
		for i, metric := range []BoundMetric{values.Metric1, values.Metric2, values.Metric3, values.Metric4} {
			if err := assignTemplateMetric(slide, def, record, fmt.Sprintf("metric%d", i+1), metric); err != nil {
				return err
			}
		}
		if err := assignTemplateText(slide, def, record, "support", "support", values.Support); err != nil {
			return err
		}
		return assignTemplateSource(slide, def, record, values.Source)
	case "takeaway-rail/metrics-rail":
		var values BoundRailContent
		if err := bindingStrictDecode(raw, &values); err != nil {
			return err
		}
		if err := assignTemplateChrome(slide, def, record, values.Eyebrow, values.Title); err != nil {
			return err
		}
		for i, metric := range []BoundMetric{values.Metric1, values.Metric2} {
			if err := assignTemplateMetric(slide, def, record, fmt.Sprintf("metric%d", i+1), metric); err != nil {
				return err
			}
		}
		if err := assignTemplateText(slide, def, record, "support", "support", values.Support); err != nil {
			return err
		}
		if err := boundPlain("railEyebrow", values.RailEyebrow); err != nil {
			return err
		}
		if err := assignTemplateText(slide, def, record, "rail-eyebrow", "railEyebrow", TextContent{Text: values.RailEyebrow}); err != nil {
			return err
		}
		if err := assignTemplateText(slide, def, record, "rail-heading", "railHeading", values.RailHeading); err != nil {
			return err
		}
		if err := assignTemplateText(slide, def, record, "rail-body", "railBody", values.RailBody); err != nil {
			return err
		}
		return assignTemplateSource(slide, def, record, values.Source)
	default:
		return fmt.Errorf("binding.unsupported_template: %s", def.Key)
	}
	return nil
}
