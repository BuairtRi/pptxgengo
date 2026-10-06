package wmdesign

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// TypographyDensity is copied from the immutable, checksum-verified tokens.
// A density selection never edits source styles or component input geometry.
type TypographyDensity struct {
	Levels     []string                `json:"levels"`
	Header     map[string][][2]float64 `json:"header"`
	Body       map[string][][2]float64 `json:"body"`
	Aliases    map[string]string       `json:"aliases"`
	Exceptions []DensityException      `json:"exceptions,omitempty"`
}

type DensityException struct {
	Element string     `json:"element"`
	Node    string     `json:"node"`
	Size    [2]float64 `json:"size"`
	Font    string     `json:"font"`
	Reason  string     `json:"reason"`
}

type SlideDensityRecord struct {
	Requested string        `json:"requested"`
	Resolved  string        `json:"resolved"`
	Header    string        `json:"header"`
	Limit     string        `json:"limit"`
	Auto      bool          `json:"auto"`
	Adjusted  bool          `json:"adjusted"`
	Steps     []DensityStep `json:"steps,omitempty"`
}
type SlideDensityAdjustment struct {
	SlideID   string        `json:"slide_id"`
	Requested string        `json:"requested"`
	Resolved  string        `json:"resolved"`
	Reason    string        `json:"reason"`
	Steps     []DensityStep `json:"steps,omitempty"`
}

type DensityStep struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

func densityLevel(level string) (string, int, error) {
	switch level {
	case "", "standard", "comfortable":
		return "comfortable", 0, nil
	case "compact":
		return level, 1, nil
	case "dense":
		return level, 2, nil
	default:
		return "", 0, fmt.Errorf("density.unknown_level: %s", level)
	}
}

func validateTypographyDensity(d *TypographyDensity) error {
	if d == nil {
		return nil
	}
	if len(d.Levels) != 3 || d.Levels[0] != "comfortable" || d.Levels[1] != "compact" || d.Levels[2] != "dense" || d.Aliases["standard"] != "comfortable" {
		return fmt.Errorf("density.invalid_levels")
	}
	for _, roles := range []map[string][][2]float64{d.Header, d.Body} {
		for role, pairs := range roles {
			if len(pairs) != 3 {
				return fmt.Errorf("density.invalid_role: %s", role)
			}
			scalar := strings.HasPrefix(role, "list-") || strings.HasPrefix(role, "ol-col")
			for i, pair := range pairs {
				if !intakeFinite(pair[0], pair[1]) || pair[0] <= 0 || (!scalar && (pair[0] < 8 || pair[1] < pair[0])) || (scalar && pair[1] != 0) || (i > 0 && (pair[0] > pairs[i-1][0] || pair[1] > pairs[i-1][1])) {
					return fmt.Errorf("density.invalid_role_metrics: %s", role)
				}
			}
		}
	}
	for _, role := range []string{"title", "heading", "eyebrow"} {
		if len(d.Header[role]) != 3 {
			return fmt.Errorf("density.missing_header_role: %s", role)
		}
	}
	for _, role := range []string{"heading", "subhead", "lead", "body", "small", "label", "eyebrow", "number", "stat", "stat-sm", "cell-body", "cell-small", "list-gap", "list-gap-small", "list-indent", "list-indent-small", "list-marker", "list-marker-small", "ol-col", "ol-col-small"} {
		if len(d.Body[role]) != 3 {
			return fmt.Errorf("density.missing_body_role: %s", role)
		}
	}
	seen := map[string]bool{}
	for _, exception := range d.Exceptions {
		want, font := [2]float64{}, ""
		switch exception.Node {
		case "gantt periods.sublabels":
			want, font = [2]float64{7.5, 9}, "mono"
		case "reviewnote":
			want, font = [2]float64{6.5, 9}, "mono caps"
		default:
			return fmt.Errorf("density.unknown_exception: %s", exception.Node)
		}
		if exception.Size != want || exception.Font != font || strings.TrimSpace(exception.Element) == "" || strings.TrimSpace(exception.Reason) == "" || seen[exception.Node] {
			return fmt.Errorf("density.invalid_exception: %s", exception.Node)
		}
		seen[exception.Node] = true
	}
	return nil
}

func densityRoleCorrections(s *Source) bool {
	return s != nil && s.Tokens.Density != nil && len(s.Tokens.Density.Body["number-long"]) == 3
}

// StyleForDensity resolves a role in body, header, or table-cell scope, returning
// a value copy. Literal source variants applied afterward retain their semantics.
func (s *Source) StyleForDensity(token, level, scope string) (Style, error) {
	baseToken := token
	if token == "number-long" && densityRoleCorrections(s) {
		baseToken = "number"
	}
	st, err := s.Style(baseToken)
	if err != nil {
		return st, err
	}
	_, tier, err := densityLevel(level)
	if err != nil {
		return st, err
	}
	if scope != "body" && scope != "header" && scope != "cell" {
		return st, fmt.Errorf("density.unknown_scope: %s", scope)
	}
	d := s.Tokens.Density
	if d == nil {
		if tier != 0 {
			return st, fmt.Errorf("density.tokens_required: %s", level)
		}
		return st, nil
	}
	roles := d.Body
	role := token
	if scope == "header" {
		roles = d.Header
	}
	if scope == "cell" && (token == "body" || token == "small") {
		role = "cell-" + token
	}
	pairs, ok := roles[role]
	if !ok {
		return st, nil
	} // source, legal, display and literal chrome keep their own type.
	if len(pairs) != 3 {
		return st, fmt.Errorf("density.invalid_role: %s", role)
	}
	st, err = primitiveStyleSize(st, pairs[tier][0])
	st.Leading = pairs[tier][1]
	if token == "number-long" {
		st.ID = token
	}
	return st, err
}

func (r *renderer) bodyStyle(token string) (Style, error) {
	scope := r.densityScope
	if scope == "" {
		scope = "body"
	}
	st, err := r.source.StyleForDensity(token, r.bodyDensity, scope)
	return st, err
}
func (r *renderer) headerStyle(token string) (Style, error) {
	return r.source.StyleForDensity(token, r.headerDensity, "header")
}
func (r *renderer) cellStyle(token string) (Style, error) {
	st, err := r.source.StyleForDensity(token, r.bodyDensity, "cell")
	return st, err
}

func (r *renderer) listMetric(role string, fallback float64) float64 {
	if r.source.Tokens.Density == nil {
		return fallback
	}
	_, tier, err := densityLevel(r.bodyDensity)
	pairs := r.source.Tokens.Density.Body[role]
	if err != nil || len(pairs) != 3 {
		return fallback
	}
	return pairs[tier][0]
}

func slideDensity(s *Source, slide SlideSpec) (SlideDensityRecord, error) {
	body, bodyTier, err := densityLevel(slide.Density)
	if err != nil {
		return SlideDensityRecord{}, err
	}
	header, _, err := densityLevel(slide.Frame.HeaderDensity)
	if err != nil {
		return SlideDensityRecord{}, err
	}
	auto := s.Tokens.Density != nil
	if slide.AutoDensity != nil {
		auto = *slide.AutoDensity
	}
	if s.Tokens.Density == nil && (body != "comfortable" || header != "comfortable" || auto) {
		return SlideDensityRecord{}, fmt.Errorf("density.tokens_required: upgrade the source bundle for typography density")
	}
	limit := slide.DensityLimit
	if s.Tokens.Density != nil && slide.TemplateBinding != nil && slide.TemplateBinding.Template != "" {
		// Resolve restrictions from the pinned source even if a caller removed
		// or relaxed the compiled provenance field. Project/binding overlays
		// cannot turn a protected source template into an unrestricted tier.
		var ok bool
		limit, ok = s.densityLimits[slide.TemplateBinding.Template]
		if !ok {
			return SlideDensityRecord{}, fmt.Errorf("density.source_limit_binding_unknown: %s", slide.TemplateBinding.Template)
		}
	}
	limit, maxTier, err := densityLimit(limit)
	if err != nil {
		return SlideDensityRecord{}, err
	}
	if slide.DensityLimit != "" && s.Tokens.Density == nil {
		return SlideDensityRecord{}, fmt.Errorf("density.tokens_required: source density limit requires density tokens")
	}
	if bodyTier > maxTier {
		return SlideDensityRecord{}, fmt.Errorf("density.prohibited_tier: slide %s requests %s; source densityLimit is %s; choose an allowed tier or another template", slide.ID, body, limit)
	}
	return SlideDensityRecord{Requested: body, Resolved: body, Header: header, Limit: limit, Auto: auto}, nil
}

func densityLimit(limit string) (string, int, error) {
	if limit == "" {
		return "dense", 2, nil
	}
	if limit != "comfortable" && limit != "compact" && limit != "dense" {
		return "", 0, fmt.Errorf("density.invalid_limit: %s", limit)
	}
	return densityLevel(limit)
}

func loadSourceDensityLimits(s *Source) error {
	if s.Tokens.Density == nil {
		return nil
	}
	s.densityLimits = map[string]string{}
	for path, raw := range s.Templates {
		if !strings.HasPrefix(path, "templates/library/") || strings.HasPrefix(filepath.Base(path), "_") {
			continue
		}
		var catalog templateSourceCatalog
		if err := sceneDecode(raw, &catalog); err != nil {
			return err
		}
		for _, rawEntry := range catalog.Templates {
			var entry libraryEntry
			if err := sceneDecode(rawEntry, &entry); err != nil {
				return err
			}
			var slide struct {
				Limit string `json:"densityLimit"`
			}
			if err := json.Unmarshal(entry.Slide, &slide); err != nil {
				return err
			}
			limit, _, err := densityLimit(slide.Limit)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", entry.ID, entry.Variant, err)
			}
			key := entry.ID + "/" + entry.Variant
			if _, exists := s.densityLimits[key]; exists {
				return fmt.Errorf("library.duplicate_template: %s", key)
			}
			s.densityLimits[key] = limit
		}
	}
	return nil
}

// Only body layout failures can request another complete slide typography tier.
// Header, validation, geometry, contrast and font errors are never fit requests.
type densityFitError struct{ cause error }

func (e *densityFitError) Error() string { return e.cause.Error() }
func (e *densityFitError) Unwrap() error { return e.cause }
func isDensityFitError(err error) bool   { var fit *densityFitError; return errors.As(err, &fit) }
func bodyDensityFitFailure(err error, body bool) error {
	if err != nil && body && densityFitCode(err) && !isDensityFitError(err) {
		return &densityFitError{cause: err}
	}
	return err
}
func densityFitCode(err error) bool {
	message := err.Error()
	for _, code := range []string{
		"text.horizontal_overflow:", "text.vertical_overflow:", "text.line_allocation_exceeded:",
		"component.vertical_overflow:", "component.line_limit:",
		"cardrow.vertical_overflow:",
		"component.uncertain_wrap_boundary:", "rich.uncertain_wrap_boundary:",
		"scene.text_vertical_overflow:", "scene.text_overflow:", "scene.textblock_overflow:", "scene.row_text_overflow:",
		"scene.card_vertical_overflow:", "scene.table_cell_overflow:", "scene.table_bullet_overflow:", "scene.table_plain_sub_overflow:",
		"scene.round12_stack_overflow:", "scene.chevron_stack_overflow:", "scene.node_stack_overflow:", "scene.layerrow_highlight_overflow",
		"scene.metric_vertical_overflow:", "scene.role_stack_overflow", "scene.org_role_overflow:",
	} {
		if strings.Contains(message, code) {
			return true
		}
	}
	return false
}

func densityMarkerOffset(st Style, marker float64, small bool) float64 {
	if small {
		return st.Leading/2 - 2
	}
	return st.Leading/2 - marker/2
}

// A split-zone violation can request another body tier only when its painted
// geometry still fits and text alone extends below the zone. Authored positions,
// widths, rotated text and non-text objects cannot be repaired by density.
func densityTextOnlyBottomOverflow(p *scenePlan, zone Rect) bool {
	if p == nil || zone.W <= 0 || zone.H <= 0 {
		return false
	}
	textOverflow := false
	flowOverflow := p.TextFlowBounds && p.Bounds.X >= zone.X-.02 && p.Bounds.X+p.Bounds.W <= zone.X+zone.W+.02 && p.Bounds.Y >= zone.Y-.02 && p.Bounds.Y < zone.Y+zone.H && p.Bounds.Y+p.Bounds.H > zone.Y+zone.H+.02
	flowContainer := false
	for _, item := range p.Items {
		var b Rect
		switch {
		case item.Text != nil:
			tr := item.Text
			b = diagramRotatedRect(tr.Rect, tr.Rotation)
			if inside(b, zone) {
				continue
			}
			if tr.Rotation != 0 || b.X < zone.X-.02 || b.X+b.W > zone.X+zone.W+.02 || b.Y < zone.Y-.02 || b.Y > zone.Y+zone.H+.02 || b.Y+b.H <= zone.Y+zone.H+.02 {
				return false
			}
			textOverflow = true
			continue
		case item.Shape != nil:
			b = item.Shape.Record.Rect
			if flowOverflow && item.Shape.Record.ID == p.ID+".container" && b == p.Bounds {
				flowContainer = true
				continue
			}
		case item.Table != nil:
			b = item.Table.Rect
		case item.Chart != nil:
			b = item.Chart.Rect
		case item.Image != nil:
			im := item.Image
			if im.X == nil || im.Y == nil || im.W == nil || im.H == nil {
				return false
			}
			b = Rect{im.X.Val * 72, im.Y.Val * 72, im.W.Val * 72, im.H.Val * 72}
		default:
			return false
		}
		if !inside(b, zone) {
			return false
		}
	}
	return textOverflow || flowOverflow && flowContainer
}

func buildWithSlideDensities(bundle string, s *Source, t *Typography, doc Document, engine string, assets map[string]AssetData) ([]byte, Report, error) {
	var report Report
	if err := validateTypographyDensity(s.Tokens.Density); err != nil {
		return nil, report, err
	}
	resolved := doc
	resolved.Slides = append([]SlideSpec(nil), doc.Slides...)
	records := make([]SlideDensityRecord, len(doc.Slides))
	reasons := make([]string, len(doc.Slides))
	for i, slide := range doc.Slides {
		dr, err := slideDensity(s, slide)
		if err != nil {
			return nil, report, fmt.Errorf("slide %s: %w", slide.ID, err)
		}
		if dr.Auto {
			_, tier, _ := densityLevel(dr.Requested)
			_, maxTier, _ := densityLevel(dr.Limit)
			for {
				trial := slide
				trial.Density = dr.Resolved
				disabled := false
				trial.AutoDensity = &disabled
				probe := doc
				probe.Sections = nil
				probe.Slides = []SlideSpec{trial}
				_, _, err = buildWithTypography(bundle, s, t, probe, engine, assets, true)
				if err == nil {
					break
				}
				if !isDensityFitError(err) {
					return nil, report, err
				}
				if tier == maxTier {
					if maxTier < 2 {
						return nil, report, fmt.Errorf("density.limit_exhausted: slide %s does not fit at %s (source densityLimit %s); split the slide, shorten content or choose another template; attempted steps %v; %w", slide.ID, dr.Resolved, dr.Limit, dr.Steps, err)
					}
					return nil, report, fmt.Errorf("density.dense_body_overflow: split the slide or shorten its content; %w", err)
				}
				reasons[i] = err.Error()
				tier++
				next := s.Tokens.Density.Levels[tier]
				dr.Steps = append(dr.Steps, DensityStep{From: dr.Resolved, To: next, Reason: err.Error()})
				dr.Resolved = next
			}
		}
		dr.Adjusted = dr.Requested != dr.Resolved
		records[i] = dr
		// Old bundles retain their original input and serialized chrome exactly.
		if s.Tokens.Density != nil || slide.Density != "" {
			resolved.Slides[i].Density = dr.Resolved
		}
	}
	deck, report, err := buildWithTypography(bundle, s, t, resolved, engine, assets, false)
	if err != nil {
		return nil, report, err
	}
	for i, dr := range records {
		slide := doc.Slides[i]
		if s.Tokens.Density != nil || slide.Density != "" || slide.Frame.HeaderDensity != "" || slide.AutoDensity != nil {
			record := dr
			report.Slides[i].Density = &record
		}
		if dr.Adjusted {
			report.DensityAdjustments = append(report.DensityAdjustments, SlideDensityAdjustment{SlideID: slide.ID, Requested: dr.Requested, Resolved: dr.Resolved, Reason: reasons[i], Steps: dr.Steps})
			report.Warnings = append(report.Warnings, fmt.Sprintf("Slide %s: body density changed %s → %s to fit its content (header %s).", slide.ID, dr.Requested, dr.Resolved, dr.Header))
		}
	}
	return deck, report, nil
}
