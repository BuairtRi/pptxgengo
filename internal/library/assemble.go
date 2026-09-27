package library

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const AssemblySchema = "pptxgengo.library-assembly.v1"

var assemblySlideID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

type AssemblyConfig struct {
	Schema        string               `json:"schema"`
	NarrativePath string               `json:"narrative_path,omitempty"`
	Slides        []AssemblySlideInput `json:"slides"`
}

type AssemblySlideInput struct {
	ContractID string `json:"contract_id"`
	ValuesPath string `json:"values_path"`
	SlideID    string `json:"slide_id"`
}

type AssemblySlideTrace struct {
	Position     int            `json:"position"`
	SlideID      string         `json:"slide_id"`
	ContractID   string         `json:"contract_id"`
	ValuesPath   string         `json:"values_path"`
	ValuesSHA256 string         `json:"values_sha256"`
	SpecPath     string         `json:"spec_path"`
	SpecSHA256   string         `json:"spec_sha256"`
	TracePath    string         `json:"trace_path"`
	Selection    SelectionTrace `json:"selection"`
}

type AssemblyTrace struct {
	Schema          string               `json:"schema"`
	ConfigSHA256    string               `json:"config_sha256"`
	NarrativePath   string               `json:"narrative_path,omitempty"`
	NarrativeSHA256 string               `json:"narrative_sha256,omitempty"`
	SpecSHA256      string               `json:"spec_sha256"`
	Slides          []AssemblySlideTrace `json:"slides"`
	FitStatus       string               `json:"fit_status"`
	RequiredNext    []string             `json:"required_next"`
}

func readStrictJSON(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}

func resolveInput(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(base, path)
}

func writeJSON(path string, v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	b = append(b, '\n')
	return b, os.WriteFile(path, b, 0644)
}

// Assemble applies only contract-owned semantic slots and optional page labels.
// It stages the full result beside out and publishes one complete directory.
func (s Store) Assemble(indexPath, configPath, out string, allowUnqualified bool) (AssemblyTrace, error) {
	if configPath == "" || out == "" {
		return AssemblyTrace{}, fmt.Errorf("config and output required")
	}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		return AssemblyTrace{}, err
	}
	out, err = filepath.Abs(out)
	if err != nil {
		return AssemblyTrace{}, err
	}
	if _, err := os.Stat(out); err == nil {
		return AssemblyTrace{}, fmt.Errorf("output exists: %s", out)
	} else if !os.IsNotExist(err) {
		return AssemblyTrace{}, err
	}
	cb, err := os.ReadFile(configPath)
	if err != nil {
		return AssemblyTrace{}, err
	}
	var config AssemblyConfig
	if err := readStrictJSON(cb, &config); err != nil {
		return AssemblyTrace{}, fmt.Errorf("assembly config: %w", err)
	}
	if config.Schema != AssemblySchema || len(config.Slides) == 0 {
		return AssemblyTrace{}, fmt.Errorf("assembly schema and nonempty slides required")
	}
	base := filepath.Dir(configPath)
	seen := map[string]bool{}
	for _, item := range config.Slides {
		if item.ContractID == "" || item.ValuesPath == "" || !assemblySlideID.MatchString(item.SlideID) || seen[item.SlideID] {
			return AssemblyTrace{}, fmt.Errorf("invalid or duplicate assembly slide %q", item.SlideID)
		}
		seen[item.SlideID] = true
	}
	var narrativePath, narrativeHash string
	if config.NarrativePath != "" {
		narrativePath = resolveInput(base, config.NarrativePath)
		b, err := os.ReadFile(narrativePath)
		if err != nil {
			return AssemblyTrace{}, err
		}
		if _, err := s.LoadNarrative(narrativePath); err != nil {
			return AssemblyTrace{}, err
		}
		narrativeHash = hashBytes(b)
	}
	parent := filepath.Dir(out)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return AssemblyTrace{}, err
	}
	stage, err := os.MkdirTemp(parent, "."+filepath.Base(out)+".stage-")
	if err != nil {
		return AssemblyTrace{}, err
	}
	defer os.RemoveAll(stage)
	if err := os.Mkdir(filepath.Join(stage, "slides"), 0755); err != nil {
		return AssemblyTrace{}, err
	}
	if err := os.Mkdir(filepath.Join(stage, "inputs"), 0755); err != nil {
		return AssemblyTrace{}, err
	}
	trace := AssemblyTrace{Schema: AssemblySchema, ConfigSHA256: hashBytes(cb), NarrativePath: config.NarrativePath, NarrativeSHA256: narrativeHash, FitStatus: "unmeasured_changed_copy", RequiredNext: []string{"pptxcompose probe", "native measurement", "fit-report", "build", "verify", "visual review"}}
	var deck map[string]any
	var baseFields []byte
	var slides []any
	for i, item := range config.Slides {
		valuesPath := resolveInput(base, item.ValuesPath)
		vb, err := os.ReadFile(valuesPath)
		if err != nil {
			return AssemblyTrace{}, err
		}
		var vf ValuesFile
		if err := readStrictJSON(vb, &vf); err != nil {
			return AssemblyTrace{}, fmt.Errorf("values %s: %w", item.ValuesPath, err)
		}
		if narrativePath != "" {
			if vf.NarrativePath != "" || vf.NarrativeSlideID != "" {
				return AssemblyTrace{}, fmt.Errorf("values %s declares narrative while assembly has narrative_path", item.ValuesPath)
			}
			vf.NarrativePath, vf.NarrativeSlideID = narrativePath, item.SlideID
		} else if vf.NarrativePath != "" {
			vf.NarrativePath = resolveInput(filepath.Dir(valuesPath), vf.NarrativePath)
		}
		seq := fmt.Sprintf("%03d", i+1)
		inputFile := filepath.Join(stage, "inputs", seq+".json")
		if _, err := writeJSON(inputFile, vf); err != nil {
			return AssemblyTrace{}, err
		}
		slideDir := filepath.Join(stage, "slides", seq+"-"+item.SlideID)
		selection, err := s.Instantiate(indexPath, item.ContractID, inputFile, slideDir, allowUnqualified)
		if err != nil {
			return AssemblyTrace{}, fmt.Errorf("slide %s: %w", item.SlideID, err)
		}
		ins, err := s.Inspect(indexPath, item.ContractID)
		if err != nil {
			return AssemblyTrace{}, err
		}
		b, err := os.ReadFile(filepath.Join(slideDir, "spec.json"))
		if err != nil {
			return AssemblyTrace{}, err
		}
		var part map[string]any
		if err := json.Unmarshal(b, &part); err != nil {
			return AssemblyTrace{}, err
		}
		pSlides, ok := part["slides"].([]any)
		if !ok || len(pSlides) != 1 {
			return AssemblyTrace{}, fmt.Errorf("slide %s instantiation did not emit one slide", item.SlideID)
		}
		pSlide, ok := pSlides[0].(map[string]any)
		if !ok {
			return AssemblyTrace{}, fmt.Errorf("slide %s output invalid", item.SlideID)
		}
		pSlide["id"] = item.SlideID
		if ptr := ins.Contract.Composition.PaginationBinding; ptr != "" {
			if err := pointerSet(pSlide, ptr, strconv.Itoa(i+1)); err != nil {
				return AssemblyTrace{}, fmt.Errorf("slide %s pagination: %w", item.SlideID, err)
			}
		}
		part["slides"] = []any{}
		fields, err := json.Marshal(part)
		if err != nil {
			return AssemblyTrace{}, err
		}
		if i == 0 {
			deck, baseFields = part, fields
		} else if !bytes.Equal(baseFields, fields) {
			return AssemblyTrace{}, fmt.Errorf("slide %s template deck fields differ", item.SlideID)
		}
		slides = append(slides, pSlide)
		part["slides"] = []any{pSlide}
		pBytes, err := writeJSON(filepath.Join(slideDir, "spec.json"), part)
		if err != nil {
			return AssemblyTrace{}, err
		}
		trace.Slides = append(trace.Slides, AssemblySlideTrace{Position: i + 1, SlideID: item.SlideID, ContractID: item.ContractID, ValuesPath: item.ValuesPath, ValuesSHA256: hashBytes(vb), SpecPath: filepath.ToSlash(filepath.Join("slides", seq+"-"+item.SlideID, "spec.json")), SpecSHA256: hashBytes(pBytes), TracePath: filepath.ToSlash(filepath.Join("slides", seq+"-"+item.SlideID, "selection-trace.json")), Selection: selection})
	}
	deck["slides"] = slides
	specBytes, err := writeJSON(filepath.Join(stage, "spec.json"), deck)
	if err != nil {
		return AssemblyTrace{}, err
	}
	trace.SpecSHA256 = hashBytes(specBytes)
	if _, err := writeJSON(filepath.Join(stage, "assembly-trace.json"), trace); err != nil {
		return AssemblyTrace{}, err
	}
	if _, err := os.Stat(out); err == nil {
		return AssemblyTrace{}, fmt.Errorf("output exists: %s", out)
	} else if !os.IsNotExist(err) {
		return AssemblyTrace{}, err
	}
	if err := os.Rename(stage, out); err != nil {
		return AssemblyTrace{}, err
	}
	return trace, nil
}
