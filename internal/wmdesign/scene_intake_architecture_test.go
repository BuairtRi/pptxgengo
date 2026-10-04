package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"path/filepath"
	"strings"
	"testing"
)

func intakeTestRenderer(t *testing.T) *renderer {
	t.Helper()
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, e := Load(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	// Synthetic intake planner fixtures retain the baseline revision they were
	// written against. Assets and typography come from the sole current bundle;
	// current-revision integration tests explicitly replace source below.
	source.Revision = LibraryRevisionV2
	typography, e := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	return &renderer{source: source, typeEngine: typography, bundle: bundle}
}
func intakeArchitecturePlan(t *testing.T, r *renderer, node any) *scenePlan {
	t.Helper()
	raw, e := json.Marshal(node)
	if e != nil {
		t.Fatal(e)
	}
	p, ok, e := r.planIntakeArchitectureScene("sample", raw, SceneContext{Surface: "light", Path: "/body/0"})
	if e != nil || !ok {
		t.Fatalf("plan recognized=%t error=%v", ok, e)
	}
	return p
}
func TestIntakeArchitecturePlaceholderAndPlane(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, n := range []map[string]any{
		{"type": "logoslot", "x": 117, "y": 351, "w": 90, "h": 36, "name": "Application Load Balancer"},
		{"type": "logoslot", "x": 651, "y": 198, "w": 60, "h": 30, "name": "Cloud Storage"},
		{"type": "plane", "x": 237, "y": 144, "w": 342, "h": 90, "surface": "inverse", "label": "User interface layer", "text": "React and TypeScript"},
		{"type": "plane", "x": 237, "y": 288, "w": 342, "h": 90, "surface": "outline", "label": "Data access layer", "text": "PostgreSQL access"},
	} {
		p := intakeArchitecturePlan(t, r, n)
		if len(p.Groups) != 1 || p.Groups[0].Contract != IntakeArchitectureContract {
			t.Fatalf("missing native group: %+v", p.Groups)
		}
		for _, item := range p.Items {
			if item.Shape != nil {
				for _, point := range item.Shape.Props.Points {
					if point.Close {
						continue
					}
					if point.X.Val < 0 || point.Y.Val < 0 {
						t.Fatal("unclipped hatch")
					}
				}
			}
		}
	}
}
func TestIntakeArchitectureLogoContain(t *testing.T) {
	r := intakeTestRenderer(t)
	im := image.NewNRGBA(image.Rect(0, 0, 200, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 200; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: 20, G: 60, B: 180, A: 128})
		}
	}
	var data bytes.Buffer
	if e := png.Encode(&data, im); e != nil {
		t.Fatal(e)
	}
	r.projectAssets = map[string]AssetData{"project:logo": {Data: data.Bytes(), SHA256: fmt.Sprintf("%x", sha256.Sum256(data.Bytes())), MIME: "image/png"}}
	p := intakeArchitecturePlan(t, r, map[string]any{"type": "logoslot", "x": 20, "y": 30, "w": 100, "h": 60, "src": "project:logo", "name": "Client", "caption": "Alliance"})
	var found bool
	for _, item := range p.Items {
		if item.Image != nil {
			found = true
			im := item.Image
			if im.Sizing != nil {
				t.Fatal("contain image carries crop")
			}
			if math.Abs(im.W.Val*72-88) > 1e-8 || math.Abs(im.H.Val*72-22) > 1e-8 || math.Abs(im.X.Val*72-26) > 1e-8 || math.Abs(im.Y.Val*72-49) > 1e-8 {
				t.Fatalf("bad padded contain placement %+v", im.PositionProps)
			}
		}
	}
	if !found || p.Bounds.H <= 60 {
		t.Fatal("missing contained image or caption allocation")
	}
}
func TestIntakeArchitectureStrictValidation(t *testing.T) {
	r := intakeTestRenderer(t)
	for _, raw := range []string{
		`{"type":"logoslot","x":0,"y":0,"w":10,"h":10,"src":"logo-pos"}`,
		`{"type":"plane","x":0,"y":0,"w":90,"h":0,"label":"A","text":"B"}`,
		`{"type":"device","x":0,"y":0,"w":20,"label":"A"}`,
		`{"type":"logoslot","x":0,"y":0,"w":90,"h":36,"unexpected":1}`,
		`{"type":"device","x":0,"y":0,"w":108}`,
	} {
		_, ok, e := r.planIntakeArchitectureScene("sample", json.RawMessage(raw), SceneContext{Surface: "light"})
		if !ok || e == nil {
			t.Fatalf("accepted invalid %s", raw)
		}
	}
	_, ok, e := r.planIntakeArchitectureScene("sample", json.RawMessage(`{"type":"other"}`), SceneContext{})
	if ok || e != nil {
		t.Fatal("unexpected dispatch ownership")
	}
}
func TestIntakeArchitectureHatchClipping(t *testing.T) {
	for lo := 0.; lo < 400; lo += 6 * math.Sqrt2 {
		for _, point := range intakeClipBand(150, 60, lo, lo+6*math.Sqrt2) {
			if point[0] < -1e-7 || point[1] < -1e-7 || point[0] > 150+1e-7 || point[1] > 60+1e-7 {
				t.Fatalf("hatch point outside %+v", point)
			}
		}
	}
	r := intakeTestRenderer(t)
	_, _, e := r.planIntakeArchitectureScene("sample", json.RawMessage(`{"type":"plane","x":0,"y":0,"w":100,"h":10,"label":"Long label","text":"Content that cannot fit"}`), SceneContext{})
	if e == nil || !strings.Contains(e.Error(), "overflow") {
		t.Fatalf("missing overflow: %v", e)
	}
}

func TestIntakeArchitectureDeviceNativeGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("renderer integration requires registered private branding assets; run make test-integration")
	}
	r := intakeTestRenderer(t)
	p := intakeArchitecturePlan(t, r, map[string]any{"type": "device", "x": 75, "y": 162, "w": 108, "icon": "monitor", "label": "Web app", "sub": "Browser access"})
	if len(p.Groups) != 2 || p.Groups[len(p.Groups)-1].Contract != IntakeArchitectureContract {
		t.Fatalf("device native ownership lost: %+v", p.Groups)
	}
	images, texts := 0, 0
	for _, item := range p.Items {
		if item.Image != nil {
			images++
		}
		if item.Text != nil {
			texts++
		}
	}
	if images != 1 || texts != 2 || p.Bounds.H <= 36 {
		t.Fatalf("device stack incorrect: %d images/%d texts/%g height", images, texts, p.Bounds.H)
	}
}

func TestIntakeArchitectureProjectSVGContain(t *testing.T) {
	r := intakeTestRenderer(t)
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 50"><path d="M0 0H200V50H0Z" fill="#070154"/></svg>`)
	r.projectAssets = map[string]AssetData{"project:vector-logo": {Data: data, SHA256: fmt.Sprintf("%x", sha256.Sum256(data)), MIME: "image/svg+xml"}}
	p := intakeArchitecturePlan(t, r, map[string]any{"type": "logoslot", "x": 20, "y": 30, "w": 100, "h": 60, "src": "project:vector-logo", "name": "Client"})
	if len(p.Items) != 1 || p.Items[0].Image == nil || p.Items[0].Image.SVGFallbackData == "" || p.Items[0].Image.Sizing != nil {
		t.Fatal("missing uncropped vector logo plus compatibility fallback")
	}
}

func TestIntakeArchitecturePlaneUsesDiamondEnvelope(t *testing.T) {
	p := &scenePlan{Items: []sceneItem{{Text: &TextRecord{ID: "plane.label", Rect: Rect{20, 1, 60, 20}, Layout: TextLayout{Style: Style{Leading: 12}, Lines: []TextLine{{Advance: 50}}, EstimatedOccupiedHeight: 10}}}}}
	if e := intakePlaneTextEnvelope(p, Rect{0, 0, 100, 90}); e == nil {
		t.Fatal("text inside rectangular box but outside diamond was accepted")
	}
}

func TestIntakeArchitectureAndGeographyAllFrozenV3Nodes(t *testing.T) {
	if testing.Short() {
		t.Skip("exhaustive frozen slide builds require registered private branding assets; run make test-integration")
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, e := Load(bundle, "")
	if e != nil {
		t.Fatal(e)
	}
	typography, e := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if e != nil {
		t.Fatal(e)
	}
	r := &renderer{source: source, typeEngine: typography, bundle: bundle}
	counts := map[string]int{}
	for template, raw := range source.Templates {
		var node any
		if e = json.Unmarshal(raw, &node); e != nil {
			t.Fatal(e)
		}
		var walk func(any, string)
		walk = func(value any, path string) {
			switch v := value.(type) {
			case map[string]any:
				kind, _ := v["type"].(string)
				if kind == "logoslot" || kind == "device" || kind == "plane" || kind == "dotmap" {
					encoded, e := json.Marshal(v)
					if e != nil {
						t.Fatal(e)
					}
					var p *scenePlan
					var ok bool
					ctx := SceneContext{Surface: "light", Path: path}
					if kind == "dotmap" {
						p, ok, e = r.planIntakeGeographyScene("node", encoded, ctx)
					} else {
						p, ok, e = r.planIntakeArchitectureScene("node", encoded, ctx)
					}
					if e != nil || !ok {
						t.Errorf("%s%s %s: %v", template, path, kind, e)
						return
					}
					if e = sceneTextEnvelope(p, ctx); e != nil {
						t.Errorf("%s%s: %v", template, path, e)
					}
					counts[kind]++
				}
				for key, child := range v {
					walk(child, path+"/"+key)
				}
			case []any:
				for i, child := range v {
					walk(child, fmt.Sprintf("%s/%d", path, i))
				}
			}
		}
		walk(node, "")
	}
	for _, kind := range []string{"logoslot", "device", "plane", "dotmap"} {
		if counts[kind] == 0 {
			t.Fatalf("frozen source has no %s specimen", kind)
		}
	}
	t.Logf("all frozen v3 intake nodes rendered: %+v", counts)
}

func TestIntakeContainedImageRotationUsesActualPicture(t *testing.T) {
	r := intakeTestRenderer(t)
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewNRGBA(image.Rect(0, 0, 200, 50))); err != nil {
		t.Fatal(err)
	}
	r.projectAssets = map[string]AssetData{"project:photo": {Data: b.Bytes(), SHA256: fmt.Sprintf("%x", sha256.Sum256(b.Bytes())), MIME: "image/png"}}
	raw := json.RawMessage(`{"type":"imageframe","x":100,"y":100,"w":100,"h":100,"photo":"project:photo","fit":"contain","rotate":45}`)
	p, _, err := r.planMediaScene("photo", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(p.Bounds.W-88.3883476483) > 1e-5 || !inside(p.Bounds, Rect{100, 100, 100, 100}) {
		t.Fatalf("contain bounds: %+v", p.Bounds)
	}
}
