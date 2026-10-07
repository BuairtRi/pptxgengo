package deckproject

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type pilotNativeRect struct{ X, Y, W, H int64 }
type pilotDiagramEvidence struct {
	Schema               string          `json:"schema"`
	Status               string          `json:"status"`
	Scope                string          `json:"scope"`
	Execution            string          `json:"execution"`
	HumanAcceptance      string          `json:"human_acceptance"`
	VerificationPlatform string          `json:"verification_platform"`
	VerifiedAt           string          `json:"verified_at"`
	BuildID              string          `json:"baseline_build_id"`
	ReceiptSHA256        string          `json:"baseline_receipt_sha256"`
	BaselineSHA256       string          `json:"baseline_sha256"`
	EditedSHA256         string          `json:"edited_sha256"`
	IdentityCount        int             `json:"identity_count"`
	BeforeInput          pilotNativeRect `json:"before_input_emu"`
	AfterInput           pilotNativeRect `json:"after_input_emu"`
	InputDeltaY          int64           `json:"input_delta_y_emu"`
	ConnectorEndpoints   string          `json:"connector_endpoints"`
}

func pilotDiagramGeometry(shape *xmlNode) (pilotNativeRect, *xmlNode, error) {
	r := pilotNativeRect{}
	if shape == nil {
		return r, nil, fmt.Errorf("missing native shape")
	}
	props := directXML(shape, lineagePML, "spPr")
	if props == nil {
		return r, nil, fmt.Errorf("missing native shape properties")
	}
	x := directXML(props, drawingML, "xfrm")
	if x == nil || (attr(x, "rot") != "" && attr(x, "rot") != "0") {
		return r, nil, fmt.Errorf("missing or rotated native transform")
	}
	off, ext := directXML(x, drawingML, "off"), directXML(x, drawingML, "ext")
	if off == nil || ext == nil {
		return r, nil, fmt.Errorf("missing native geometry")
	}
	vals := []*int64{&r.X, &r.Y, &r.W, &r.H}
	for i, a := range []string{attr(off, "x"), attr(off, "y"), attr(ext, "cx"), attr(ext, "cy")} {
		n, e := strconv.ParseInt(a, 10, 64)
		if e != nil || n < 0 || n > 1<<40 {
			return r, nil, fmt.Errorf("invalid bounded geometry")
		}
		*vals[i] = n
	}
	return r, x, nil
}
func pilotFlip(x *xmlNode, name string) (bool, error) {
	switch attr(x, name) {
	case "", "0", "false":
		return false, nil
	case "1", "true":
		return true, nil
	default:
		return false, fmt.Errorf("invalid native flip")
	}
}
func pilotDiagramTask(b *TextBaseline, edited []byte) (pilotDiagramEvidence, error) {
	out := pilotDiagramEvidence{Schema: "pptxgengo.native-diagram-task.v1", Status: "verified_supplied_copy", Scope: "one synthetic combined-shape diagram; plain copy, vertical move, attached straight endpoint geometry and Save As identities only", Execution: "supplied-file verification; application provenance is not independently verified", HumanAcceptance: "not_recorded", VerificationPlatform: runtime.GOOS + "/" + runtime.GOARCH, VerifiedAt: time.Now().UTC().Format(time.RFC3339), BuildID: b.Receipt.BuildID, ReceiptSHA256: b.ReceiptSHA256, BaselineSHA256: b.Receipt.Outputs["deck.pptx"], EditedSHA256: digest(edited)}
	actual, e := InspectNativeLineage(edited, b.Objects)
	if e != nil {
		return out, e
	}
	if len(actual.Issues) != 0 || len(actual.Objects) != 6 || len(b.Objects.Objects) != 6 {
		return out, fmt.Errorf("diagram identity inventory changed")
	}
	out.IdentityCount = len(actual.Objects)
	byToken := map[string]NativeLineageObject{}
	baseByToken := map[string]NativeLineageObject{}
	roles := map[string]string{}
	for _, o := range actual.Objects {
		byToken[o.ShapeToken] = o
	}
	for _, o := range b.inspection.Objects {
		baseByToken[o.ShapeToken] = o
	}
	for _, o := range b.Objects.Objects {
		if o.SlideID != "native-diagram" {
			return out, fmt.Errorf("expected authored diagram fixture")
		}
		q := byToken[o.ShapeToken]
		if o.NodeID == "input" || o.NodeID == "output" || o.NodeID == "input-output" {
			roles[o.NodeID] = o.ShapeToken
		}
		want := o.NativeText
		if o.NodeID == "input" {
			if want != "Editable input" {
				return out, fmt.Errorf("wrong original input")
			}
			want = "Reviewed input"
		}
		if nativeParagraphText(nativeParagraphs(q.shape)) != want {
			return out, fmt.Errorf("unexpected native text for authored node %s", o.NodeID)
		}
	}
	if len(roles) != 3 {
		return out, fmt.Errorf("missing unique authored pilot roles")
	}
	input, output, edge := byToken[roles["input"]], byToken[roles["output"]], byToken[roles["input-output"]]
	if input.Kind != "sp" || output.Kind != "sp" || edge.Kind != "cxnSp" || input.ParentToken != "" || output.ParentToken != "" || edge.ParentToken != "" {
		return out, fmt.Errorf("combined top-level native kinds changed")
	}
	for _, o := range []NativeLineageObject{input, output} {
		props := directXML(o.shape, lineagePML, "spPr")
		if props == nil || directXML(props, drawingML, "solidFill") == nil {
			return out, fmt.Errorf("node lost combined text and fill")
		}
		geom := directXML(props, drawingML, "prstGeom")
		if geom == nil || attr(geom, "prst") != "rect" || directXML(props, drawingML, "custGeom") != nil {
			return out, fmt.Errorf("combined rectangle geometry changed")
		}
	}
	before, _, e := pilotDiagramGeometry(baseByToken[roles["input"]].shape)
	if e != nil {
		return out, e
	}
	after, xf, e := pilotDiagramGeometry(input.shape)
	if e != nil {
		return out, e
	}
	for _, name := range []string{"flipH", "flipV"} {
		v, e := pilotFlip(xf, name)
		if e != nil || v {
			return out, fmt.Errorf("input rectangle flipped")
		}
	}
	if after.W == 0 || after.H == 0 || after.X != before.X || after.W != before.W || after.H != before.H || after.Y <= before.Y {
		return out, fmt.Errorf("expected positive vertical move without rectangle resize")
	}
	out.BeforeInput, out.AfterInput, out.InputDeltaY = before, after, after.Y-before.Y
	beforeOutput, _, e := pilotDiagramGeometry(baseByToken[roles["output"]].shape)
	if e != nil {
		return out, e
	}
	target, xf, e := pilotDiagramGeometry(output.shape)
	if e != nil {
		return out, e
	}
	for _, name := range []string{"flipH", "flipV"} {
		v, e := pilotFlip(xf, name)
		if e != nil || v {
			return out, fmt.Errorf("output rectangle flipped")
		}
	}
	if target != beforeOutput {
		return out, fmt.Errorf("unrequested output geometry change")
	}
	nv := directXML(edge.shape, lineagePML, "nvCxnSpPr")
	if nv == nil {
		return out, fmt.Errorf("missing connector nonvisual properties")
	}
	c := directXML(nv, lineagePML, "cNvCxnSpPr")
	if c == nil {
		return out, fmt.Errorf("missing connector endpoint container")
	}
	begin, end := directXML(c, drawingML, "stCxn"), directXML(c, drawingML, "endCxn")
	if begin == nil || end == nil || attr(begin, "id") != input.NativeID || attr(begin, "idx") != "3" || attr(end, "id") != output.NativeID || attr(end, "idx") != "1" {
		return out, fmt.Errorf("connector no longer bound to exact tagged nodes/sites")
	}
	edgeProps := directXML(edge.shape, lineagePML, "spPr")
	if edgeProps == nil {
		return out, fmt.Errorf("missing connector paint")
	}
	edgeGeom := directXML(edgeProps, drawingML, "prstGeom")
	if edgeGeom == nil || attr(edgeGeom, "prst") != "line" || directXML(edgeProps, drawingML, "custGeom") != nil {
		return out, fmt.Errorf("expected native straight connector")
	}
	er, ex, e := pilotDiagramGeometry(edge.shape)
	if e != nil {
		return out, e
	}
	fh, e := pilotFlip(ex, "flipH")
	if e != nil {
		return out, e
	}
	fv, e := pilotFlip(ex, "flipV")
	if e != nil {
		return out, e
	}
	bx, by, tx, ty := er.X, er.Y, er.X+er.W, er.Y+er.H
	if fh {
		bx, tx = tx, bx
	}
	if fv {
		by, ty = ty, by
	}
	near := func(a, b int64) bool { d := a - b; return d >= -1 && d <= 1 }
	if !near(bx, after.X+after.W) || !near(by, after.Y+after.H/2) || !near(tx, target.X) || !near(ty, target.Y+target.H/2) {
		return out, fmt.Errorf("connector geometry does not follow moved endpoint sites")
	}
	out.ConnectorEndpoints = "verified exact tagged nodes; right-center to left-center; endpoint rounding tolerance one EMU"
	return out, nil
}

// Read-only fixture qualification works from receipt-pinned bytes. It does not
// migrate the recorded compiler or approve source/geometry adoption.
func TestNativeEditabilityDiagramSuppliedCopy(t *testing.T) {
	fixture, edited, out := os.Getenv("PPTXGENGO_NATIVE_DIAGRAM_FIXTURE"), os.Getenv("PPTXGENGO_NATIVE_DIAGRAM_EDITED"), os.Getenv("PPTXGENGO_NATIVE_DIAGRAM_VERIFY_OUT")
	if fixture == "" && edited == "" && out == "" {
		t.Skip("explicit supplied diagram fixture/copy/new evidence directory required")
	}
	if fixture == "" || edited == "" || out == "" {
		t.Fatal("all three diagram qualification paths required")
	}
	p, e := Load(fixture)
	if e != nil {
		t.Fatal(e)
	}
	b, e := ReadTextBaseline(p, "", "")
	if e != nil {
		t.Fatal(e)
	}
	data, e := readReconcileFile(edited, lineageMaxPackage)
	if e != nil {
		t.Fatal(e)
	}
	report, e := pilotDiagramTask(b, data)
	if e != nil {
		t.Fatal(e)
	}
	again, e := ReadTextBaseline(p, "", "")
	if e != nil || again.ReceiptSHA256 != b.ReceiptSHA256 {
		t.Fatal("baseline changed during verification", e)
	}
	if e = os.Mkdir(out, 0700); e != nil {
		t.Fatal("evidence destination must be new", e)
	}
	for name, raw := range map[string][]byte{"evidence.json": canonical(report), "edited.pptx": data} {
		if e = writeExclusive(filepath.Join(out, name), raw, 0444); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("read-only supplied diagram evidence: %s", out)
}

func TestNativeEditabilityDiagramTaskGeometry(t *testing.T) {
	_, b := nativeEditingDiagramFixture(t)
	edited := reconcileEditFields(t, b, map[string]string{"input": "Reviewed input"})
	part := ""
	for _, o := range b.Objects.Objects {
		if o.NodeID == "input" {
			part = o.NativePart
		}
	}
	changed := lineageEdit(t, edited, part, func(raw []byte) []byte {
		spans, e := lineageSpans(raw)
		if e != nil {
			t.Fatal(e)
		}
		patches := []lineagePatch{}
		var walk func(*lineageSpan)
		walk = func(s *lineageSpan) {
			if s.node.Name.Space == lineagePML && (s.node.Name.Local == "sp" || s.node.Name.Local == "cxnSp") {
				name := ""
				for _, nv := range s.children {
					for _, id := range nv.children {
						if id.node.Name.Space == lineagePML && id.node.Name.Local == "cNvPr" {
							name = attr(id.node, "name")
						}
					}
				}
				v := raw[s.start:s.end]
				replacement := v
				if name == "input" {
					replacement = bytes.Replace(v, []byte(`y="2514600"`), []byte(`y="3200400"`), 1)
				}
				if name == "input-output" {
					replacement = bytes.Replace(v, []byte(`<a:xfrm>`), []byte(`<a:xfrm flipV="1">`), 1)
					replacement = bytes.Replace(replacement, []byte(`cy="0"`), []byte(`cy="685800"`), 1)
				}
				if !bytes.Equal(v, replacement) {
					patches = append(patches, lineagePatch{s.start, s.end, string(replacement)})
				}
			}
			for _, child := range s.children {
				walk(child)
			}
		}
		walk(spans)
		if len(patches) != 2 {
			t.Fatal("expected two geometry fixture patches")
		}
		data, e := lineageApply(raw, patches)
		if e != nil {
			t.Fatal(e)
		}
		return data
	})
	r, e := pilotDiagramTask(b, changed)
	if e != nil || r.InputDeltaY != 685800 {
		t.Fatal(r, e)
	}
	for name, data := range map[string][]byte{"original": b.files["deck.pptx"], "text_only": edited, "wrong_copy": reconcileEditFields(t, b, map[string]string{"input": "Other input"}), "wrong_connector": lineageEdit(t, changed, part, func(raw []byte) []byte { return bytes.Replace(raw, []byte(`idx="3"`), []byte(`idx="0"`), 1) }), "wrong_reroute": lineageEdit(t, changed, part, func(raw []byte) []byte { return bytes.Replace(raw, []byte(`cy="685800"`), []byte(`cy="685000"`), 1) })} {
		t.Run(name, func(t *testing.T) {
			if _, e := pilotDiagramTask(b, data); e == nil {
				t.Fatal("invalid supplied task passed")
			}
		})
	}
}
