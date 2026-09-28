package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/buairtri/pptxgengo/internal/nativepkg"
)

// T045's five gauges are source freeforms: five curved cells and one pointer
// in each row. These are the source object IDs, ordered left to right.
var gaugeCells = [5][5]string{
	{"57", "60", "56", "58", "59"},
	{"17", "20", "16", "18", "19"},
	{"37", "40", "36", "38", "39"},
	{"47", "50", "46", "48", "49"},
	{"27", "30", "26", "28", "29"},
}
var gaugePointers = [5]string{"61", "21", "41", "51", "31"}

const gaugeSceneSHA = "7b7c0f3b05df39af2e1dafac9c83a4798f55350f935b084861f4b58025a8cb05"
const gaugeSourceSHA = "9180d0d747500358ecb427dbe7ac771fa0f83de0e55c298ce7ceeb2482334fda"
const gaugeSceneRel = "slides/uhg-045.json"

type gaugeControl struct {
	HighlightCells []int  `json:"highlight_cells"`
	PointerCell    int    `json:"pointer_cell,omitempty"`
	HighlightColor string `json:"highlight_color,omitempty"`
}
type gaugeValues struct {
	Schema string         `json:"schema"`
	Gauges []gaugeControl `json:"gauges"`
}

func gaugeReadValues(path string) (gaugeValues, error) {
	var v gaugeValues
	b, err := os.ReadFile(path)
	if err != nil {
		return v, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&v); err != nil {
		return v, fmt.Errorf("gauge values: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return v, fmt.Errorf("gauge values contain trailing JSON")
	}
	if v.Schema != "pptxgengo.source-gauges.t045.v1" || len(v.Gauges) != 5 {
		return v, fmt.Errorf("gauge values require schema pptxgengo.source-gauges.t045.v1 and exactly five gauges")
	}
	for i := range v.Gauges {
		g := &v.Gauges[i]
		if len(g.HighlightCells) < 1 || len(g.HighlightCells) > 5 {
			return v, fmt.Errorf("gauge %d requires 1..5 highlighted cells", i+1)
		}
		seen := map[int]bool{}
		for _, n := range g.HighlightCells {
			if n < 1 || n > 5 || seen[n] {
				return v, fmt.Errorf("gauge %d highlighted cells must be unique integers 1..5", i+1)
			}
			seen[n] = true
		}
		if g.PointerCell == 0 {
			if len(g.HighlightCells) != 1 {
				return v, fmt.Errorf("gauge %d needs pointer_cell when multiple cells are highlighted", i+1)
			}
			g.PointerCell = g.HighlightCells[0]
		}
		if g.PointerCell < 1 || g.PointerCell > 5 {
			return v, fmt.Errorf("gauge %d pointer_cell must be 1..5", i+1)
		}
		if g.HighlightColor == "" {
			g.HighlightColor = "pink"
		}
		switch g.HighlightColor {
		case "gray", "navy", "blue", "pink":
		default:
			return v, fmt.Errorf("gauge %d highlight_color must be gray, navy, blue, or pink", i+1)
		}
	}
	return v, nil
}

func gaugeObject(scene *nativepkg.Node, id string) (*nativepkg.Node, error) {
	var found *nativepkg.Node
	scene.Walk(func(n *nativepkg.Node) {
		if accentNodeID(n) == id {
			if found == nil {
				found = n
			} else {
				found = &nativepkg.Node{Name: "duplicate"}
			}
		}
	})
	if found == nil || found.Name != "p:sp" {
		return nil, fmt.Errorf("T045 source shape %s is absent, duplicated, or not an editable shape", id)
	}
	return found, nil
}

func gaugeBinding(s *nativepkg.Slide, id string) (*nativepkg.Binding, error) {
	for i := range s.Bindings {
		if s.Bindings[i].BindingID == id {
			return &s.Bindings[i], nil
		}
	}
	return nil, fmt.Errorf("missing T045 gauge binding %s", id)
}

func gaugeSentinel(n *nativepkg.Node, attr string) (string, error) {
	v := n.Attr(attr)
	if !strings.HasPrefix(v, "__BINDING:") {
		return "", fmt.Errorf("T045 gauge attribute %s is not source bound", attr)
	}
	return strings.TrimPrefix(v, "__BINDING:"), nil
}

func gaugeSetBinding(s *nativepkg.Slide, n *nativepkg.Node, attr, value string) error {
	id, err := gaugeSentinel(n, attr)
	if err != nil {
		return err
	}
	b, err := gaugeBinding(s, id)
	if err != nil {
		return err
	}
	if b.Attribute != attr {
		return fmt.Errorf("T045 gauge binding %s does not address %s", id, attr)
	}
	b.Value = value
	return nil
}

func gaugeCellFill(s *nativepkg.Slide, id, color string) error {
	shape, err := gaugeObject(s.Scene, id)
	if err != nil {
		return err
	}
	sp := shape.Child("spPr")
	if sp == nil || sp.Child("custGeom") == nil || sp.Child("solidFill") == nil {
		return fmt.Errorf("T045 cell %s lost its source freeform/fill", id)
	}
	fill := sp.Child("solidFill")
	clr := fill.Child("schemeClr")
	if clr == nil || len(fill.Children) != 1 {
		return fmt.Errorf("T045 cell %s has an unsupported fill structure", id)
	}
	scheme := map[string]string{"inactive": "accent3", "navy": "dk1", "pink": "accent1"}
	if color == "blue" || color == "gray" {
		// Retain the source freeform and change only this fill node to an
		// explicit RGB value. Selected gray differs from inactive source gray.
		rgb := "0047FF"
		if color == "gray" {
			rgb = "7F7F7F"
		}
		if err := gaugeSetBinding(s, clr, "val", rgb); err != nil {
			return err
		}
		clr.Name = "a:srgbClr"
		return nil
	}
	return gaugeSetBinding(s, clr, "val", scheme[color])
}

func gaugeTransform(s *nativepkg.Slide, source *nativepkg.Slide, row, cell int) error {
	target, err := gaugeObject(s.Scene, gaugePointers[row])
	if err != nil {
		return err
	}
	x := target.Child("spPr").Child("xfrm")
	if x == nil || target.Child("spPr").Child("custGeom") == nil {
		return fmt.Errorf("T045 pointer %s lost source freeform/transform", gaugePointers[row])
	}
	// Each source row demonstrates one of the five discrete pointer targets.
	// Copy the source exemplar's local transform, never interpolate a new gauge.
	exemplarID := gaugePointers[cell-1]
	exemplar, err := gaugeObject(source.Scene, exemplarID)
	if err != nil {
		return err
	}
	ex := exemplar.Child("spPr").Child("xfrm")
	for _, part := range []struct {
		node  *nativepkg.Node
		attrs []string
	}{
		{x, []string{"rot"}},
		{x.Child("off"), []string{"x", "y"}},
		{x.Child("ext"), []string{"cx", "cy"}},
	} {
		if part.node == nil {
			return fmt.Errorf("T045 pointer %s has incomplete transform", gaugePointers[row])
		}
		var from *nativepkg.Node
		switch part.attrs[0] {
		case "rot":
			from = ex
		case "x":
			from = ex.Child("off")
		case "cx":
			from = ex.Child("ext")
		}
		if from == nil {
			return fmt.Errorf("T045 pointer exemplar %s has incomplete transform", exemplarID)
		}
		for _, attr := range part.attrs {
			id, err := gaugeSentinel(from, attr)
			if err != nil {
				return err
			}
			b, err := gaugeBinding(source, id)
			if err != nil {
				return err
			}
			if err := gaugeSetBinding(s, part.node, attr, b.Value); err != nil {
				return err
			}
		}
	}
	flip := "0"
	if ex.Attr("flipH") != "" {
		id, err := gaugeSentinel(ex, "flipH")
		if err != nil {
			return err
		}
		b, err := gaugeBinding(source, id)
		if err != nil {
			return err
		}
		flip = b.Value
	}
	if x.Attr("flipH") != "" {
		if err := gaugeSetBinding(s, x, "flipH", flip); err != nil {
			return err
		}
	} else if flip != "0" {
		x.SetAttr("flipH", flip)
	}
	return nil
}

func gaugeValidateSource(applied, source *nativepkg.Slide) error {
	if applied.Number != 45 || applied.Part != "ppt/slides/slide45.xml" ||
		applied.Number != source.Number || applied.Part != source.Part ||
		!reflect.DeepEqual(applied.Scene, source.Scene) || len(applied.Bindings) != len(source.Bindings) {
		return fmt.Errorf("gauge input must retain the pinned T045 scene structure")
	}
	for i := range source.Bindings {
		a, b := applied.Bindings[i], source.Bindings[i]
		if a.Value != b.Value && a.Property != "text" && a.Property != "fill.srgbClr.val" && a.Property != "fill.schemeClr.val" {
			return fmt.Errorf("T045 source geometry/resource binding %s changed before gauge application", b.BindingID)
		}
		a.Value, b.Value = "", ""
		if !reflect.DeepEqual(a, b) {
			return fmt.Errorf("T045 binding structure differs at %d", i)
		}
	}
	return nil
}

func applyGauge(args []string) error {
	f := flag.NewFlagSet("apply-gauge", flag.ContinueOnError)
	project := f.String("project", "", "source or component-applied scene project")
	reference := f.String("reference", "", "pinned original T045 source scene JSON")
	valuesPath := f.String("values", "", "five gauge controls JSON")
	out := f.String("out", "", "new scene project directory")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 || *project == "" || *reference == "" || *valuesPath == "" || *out == "" {
		return fmt.Errorf("usage: pptxtemplate apply-gauge --project PROJECT --reference PINNED_T045_SCENE --values GAUGES.json --out NEW_PROJECT")
	}
	v, err := gaugeReadValues(*valuesPath)
	if err != nil {
		return err
	}
	refHash, err := hashFile(*reference)
	if err != nil {
		return err
	}
	if refHash != gaugeSceneSHA {
		return fmt.Errorf("T045 reference source scene hash mismatch")
	}
	var source nativepkg.Slide
	if err := read(*reference, &source); err != nil {
		return err
	}
	input, err := filepath.Abs(*project)
	if err != nil {
		return err
	}
	input, err = filepath.EvalSymlinks(input)
	if err != nil {
		return err
	}
	var manifest nativepkg.Manifest
	if err := read(filepath.Join(input, "manifest.json"), &manifest); err != nil {
		return err
	}
	if manifest.Schema != "pptxgengo.native-scene.v2" || manifest.SourceSHA256 != gaugeSourceSHA {
		return fmt.Errorf("gauge project is not the pinned T045 source project")
	}
	var applied nativepkg.Slide
	if err := read(filepath.Join(input, gaugeSceneRel), &applied); err != nil {
		return err
	}
	if err := gaugeValidateSource(&applied, &source); err != nil {
		return err
	}
	for row, g := range v.Gauges {
		active := map[int]bool{}
		for _, cell := range g.HighlightCells {
			active[cell] = true
		}
		for cell, id := range gaugeCells[row] {
			color := "inactive"
			if active[cell+1] {
				color = g.HighlightColor
			}
			if err := gaugeCellFill(&applied, id, color); err != nil {
				return fmt.Errorf("gauge %d cell %d: %w", row+1, cell+1, err)
			}
		}
		if err := gaugeTransform(&applied, &source, row, g.PointerCell); err != nil {
			return fmt.Errorf("gauge %d pointer: %w", row+1, err)
		}
	}
	// Check every source placeholder still binds after the narrowly scoped edits.
	encoded, err := json.Marshal(applied)
	if err != nil {
		return err
	}
	var resolved nativepkg.Slide
	if err := json.Unmarshal(encoded, &resolved); err != nil {
		return err
	}
	if err := nativepkg.ApplyBindings(resolved.Scene, resolved.Bindings); err != nil {
		return err
	}
	dest, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	dest, err = realDestination(dest)
	if err != nil {
		return err
	}
	if dest == input || strings.HasPrefix(dest, input+string(os.PathSeparator)) {
		return fmt.Errorf("output must be outside input project")
	}
	if _, err := os.Lstat(dest); !os.IsNotExist(err) {
		return fmt.Errorf("output must be a new directory")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Dir(dest), ".gauge-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := copyTree(input, stage); err != nil {
		return err
	}
	if err := write(filepath.Join(stage, gaugeSceneRel), applied); err != nil {
		return err
	}
	inputHash, _ := hashFile(filepath.Join(input, gaugeSceneRel))
	outputHash, _ := hashFile(filepath.Join(stage, gaugeSceneRel))
	valuesBytes, _ := os.ReadFile(*valuesPath)
	valuesHash := sha256.Sum256(valuesBytes)
	report := map[string]any{"schema": "pptxgengo.source-gauge-application.v1", "template_id": "t045-graphics-and-layouts-045", "reference_scene_sha256": gaugeSceneSHA, "input_scene_sha256": inputHash, "output_scene_sha256": outputHash, "values_sha256": hex.EncodeToString(valuesHash[:]), "gauges": v.Gauges, "source_geometry": "retained five-cell freeform gauges and five source pointer exemplars", "qualification": "requires native open, text fit and visual review"}
	if err := write(filepath.Join(stage, "gauge-application.json"), report); err != nil {
		return err
	}
	if err := os.Rename(stage, dest); err != nil {
		return err
	}
	fmt.Println(dest)
	return nil
}
