package wmdesign

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
)

// This supplement records controls captured in local PowerPoint. It does not
// establish native font file identity or qualify arbitrary slide content.
// Historical sources continue to use only their frozen base calibration.
//
//go:embed density_calibration.json
var densityCalibrationData []byte

//go:embed density_quote_calibration.json
var densityQuoteCalibrationData []byte

const DensityCalibrationSHA = "7d15b02b2296c5851483f3b34e8624d1eb2a0bec7c273cc14028a98d138ac5f1"
const DensityQuoteCalibrationSHA = "aad404e76ed5634a494a25f101db1f210e92570abe3ee3162cbabf93ce24ecae"

func densityCalibrationForSource(source *Source) string {
	if source != nil && source.Tokens.Density != nil {
		if source.densityVisualRules {
			return DensityQuoteCalibrationSHA
		}
		return DensityCalibrationSHA
	}
	return ""
}

func NewSourceTypographyEngine(source *Source, root, engine string) (*Typography, error) {
	t, err := NewTypographyEngine(root, engine)
	if err != nil {
		return nil, err
	}
	if source == nil || source.Tokens.Density == nil || engine != CandidateEngine {
		return t, nil
	}
	data, expectedSHA := densityCalibrationData, DensityCalibrationSHA
	if source.densityVisualRules {
		data, expectedSHA = densityQuoteCalibrationData, DensityQuoteCalibrationSHA
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != expectedSHA {
		return nil, fmt.Errorf("text.density_calibration_drift")
	}
	if err = t.applyDensityCalibration(data); err != nil {
		return nil, err
	}
	t.densitySupplementSHA = expectedSHA
	return t, nil
}

func (t *Typography) applyDensityCalibration(data []byte) error {
	var supplement struct {
		Schema  string           `json:"schema"`
		BaseSHA string           `json:"base_calibration_sha256"`
		Anchors []VerticalAnchor `json:"anchors"`
	}
	if err := json.Unmarshal(data, &supplement); err != nil {
		return fmt.Errorf("text.density_calibration_decode: %w", err)
	}
	if supplement.Schema != "pptxgengo.wmds-density-native-anchors.v1" || supplement.BaseSHA != CandidateCalibrationSHA || len(supplement.Anchors) == 0 {
		return fmt.Errorf("text.invalid_density_calibration_schema")
	}
	fonts := map[string]bool{}
	for _, face := range t.faces {
		fonts[face.Identity.SHA256] = true
	}
	newAnchors := map[string]VerticalAnchor{}
	probeIDs := map[string]bool{}
	for _, anchor := range supplement.Anchors {
		values := []float64{anchor.Size, anchor.Leading, anchor.Baseline, anchor.TerminalHeight, anchor.EffectiveLeading}
		for _, value := range values {
			if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("text.invalid_density_vertical_anchor: %s", anchor.ProbeID)
			}
		}
		if !fonts[anchor.FontSHA] || anchor.ProbeID == "" {
			return fmt.Errorf("text.invalid_density_anchor_identity: %s", anchor.ProbeID)
		}
		if probeIDs[anchor.ProbeID] {
			return fmt.Errorf("text.duplicate_density_probe_id: %s", anchor.ProbeID)
		}
		probeIDs[anchor.ProbeID] = true
		key := anchorKey(anchor.FontSHA, anchor.Size, anchor.Leading)
		if _, exists := t.anchors[key]; exists {
			return fmt.Errorf("text.density_anchor_overrides_base: %s", anchor.ProbeID)
		}
		if _, exists := newAnchors[key]; exists {
			return fmt.Errorf("text.duplicate_density_vertical_anchor: %s", anchor.ProbeID)
		}
		newAnchors[key] = anchor
	}
	// Validate the whole supplement before mutating a shaping engine.
	t.densityAnchors = map[string]bool{}
	for key, anchor := range newAnchors {
		t.anchors[key] = anchor
		t.densityAnchors[key] = true
	}
	return nil
}

func (t *Typography) anchorCalibrationSHA(anchor VerticalAnchor) string {
	if t.densityAnchors[anchorKey(anchor.FontSHA, anchor.Size, anchor.Leading)] {
		if t.densitySupplementSHA != "" {
			return t.densitySupplementSHA
		}
		return DensityCalibrationSHA
	}
	return CandidateCalibrationSHA
}

func effectiveAnchorLeading(anchor VerticalAnchor, authored float64) float64 {
	if anchor.EffectiveLeading > 0 {
		return anchor.EffectiveLeading
	}
	return authored
}
