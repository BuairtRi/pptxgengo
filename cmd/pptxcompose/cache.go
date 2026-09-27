package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buairtri/pptxgengo/internal/compose"
)

type environmentFont struct {
	Family         string `json:"family"`
	PostscriptName string `json:"postscript_name"`
	Style          string `json:"style"`
	Path           string `json:"path"`
	SHA            string `json:"sha256"`
}
type measurementEnvironment struct {
	Schema            string            `json:"schema"`
	OS                string            `json:"os"`
	PowerPointVersion string            `json:"powerpoint_version"`
	PowerPointBuild   string            `json:"powerpoint_build"`
	Fonts             []environmentFont `json:"fonts"`
	AdapterSHA        string            `json:"adapter_sha256"`
	ExecutableSHA     string            `json:"executable_sha256"`
	InspectorSHA      string            `json:"inspector_sha256"`
}

func currentEnvironment(adapter string) (*measurementEnvironment, error) {
	inspector := "scripts/compose-environment.swift"
	source, e := os.ReadFile(inspector)
	if e != nil {
		return nil, e
	}
	b, e := exec.Command("swift", inspector).Output()
	if e != nil {
		return nil, fmt.Errorf("inspect native measurement environment: %w", e)
	}
	var env measurementEnvironment
	if e = json.Unmarshal(b, &env); e != nil {
		return nil, e
	}
	if env.OS == "" || env.PowerPointVersion == "" || env.PowerPointBuild == "" || len(env.Fonts) != 4 {
		return nil, fmt.Errorf("incomplete native environment")
	}
	for i := range env.Fonts {
		data, e := os.ReadFile(env.Fonts[i].Path)
		if e != nil {
			return nil, e
		}
		env.Fonts[i].SHA = hash(data)
	}
	data, e := os.ReadFile(adapter)
	if e != nil {
		return nil, e
	}
	env.AdapterSHA = hash(data)
	executable, e := os.Executable()
	if e != nil {
		return nil, e
	}
	data, e = os.ReadFile(executable)
	if e != nil {
		return nil, e
	}
	env.ExecutableSHA = hash(data)
	env.Schema = "pptxgengo.measurement-environment.v1"
	env.InspectorSHA = hash(source)
	return &env, nil
}
func environmentKey(env *measurementEnvironment) string { return hash(jsonBytes(env)) }

// Every value that can affect the probe rendering belongs here. IDs locate
// observations but do not alter glyph geometry. This version supports Arial only.
type textContract struct {
	Text       string                  `json:"text"`
	Paragraphs []compose.ParagraphSpec `json:"paragraphs,omitempty"`
	Width      float64                 `json:"width_pt"`
	InsetX     float64                 `json:"inset_x"`
	InsetY     float64                 `json:"inset_y"`
	FontFace   string                  `json:"font_face"`
	FontSize   float64                 `json:"font_size_pt"`
	Bold       bool                    `json:"bold"`
	Align      string                  `json:"align"`
	Foreground string                  `json:"foreground"`
	Background string                  `json:"background"`
}

func contract(q compose.ProbeRequest) textContract {
	return textContract{q.Text, q.Paragraphs, q.TextWidthPt, q.HorizontalInsetPt, q.VerticalInsetPt, q.FontFace, q.FontSizePt, q.Bold, q.Align, q.Foreground, q.Background}
}
func contractKey(q compose.ProbeRequest) string { return hash(jsonBytes(contract(q))) }

type cacheEntry struct {
	Schema         string       `json:"schema"`
	EnvironmentSHA string       `json:"environment_sha256"`
	Contract       textContract `json:"contract"`
	Bundle         string       `json:"bundle"`
	Evidence       string       `json:"evidence"`
	EvidenceSHA    string       `json:"evidence_sha256"`
	RequestID      string       `json:"request_id"`
}
type cacheUse struct {
	RequestID       string `json:"request_id"`
	ContractSHA     string `json:"contract_sha256"`
	SourceRequestID string `json:"source_request_id"`
	Bundle          string `json:"bundle"`
	Evidence        string `json:"evidence"`
	EvidenceSHA     string `json:"evidence_sha256"`
}
type validatedSource struct {
	manifest     manifest
	evidence     evidence
	measurements compose.Measurements
	evidenceSHA  string
}

func validateProbeSource(dir, evPath string) (validatedSource, error) {
	var result validatedSource
	mb, e := readJSON(filepath.Join(dir, "manifest.json"), &result.manifest)
	if e != nil {
		return result, e
	}
	eb, e := readJSON(evPath, &result.evidence)
	if e != nil {
		return result, e
	}
	result.evidenceSHA = hash(eb)
	pm, ev := result.manifest, result.evidence
	dp, e := deckPath(dir, pm)
	if e != nil {
		return result, e
	}
	db, e := os.ReadFile(dp)
	if e != nil {
		return result, e
	}
	if pm.Schema != "pptxgengo.compose-bundle.v1" || pm.Plan != nil || len(pm.Requests) == 0 || ev.Schema != "pptxgengo.compose-evidence.v1" || ev.SpecSHA != pm.SpecSHA || ev.ManifestSHA != hash(mb) || ev.DeckSHA != hash(db) || pm.DeckSHA != hash(db) {
		return result, fmt.Errorf("stale or mismatched cache source evidence")
	}
	// Bind original requests to the actual probe manifest, not only to its hashes.
	expected := probeSlides(pm.Requests)
	if !bytes.Equal(jsonBytes(expected), jsonBytes(pm.Slides)) {
		return result, fmt.Errorf("cache source requests do not describe probe deck")
	}
	var native nativeResult
	if e = json.Unmarshal(ev.Native, &native); e != nil {
		return result, e
	}
	if native.Schema != "pptxgengo.compose-text-measurement.v8" {
		return result, fmt.Errorf("cache requires v8 native character/paragraph/style/bullet evidence")
	}
	result.measurements, e = checkNative(pm, native, false)
	if e != nil {
		return result, e
	}
	if ev.Environment == nil || ev.Environment.Schema != "pptxgengo.measurement-environment.v1" || ev.AdapterSHA != ev.Environment.AdapterSHA {
		return result, fmt.Errorf("cache requires native evidence with an environment fingerprint")
	}
	if pm.Environment != nil && environmentKey(pm.Environment) != environmentKey(ev.Environment) {
		return result, fmt.Errorf("probe environment differs from measured environment")
	}
	return result, nil
}
func cacheLookup(dir string, env *measurementEnvironment, requests []compose.ProbeRequest) (compose.Measurements, []compose.ProbeRequest, []cacheUse, error) {
	measured := compose.Measurements{ByRequestID: map[string]compose.Measurement{}}
	var missing []compose.ProbeRequest
	missingKeys := map[string]bool{}
	var uses []cacheUse
	sources := map[string]validatedSource{}
	envKey := environmentKey(env)
	for _, q := range requests {
		path := filepath.Join(dir, "entries", envKey, contractKey(q)+".json")
		var entry cacheEntry
		_, e := readJSON(path, &entry)
		if os.IsNotExist(e) {
			if !missingKeys[contractKey(q)] {
				missing = append(missing, q)
				missingKeys[contractKey(q)] = true
			}
			continue
		}
		if e != nil {
			return measured, nil, nil, e
		}
		if entry.Schema != "pptxgengo.measurement-cache-entry.v1" || entry.EnvironmentSHA != envKey || !bytes.Equal(jsonBytes(entry.Contract), jsonBytes(contract(q))) {
			return measured, nil, nil, fmt.Errorf("cache contract mismatch: %s", q.ID)
		}
		sourceKey := entry.Bundle + "\x00" + entry.Evidence
		source, ok := sources[sourceKey]
		if !ok {
			source, e = validateProbeSource(entry.Bundle, entry.Evidence)
			if e != nil {
				return measured, nil, nil, e
			}
			sources[sourceKey] = source
		}
		if source.evidenceSHA != entry.EvidenceSHA || environmentKey(source.evidence.Environment) != envKey {
			return measured, nil, nil, fmt.Errorf("cache source hash/environment mismatch: %s", q.ID)
		}
		found := false
		for _, original := range source.manifest.Requests {
			if original.ID == entry.RequestID && contractKey(original) == contractKey(q) {
				found = true
				break
			}
		}
		m, ok := source.measurements.ByRequestID[entry.RequestID]
		if !found || !ok {
			return measured, nil, nil, fmt.Errorf("cache source request missing/mismatched: %s", q.ID)
		}
		measured.ByRequestID[q.ID] = m
		uses = append(uses, cacheUse{q.ID, contractKey(q), entry.RequestID, entry.Bundle, entry.Evidence, entry.EvidenceSHA})
	}
	return measured, missing, uses, nil
}
func cacheImport(dir, bundleDir, evidencePath string, env *measurementEnvironment) (int, error) {
	source, e := validateProbeSource(bundleDir, evidencePath)
	if e != nil {
		return 0, e
	}
	if environmentKey(source.evidence.Environment) != environmentKey(env) {
		return 0, fmt.Errorf("cannot import evidence from a different native environment")
	}
	bundleDir, e = filepath.Abs(bundleDir)
	if e != nil {
		return 0, e
	}
	evidencePath, e = filepath.Abs(evidencePath)
	if e != nil {
		return 0, e
	}
	entryDir := filepath.Join(dir, "entries", environmentKey(env))
	if e = os.MkdirAll(entryDir, 0755); e != nil {
		return 0, e
	}
	count := 0
	for _, q := range source.manifest.Requests {
		path := filepath.Join(entryDir, contractKey(q)+".json")
		if _, e = os.Stat(path); e == nil { // Validate existing provenance before keeping it.
			_, missing, _, e := cacheLookup(dir, env, []compose.ProbeRequest{q})
			if e != nil {
				return count, e
			}
			if len(missing) != 0 {
				return count, fmt.Errorf("cache entry disappeared")
			}
			continue
		} else if !os.IsNotExist(e) {
			return count, e
		}
		entry := cacheEntry{"pptxgengo.measurement-cache-entry.v1", environmentKey(env), contract(q), bundleDir, evidencePath, source.evidenceSHA, q.ID}
		if e = writeNew(path, jsonBytes(entry)); e != nil {
			return count, e
		}
		count++
	}
	return count, nil
}
