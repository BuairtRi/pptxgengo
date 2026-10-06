package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDensityCalibrationSourceIsolation(t *testing.T) {
	source := densityTestSource(t)
	root := filepath.Join(densityTestBundle(), "fonts")
	base, err := NewTypographyEngine(root, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	current, err := NewSourceTypographyEngine(source, root, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	expected := 306
	if source.densityVisualRules {
		expected = 312
	}
	if len(base.anchors) != 17 || len(current.densityAnchors) != expected {
		t.Fatalf("unexpected calibration counts: base=%d supplement=%d", len(base.anchors), len(current.densityAnchors))
	}
	for key, anchor := range base.anchors {
		if !reflect.DeepEqual(current.anchors[key], anchor) || current.anchorCalibrationSHA(anchor) != CandidateCalibrationSHA {
			t.Fatalf("base anchor changed: %s", key)
		}
	}
	oldVisualSource := *source
	oldVisualSource.densityVisualRules = false
	oldVisual, err := NewSourceTypographyEngine(&oldVisualSource, root, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldVisual.densityAnchors) != 306 || oldVisual.densitySupplementSHA != DensityCalibrationSHA {
		t.Fatal("modern quote measurements changed the historical density supplement")
	}
	if source.densityVisualRules && current.densitySupplementSHA != DensityQuoteCalibrationSHA {
		t.Fatal("modern quote calibration provenance is missing")
	}
	copySource := *source
	copySource.Tokens = source.Tokens
	copySource.Tokens.Density = nil
	legacy, err := NewSourceTypographyEngine(&copySource, root, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(legacy.densityAnchors) != 0 || !reflect.DeepEqual(legacy.anchors, base.anchors) {
		t.Fatal("density supplement changed a historical source")
	}
	oldEngine, err := NewSourceTypographyEngine(source, root, Engine)
	if err != nil {
		t.Fatal(err)
	}
	if len(oldEngine.densityAnchors) != 0 {
		t.Fatal("supplement changed the legacy engine")
	}
}

// This opt-in derivation uses the same preserved native measurements as the
// calibration closure tests. It writes a candidate, never updates calibration.
func TestDeriveDensityNativeCapture(t *testing.T) {
	dir := os.Getenv("WMDS_DENSITY_CAPTURE_DIR")
	if dir == "" {
		t.Skip("set WMDS_DENSITY_CAPTURE_DIR to derive captured controls")
	}
	anchors := deriveDensityNativeCapture(t, dir)
	data, err := json.MarshalIndent(anchors, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "derived-anchors.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("WMDS_DENSITY_CAPTURE_CANDIDATE") == "1" {
		var candidate map[string]json.RawMessage
		if err = json.Unmarshal(densityCalibrationData, &candidate); err != nil {
			t.Fatal(err)
		}
		var existing []VerticalAnchor
		if err = json.Unmarshal(candidate["anchors"], &existing); err != nil {
			t.Fatal(err)
		}
		keys := map[string]bool{}
		for _, a := range existing {
			keys[anchorKey(a.FontSHA, a.Size, a.Leading)] = true
		}
		for _, a := range anchors {
			if keys[anchorKey(a.FontSHA, a.Size, a.Leading)] {
				t.Fatal("candidate capture duplicates an existing anchor")
			}
			existing = append(existing, a)
		}
		candidate["anchors"], _ = json.Marshal(existing)
		var provenance map[string]json.RawMessage
		if err = json.Unmarshal(candidate["provenance"], &provenance); err != nil {
			t.Fatal(err)
		}
		var sets []map[string]any
		if err = json.Unmarshal(provenance["capture_sets"], &sets); err != nil {
			t.Fatal(err)
		}
		set := map[string]any{"name": filepath.Base(dir), "control_count": len(anchors), "native_measurement_count": len(anchors)}
		for file, key := range map[string]string{"control-manifest.json": "manifest_sha256", "native-measurements.json": "native_measurements_sha256", "bare-controls.pptx": "bare_pptx_sha256", "native-render.pdf": "pdf_sha256", "native-render-receipt.json": "render_receipt_sha256", "baselines.json": "baselines_sha256"} {
			b, e := os.ReadFile(filepath.Join(dir, file))
			if e != nil {
				t.Fatal(e)
			}
			set[key] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		var manifest struct {
			Source   string           `json:"source_commit"`
			Controls []densityControl `json:"controls"`
		}
		b, e := os.ReadFile(filepath.Join(dir, "control-manifest.json"))
		if e != nil {
			t.Fatal(e)
		}
		if err = json.Unmarshal(b, &manifest); err != nil {
			t.Fatal(err)
		}
		pages := 0
		for _, control := range manifest.Controls {
			if control.Page > pages {
				pages = control.Page
			}
		}
		set["visible_slide_count"], set["source_commit"] = pages, manifest.Source
		provenance["capture_sets"], _ = json.Marshal(append(sets, set))
		candidate["provenance"], _ = json.Marshal(provenance)
		b, err = json.MarshalIndent(candidate, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "calibration-candidate.json"), b, 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("candidate %d anchors sha256 %x", len(existing), sha256.Sum256(b))
	}
	t.Logf("independently derived %d native anchors", len(anchors))
}

func deriveDensityNativeCapture(t *testing.T, dir string) []VerticalAnchor {
	t.Helper()
	read := func(name string, dst any) {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, dst); err != nil {
			t.Fatal(err)
		}
	}
	var manifest struct {
		Controls []densityControl `json:"controls"`
		BareSHA  string           `json:"bare_pptx_sha256"`
	}
	read("control-manifest.json", &manifest)
	bare, err := os.ReadFile(filepath.Join(dir, "bare-controls.pptx"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(bare)) != manifest.BareSHA {
		t.Fatal("native capture input identity mismatch", err)
	}
	type bounds struct{ Left, Top, Width, Height float64 }
	type character struct {
		Text   string  `json:"text"`
		Font   string  `json:"font_name"`
		Size   float64 `json:"font_size_pt"`
		Bold   bool    `json:"bold"`
		Italic bool    `json:"italic"`
		Bounds bounds  `json:"bounds"`
	}
	var native struct {
		Measurements []struct {
			Page       int         `json:"slide_index"`
			ID         string      `json:"shape_name"`
			Text       string      `json:"text"`
			Frame      bounds      `json:"shape_frame"`
			Characters []character `json:"characters"`
			Font       struct {
				Name string  `json:"name"`
				Size float64 `json:"size_pt"`
				Bold bool    `json:"bold"`
			} `json:"font"`
			Paragraphs []struct {
				Leading  float64 `json:"space_within"`
				Relative bool    `json:"line_rule_within"`
			} `json:"paragraphs"`
		} `json:"measurements"`
	}
	read("native-measurements.json", &native)
	var pdf struct {
		SHA   string `json:"pdf_sha256"`
		Pages []struct {
			Page    int `json:"page_number"`
			Origins []struct {
				Baseline float64 `json:"baseline_pt"`
			} `json:"origins"`
		} `json:"pages"`
	}
	read("baselines.json", &pdf)
	pdfBytes, err := os.ReadFile(filepath.Join(dir, "native-render.pdf"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(pdfBytes)) != pdf.SHA {
		t.Fatal("PDF baseline input identity mismatch", err)
	}
	if len(native.Measurements) != len(manifest.Controls) || len(manifest.Controls) == 0 {
		t.Fatal("capture count mismatch")
	}
	pageOrigins := map[int][]float64{}
	for _, page := range pdf.Pages {
		for _, origin := range page.Origins {
			pageOrigins[page.Page] = append(pageOrigins[page.Page], origin.Baseline)
		}
	}
	var anchors []VerticalAnchor
	seen := map[string]bool{}
	for i, control := range manifest.Controls {
		n := native.Measurements[i]
		text := strings.ReplaceAll(control.Text, "\\n", "\n")
		text = strings.ReplaceAll(text, "\n", "\r")
		if n.ID != control.ID || n.Page != control.Page || n.Text != text || seen[n.ID] {
			t.Fatalf("native control mismatch: %s", control.ID)
		}
		seen[n.ID] = true
		if math.Abs(n.Frame.Left-control.Rect.X) > .001 || math.Abs(n.Frame.Top-control.Rect.Y) > .001 || math.Abs(n.Frame.Width-control.Rect.W) > .001 || math.Abs(n.Frame.Height-control.Rect.H) > .001 {
			t.Fatalf("native frame mismatch: %s", control.ID)
		}
		if n.Font.Name != control.Font.Typeface || math.Abs(n.Font.Size-control.Style.Size) > .011 || n.Font.Bold != control.Font.Bold {
			t.Fatalf("native font mismatch: %s", control.ID)
		}
		if len(n.Paragraphs) != 2 {
			t.Fatalf("expected two native paragraphs: %s", control.ID)
		}
		for _, p := range n.Paragraphs {
			if p.Relative || math.Abs(p.Leading-control.Style.Leading) > .011 {
				t.Fatalf("native paragraph mismatch: %s", control.ID)
			}
		}
		lines := strings.Split(text, "\r")
		if len(lines) != 2 {
			t.Fatal("expected two control lines")
		}
		var chars []character
		for _, c := range n.Characters {
			if c.Text != "\r" && c.Text != "\n" {
				chars = append(chars, c)
			}
		}
		if len(chars) != len([]rune(lines[0]+lines[1])) {
			t.Fatalf("native character count mismatch: %s", control.ID)
		}
		tops := []float64{math.Inf(1), math.Inf(1)}
		bottom := math.Inf(-1)
		offset := 0
		for line, lineText := range lines {
			for _, ch := range lineText {
				c := chars[offset]
				offset++
				if c.Text != string(ch) || c.Font != control.Font.Typeface || math.Abs(c.Size-control.Style.Size) > .011 || c.Bold != control.Font.Bold || c.Italic != control.Font.NativeItalic {
					t.Fatalf("native character assignment mismatch: %s", control.ID)
				}
				if strings.TrimSpace(c.Text) == "" {
					continue
				}
				tops[line] = math.Min(tops[line], c.Bounds.Top)
				if line == 1 {
					bottom = math.Max(bottom, c.Bounds.Top+c.Bounds.Height)
				}
			}
		}
		origins := pageOrigins[control.Page]
		if len(origins) < 2 {
			t.Fatalf("missing PDF origins: %s", control.ID)
		}
		pitch := tops[1] - tops[0]
		if math.Abs((origins[1]-origins[0])-pitch) > .25 {
			t.Fatalf("native/PDF pitch mismatch: %s", control.ID)
		}
		anchor := VerticalAnchor{FontSHA: control.Font.SHA256, Size: control.Style.Size, Leading: control.Style.Leading, Baseline: origins[0] - n.Frame.Top, TerminalHeight: bottom - tops[1], ProbeID: control.ID, EffectiveLeading: pitch}
		for _, value := range []float64{anchor.Baseline, anchor.TerminalHeight, anchor.EffectiveLeading} {
			if value <= 0 || math.IsInf(value, 0) || math.IsNaN(value) {
				t.Fatalf("invalid derived anchor: %s", control.ID)
			}
		}
		anchors = append(anchors, anchor)
		pageOrigins[control.Page] = origins[2:]
	}
	for page, origins := range pageOrigins {
		if len(origins) != 0 {
			t.Fatalf("unmatched PDF origins on page %d", page)
		}
	}
	return anchors
}

func TestDensityCalibrationObservedPitchAndProvenance(t *testing.T) {
	source := densityTestSource(t)
	typography, err := NewSourceTypographyEngine(source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	st := Style{Family: "IBM Plex Sans", Weight: 500, Size: 10, Leading: 11.5}
	layout, err := typography.Measure("Agyp Q012\nAgyp Q012", st, 846)
	if err != nil {
		t.Fatal(err)
	}
	if len(layout.Lines) != 2 || math.Abs(layout.Lines[1].Baseline-layout.Lines[0].Baseline-12) > .001 {
		t.Fatalf("observed native pitch was not used: %+v", layout.Lines)
	}
	if layout.Style.Leading != 11.5 || layout.CalibrationSHA256 != densityCalibrationForSource(source) || layout.NativeQualified {
		t.Fatalf("authored spacing or qualification/provenance changed: %+v", layout)
	}
	if layout.AllocationHeight < 24 {
		t.Fatal("allocation underestimates observed spacing")
	}
	st.Leading = 11.51
	unknown, err := typography.Measure("Agyp Q012", st, 846)
	if err != nil {
		t.Fatal(err)
	}
	if unknown.CalibrationSHA256 != "" {
		t.Fatal("nearby unobserved leading was calibrated")
	}
}

func TestDensityQuoteCalibrationExactPairsAndHistoricalIsolation(t *testing.T) {
	source := densityTestSource(t)
	if !source.densityVisualRules {
		t.Skip("modern quote-density rules required")
	}
	modern, err := NewSourceTypographyEngine(source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	old := *source
	old.densityVisualRules = false
	historical, err := NewSourceTypographyEngine(&old, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		Controls []densityControl `json:"controls"`
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(densityTestBundle()), "native-density-controls", "quote6", "control-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &capture); err != nil {
		t.Fatal(err)
	}
	for _, control := range capture.Controls {
		key := anchorKey(control.Font.SHA256, control.Style.Size, control.Style.Leading)
		anchor, exists := modern.anchors[key]
		if !exists || modern.anchorCalibrationSHA(anchor) != DensityQuoteCalibrationSHA {
			t.Fatalf("quote pair lacks native provenance: %s", control.ID)
		}
		if _, exists = historical.anchors[key]; exists {
			t.Fatalf("quote pair leaked into historical source: %s", control.ID)
		}
		layout, err := modern.Measure("Agyp Q012", control.Style, 846)
		if err != nil || layout.CalibrationSHA256 != DensityQuoteCalibrationSHA || layout.NativeQualified {
			t.Fatalf("quote pair measurement provenance mismatch: %s: %+v %v", control.ID, layout, err)
		}
	}
}

func TestDensityCalibrationRejectsInvalidSupplementWithoutMutation(t *testing.T) {
	typography, err := NewTypographyEngine(filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	var original map[string]interface{}
	if err = json.Unmarshal(densityCalibrationData, &original); err != nil {
		t.Fatal(err)
	}
	before := len(typography.anchors)
	for _, kind := range []string{"duplicate", "duplicate-probe", "unknown-font", "zero-pitch", "base-override", "wrong-base"} {
		t.Run(kind, func(t *testing.T) {
			var fixture map[string]interface{}
			_ = json.Unmarshal(densityCalibrationData, &fixture)
			anchors := fixture["anchors"].([]interface{})
			first := anchors[0].(map[string]interface{})
			switch kind {
			case "duplicate":
				fixture["anchors"] = append(anchors, first)
			case "duplicate-probe":
				anchors[1].(map[string]interface{})["source_probe_id"] = first["source_probe_id"]
			case "unknown-font":
				first["font_sha256"] = "unknown"
			case "zero-pitch":
				first["observed_leading_pt"] = 0
			case "wrong-base":
				fixture["base_calibration_sha256"] = "unknown"
			case "base-override":
				for _, anchor := range typography.anchors {
					first["font_sha256"], first["size_pt"], first["leading_pt"] = anchor.FontSHA, anchor.Size, anchor.Leading
					break
				}
			}
			data, _ := json.Marshal(fixture)
			if typography.applyDensityCalibration(data) == nil {
				t.Fatal("invalid supplement accepted")
			}
			if len(typography.anchors) != before || len(typography.densityAnchors) != 0 {
				t.Fatal("failed validation mutated anchors")
			}
		})
	}
}

func TestDensityCalibrationNativeEvidenceClosure(t *testing.T) {
	t.Run("historical306", func(t *testing.T) {
		densityCalibrationEvidenceClosure(t, densityCalibrationData, "density_calibration.json")
	})
	t.Run("modern312", func(t *testing.T) {
		densityCalibrationEvidenceClosure(t, densityQuoteCalibrationData, "density_quote_calibration.json")
	})
}

func densityCalibrationEvidenceClosure(t *testing.T, data []byte, filename string) {
	root := filepath.Join(filepath.Dir(densityTestBundle()), "native-density-controls")
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	if !bytes.Equal(read(filename), data) {
		t.Fatal("embedded anchors differ from durable native evidence")
	}
	var supplement struct {
		Anchors    []VerticalAnchor `json:"anchors"`
		Provenance struct {
			CaptureSets []map[string]interface{} `json:"capture_sets"`
		} `json:"provenance"`
	}
	if err := json.Unmarshal(data, &supplement); err != nil {
		t.Fatal(err)
	}
	probes := map[string]densityControl{}
	for _, set := range supplement.Provenance.CaptureSets {
		name := set["name"].(string)
		files := map[string]string{
			"control-manifest.json":      "manifest_sha256",
			"native-measurements.json":   "native_measurements_sha256",
			"bare-controls.pptx":         "bare_pptx_sha256",
			"native-render.pdf":          "pdf_sha256",
			"native-render-receipt.json": "render_receipt_sha256",
			"baselines.json":             "baselines_sha256",
		}
		for file, field := range files {
			if actual := fmt.Sprintf("%x", sha256.Sum256(read(filepath.Join(name, file)))); actual != set[field] {
				t.Fatalf("native evidence changed: %s/%s", name, file)
			}
		}
		var manifest struct {
			Controls []densityControl `json:"controls"`
		}
		if err := json.Unmarshal(read(filepath.Join(name, "control-manifest.json")), &manifest); err != nil {
			t.Fatal(err)
		}
		if len(manifest.Controls) != int(set["control_count"].(float64)) {
			t.Fatal("native control count changed")
		}
		for _, control := range manifest.Controls {
			if _, exists := probes[control.ID]; exists {
				t.Fatal("duplicate durable probe ID")
			}
			probes[control.ID] = control
		}
	}
	if len(probes) != len(supplement.Anchors) {
		t.Fatal("native evidence inventory differs from anchors")
	}
	for _, anchor := range supplement.Anchors {
		probe, exists := probes[anchor.ProbeID]
		if !exists || anchorKey(probe.Font.SHA256, probe.Style.Size, probe.Style.Leading) != anchorKey(anchor.FontSHA, anchor.Size, anchor.Leading) {
			t.Fatalf("anchor has no matching captured control: %s", anchor.ProbeID)
		}
	}
	anchorsByID := map[string]VerticalAnchor{}
	for _, anchor := range supplement.Anchors {
		anchorsByID[anchor.ProbeID] = anchor
	}
	for _, set := range supplement.Provenance.CaptureSets {
		for _, derived := range deriveDensityNativeCapture(t, filepath.Join(root, set["name"].(string))) {
			stored := anchorsByID[derived.ProbeID]
			if stored.FontSHA != derived.FontSHA || stored.Size != derived.Size || stored.Leading != derived.Leading {
				t.Fatalf("native derivation identity changed: %s", derived.ProbeID)
			}
			for i, value := range []float64{stored.Baseline - derived.Baseline, stored.TerminalHeight - derived.TerminalHeight, stored.EffectiveLeading - derived.EffectiveLeading} {
				if math.Abs(value) > .000001 {
					t.Fatalf("native derivation differs: %s field %d", derived.ProbeID, i)
				}
			}
		}
	}
}
