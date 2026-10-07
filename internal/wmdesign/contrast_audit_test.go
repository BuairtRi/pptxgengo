package wmdesign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type contrastSpecimen struct {
	Template       string          `json:"template"`
	Density        string          `json:"body_density"`
	Header         string          `json:"header_density"`
	Limit          string          `json:"density_limit"`
	Status         string          `json:"status"`
	Unsupported    string          `json:"unsupported_reason,omitempty"`
	Checks         []contrastCheck `json:"checks"`
	CoverageErrors []string        `json:"coverage_errors,omitempty"`
}

func TestLibraryDensityAllTierContrastAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("source specimen rendering requires registered private branding; run make test-integration")
	}
	source := densityTestSource(t)
	typography, err := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := LibrarySourceReference(densityTestBundle(), "", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	var specimens []contrastSpecimen
	var failures, coverage []string
	for _, slide := range doc.Slides {
		frame, err := source.ResolveFrame(slide.Frame)
		if err != nil {
			t.Fatal(err)
		}
		for _, level := range source.Tokens.Density.Levels {
			limit, maxTier, err := densityLimit(slide.DensityLimit)
			if err != nil {
				t.Fatal(err)
			}
			_, tier, _ := densityLevel(level)
			if tier > maxTier {
				specimens = append(specimens, contrastSpecimen{Template: slide.TemplateBinding.Template, Density: level, Header: "comfortable", Limit: limit, Status: "unsupported", Unsupported: fmt.Sprintf("source densityLimit %s prohibits %s", limit, level)})
				continue
			}
			probe := &contrastProbe{}
			r := renderer{source: source, typeEngine: typography, bundle: densityTestBundle(), bodyDensity: level, headerDensity: "comfortable", densityScope: "body", contrastProbe: probe}
			if err := r.registerSceneTargets(slide, frame); err != nil {
				probe.CoverageErrors = append(probe.CoverageErrors, "target preparation: "+err.Error())
			}
			if !frame.Request.NoHeader {
				for _, header := range []struct{ id, text, role, ink string }{{"eyebrow", slide.Eyebrow, "eyebrow", "emphasis"}, {"title", slide.Title, frame.TitleStyle, "display"}} {
					if header.text == "" {
						continue
					}
					style, _ := r.headerStyle(header.role)
					ink, err := source.Ink(frame.Request.Surface, header.ink)
					if err != nil {
						t.Fatal(err)
					}
					r.auditTextContrast(header.id, style, ink, frame.Request.Surface)
				}
			}
			for _, node := range slide.Nodes {
				if node.Scene == nil {
					probe.CoverageErrors = append(probe.CoverageErrors, fmt.Sprintf("%s: unhandled non-scene node kind %s", node.ID, node.Kind))
					continue
				}
				var tag struct {
					X, W     float64
					On, Type string
				}
				if err := json.Unmarshal(node.Scene.Node, &tag); err != nil {
					t.Fatal(err)
				}
				surface, zone := frame.Request.Surface, frame.Body
				if tag.On != "" {
					surface = tag.On
				}
				if frame.Rail.W > 0 && tag.X >= frame.Rail.X-.02 && tag.X < frame.Rail.X+frame.Rail.W {
					surface = frame.Request.RailSurface
					zone = Rect{frame.Rail.X, 0, frame.Rail.W, frame.Rail.Y + frame.Rail.H}
				}
				if node.Scene.Allocation != nil {
					zone = *node.Scene.Allocation
				}
				ctx := SceneContext{Surface: surface, Zone: zone, Path: node.Scene.Path, Keys: node.Scene.Keys, Notes: node.Scene.Notes}
				if plan, err := r.planSceneNode(node.ID, node.Scene.Node, ctx); err != nil {
					probe.CoverageErrors = append(probe.CoverageErrors, fmt.Sprintf("%s: %v; later text in this node may be untraversed", node.ID, err))
				} else {
					r.sceneContext = ctx
					r.auditPlanContrastCoverage(plan)
					r.sceneContext = SceneContext{}
				}
			}
			if slide.DraftReview != nil {
				if _, err := r.planDraftReview(slide.ID, *slide.DraftReview); err != nil {
					probe.CoverageErrors = append(probe.CoverageErrors, err.Error())
				}
			}
			specimen := contrastSpecimen{Template: slide.TemplateBinding.Template, Density: level, Header: "comfortable", Limit: limit, Status: "checked", Checks: probe.Checks, CoverageErrors: probe.CoverageErrors}
			specimens = append(specimens, specimen)
			for _, check := range probe.Checks {
				if !check.Passed {
					failures = append(failures, fmt.Sprintf("%s/%s/%s: role%s %.3fpt weight%d %s on%s ratio%.4f <%.1f", specimen.Template, level, check.ID, check.Role, check.Size, check.Weight, check.Ink, check.Background, check.Ratio, check.Minimum))
				}
			}
			for _, gap := range probe.CoverageErrors {
				coverage = append(coverage, fmt.Sprintf("%s/%s: %s", specimen.Template, level, gap))
			}
		}
	}
	if len(specimens) != 1947 {
		t.Fatalf("contrast specimen cardinality %d", len(specimens))
	}
	out := os.Getenv("WMDS_CONTRAST_AUDIT_OUT")
	if out == "" {
		out = t.TempDir()
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(struct {
		SourceCommit string             `json:"source_commit"`
		Specimens    []contrastSpecimen `json:"specimens"`
		Failures     []string           `json:"failures"`
		Coverage     []string           `json:"coverage_errors"`
	}{source.Commit, specimens, failures, coverage}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "contrast-audit.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("contrast audit: %d specimens, %d contrast failures, %d coverage gaps; %s", len(specimens), len(failures), len(coverage), out)
	if len(coverage) > 0 {
		for _, failure := range coverage[:min(20, len(coverage))] {
			t.Log(failure)
		}
		t.Errorf("contrast coverage incomplete: %d gaps", len(coverage))
	}
	if len(failures) > 0 {
		for _, failure := range failures[:min(30, len(failures))] {
			t.Log(failure)
		}
		t.Errorf("density contrast failures: %d", len(failures))
	}
}

func TestContrastProbeRejectsInvalidInputsAndCollectsEveryFailure(t *testing.T) {
	s := densityTestSource(t)
	typography, err := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	probe := &contrastProbe{}
	r := renderer{source: s, typeEngine: typography, contrastProbe: probe}
	st, _ := s.Style("body")
	r.auditTextContrast("bad-surface", st, "FFFFFF", "unknown")
	if len(probe.CoverageErrors) != 1 {
		t.Fatal("invalid surface silently skipped contrast coverage")
	}
	for _, size := range []float64{0, -1} {
		invalid := st
		invalid.Size = size
		if _, err := r.measureText("copy", invalid, 270); err == nil {
			t.Fatal("invalid style accepted")
		}
	}
	for _, id := range []string{"first", "second"} {
		if !r.contrastAllows(id, st, "FFFFFF", "FFFFFF", 4.5) {
			t.Fatal("collector stopped instead of recording failure")
		}
	}
	if len(probe.Checks) != 2 || probe.Checks[0].Passed || probe.Checks[1].Passed {
		t.Fatal("contrast failures were suppressed")
	}
	r.contrastProbe = nil
	if r.contrastAllows("normal", st, "FFFFFF", "FFFFFF", 4.5) {
		t.Fatal("normal runtime contrast policy relaxed")
	}
	if !strings.Contains(probe.CoverageErrors[0], "bad-surface") {
		t.Fatal("coverage error identity lost")
	}
}
