package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaturityFeedbackNativeMarkerCentering(t *testing.T) {
	for _, revision := range []string{LibraryRevisionV3, IntakeTeamCurveMonotoneRevision} {
		r := intakeTestRenderer(t)
		r.source.Revision = revision
		for _, number := range []string{"1", "12", "4.5"} {
			for _, active := range []bool{false, true} {
				p := &scenePlan{}
				if err := r.maturityMarker(p, "marker", curvePoint{200, 300}, number, active); err != nil {
					t.Fatal(err)
				}
				tr := p.Items[1].Text
				if tr.Layout.Original != number || tr.Layout.Style.Weight != 600 || tr.Layout.Style.TrackingPt != 0 {
					t.Fatal("marker copy/type changed")
				}
				if math.Abs(tr.Rect.Y+tr.Rect.H/2-300) > 1e-9 {
					t.Fatal("marker box off center")
				}
				if revision == IntakeTeamCurveMonotoneRevision {
					if tr.Rect.H != 24 || tr.VerticalAlign != "middle" {
						t.Fatal("native centering missing")
					}
				} else if tr.VerticalAlign != "" || tr.Rect.H == 24 {
					t.Fatal("v3 marker changed")
				}
			}
		}
	}
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeTeamCurveMonotoneRevision
	raw := json.RawMessage(`{"type":"maturity","x":100,"y":126,"w":600,"h":300,"stages":[{"label":"Start","n":"1"},{"label":"Next","n":"12"}],"branch":{"from":0,"label":"Beyond","n":"4.5"}}`)
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{{ID: "markers", Frame: FrameRequest{NoHeader: true}, Nodes: []Node{{ID: "m", Kind: "scene", Scene: &SceneSpec{Node: raw}}}}}}
	data, _, err := buildWithLoadedSource(bundle, source, doc, CandidateEngine, nil)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		rd, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(rd)
		rd.Close()
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, shape := range strings.Split(string(b), "<p:sp>")[1:] {
			shape = strings.Split(shape, "</p:sp>")[0]
			if !strings.Contains(shape, ".number\"") {
				continue
			}
			count++
			if !strings.Contains(shape, `anchor="ctr"`) || !strings.Contains(shape, `cy="304800"`) {
				t.Fatal("native marker box/alignment missing")
			}
		}
		if count != 3 {
			t.Fatalf("native markers %d", count)
		}
		return
	}
	t.Fatal("slide XML missing")
}

func TestMaturityFeedbackFrozenFrameClearance(t *testing.T) {
	root := filepath.Join("..", "..", "planning", "wm-design-contracts", "v4", "intake-20261003-frozen", "source")
	entries := intakeRepairEntries(t, filepath.Join(root, "templates", "library"))
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v3")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	source.Revision = IntakeTeamCurveMonotoneRevision
	b, err := os.ReadFile(filepath.Join(root, "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &source.Frames); err != nil {
		t.Fatal(err)
	}
	count := 0
	for key, entry := range entries {
		slide := intakeRepairSlide(t, entry)
		hasMaturity := false
		for _, n := range slide.Nodes {
			if n.Scene != nil {
				var tag struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(n.Scene.Node, &tag); err != nil {
					t.Fatal(err)
				}
				if tag.Type == "maturity" {
					hasMaturity = true
				}
			}
		}
		if !hasMaturity {
			continue
		}
		count++
		t.Run(key, func(t *testing.T) {
			if err := ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide); err != nil {
				t.Fatal(err)
			}
			_, report, err := buildWithLoadedSource(bundle, source, Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, Slides: []SlideSpec{slide}}, CandidateEngine, nil)
			if err != nil {
				t.Fatal(err)
			}
			s := report.Slides[0]
			for _, tr := range s.Texts {
				if !strings.Contains(tr.ID, ".stages.") || !strings.HasSuffix(tr.ID, ".label") {
					continue
				}
				top := s.Frame.Body.Y
				if s.Frame.Request.Split != "" {
					if tr.Rect.X >= s.Frame.TallBody.X && tr.Rect.X+tr.Rect.W <= s.Frame.TallBody.X+s.Frame.TallBody.W+.02 {
						top = s.Frame.TallBody.Y
					} else {
						top = s.Frame.ShortBody.Y
					}
				}
				if tr.Rect.Y < top-.02 {
					t.Fatalf("stage %s at %.3f above body %.3f", tr.ID, tr.Rect.Y, top)
				}
			}
		})
	}
	if count < 15 {
		t.Fatalf("only %d maturity specimens checked", count)
	}
}
