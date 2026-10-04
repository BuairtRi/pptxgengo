package wmdesign

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Audit every scene even when an earlier scene rejects its whole slide. This
// exposes capability gaps hidden behind a first layout failure. It is not a
// substitute for full-slide frame and native presentation checks.
func TestIntakeManualNodeSmoke(t *testing.T) {
	out := os.Getenv("WMDS_INTAKE_NODE_OUT")
	if out == "" {
		t.Skip("manual node diagnostic only")
	}
	if err := os.Mkdir(out, 0755); err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("WMDS_INTAKE_SNAPSHOT")
	if root == "" {
		t.Fatal("snapshot required")
	}
	entries := intakeRepairEntries(t, filepath.Join(root, "source", "templates", "library"))
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	baseline := intakeRepairEntries(t, filepath.Join(bundle, "source", "templates", "library"))
	base := intakeTestRenderer(t)
	base.source.Revision = IntakeTeamCurveMonotoneRevision
	var err error
	pinned := os.Getenv("WMDS_INTAKE_NODE_BUNDLE")
	if pinned != "" {
		base.source, err = Load(pinned, "")
		if err != nil {
			t.Fatal(err)
		}
		base.bundle = pinned
		base.typeEngine, err = NewTypographyEngine(filepath.Join(pinned, "fonts"), CandidateEngine)
		if err != nil {
			t.Fatal(err)
		}
	}
	frames, err := os.ReadFile(filepath.Join(root, "source", "frames", "v0", "frames.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(frames, &base.source.Frames); err != nil {
		t.Fatal(err)
	}
	type result struct {
		Template string `json:"template"`
		Node     string `json:"node"`
		Type     string `json:"type"`
		Error    string `json:"error,omitempty"`
	}
	var results []result
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := entries[key]
		slide := intakeRepairSlide(t, entry)
		if old, ok := baseline[key]; pinned == "" && ok {
			a, err := libraryObject(old.Slide)
			if err != nil {
				t.Fatal(err)
			}
			b, err := libraryObject(entry.Slide)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.DeepEqual(a, b) {
				if err = applyLibraryRefinements(key, LibraryRevisionV3, &slide); err != nil {
					t.Fatal(err)
				}
			}
		}
		if pinned != "" {
			err = applyLibraryRefinements(key, base.source.Revision, &slide)
		} else {
			err = ApplyIncomingIntakeRepairs(key, IntakeRepairRevision, &slide)
		}
		if err != nil {
			t.Fatal(err)
		}
		frame, err := base.source.ResolveFrame(slide.Frame)
		if err != nil {
			t.Fatal(err)
		}
		r := *base
		if err = r.registerSceneTargets(slide, frame); err != nil {
			t.Fatal(err)
		}
		for _, node := range slide.Nodes {
			if node.Scene == nil {
				continue
			}
			var tag struct {
				Type string  `json:"type"`
				Kind string  `json:"kind"`
				X    float64 `json:"x"`
				Y    float64 `json:"y"`
				W    float64 `json:"w"`
			}
			if err = json.Unmarshal(node.Scene.Node, &tag); err != nil {
				t.Fatal(err)
			}
			// Use the same scene context selection as production render.go.
			// A synthetic full-width split body hides the source's actual
			// column height and produces diagnostics for a different layout.
			ctx := SceneContext{Surface: frame.Request.Surface, Zone: Rect{frame.Body.X, 0, frame.Body.W, frame.Body.Y + frame.Body.H}, Path: node.Scene.Path, Keys: node.Scene.Keys, Notes: node.Scene.Notes}
			if (r.source.Revision == IntakeTeamCurveMonotoneRevision || isExpandedLibrary(r.source.Revision)) && tag.Type == "maturity" {
				ctx.Zone = frame.Body
			}
			if frame.Rail.W > 0 && tag.X >= frame.Rail.X-.02 && tag.X < frame.Rail.X+frame.Rail.W {
				ctx.Zone = Rect{frame.Rail.X, 0, frame.Rail.W, frame.Rail.Y + frame.Rail.H}
				ctx.Surface = frame.Request.RailSurface
			}
			if frame.Request.Split != "" {
				if tag.Type == "connector" {
					ctx.Zone = frame.Body
				} else if tag.X >= frame.TallBody.X-.02 && tag.X+tag.W <= frame.TallBody.X+frame.TallBody.W+.02 {
					ctx.Zone = frame.TallBody
				} else {
					ctx.Zone = frame.ShortBody
				}
				if isModernLibrary(r.source.Revision) && tag.Type == "chart" && tag.Kind == "quadrant" && tag.Y == 27 && ctx.Zone == frame.TallBody {
					ctx.Zone.Y -= 9
					ctx.Zone.H += 9
				}
			}
			switch tag.Type {
			case "imageframe", "square", "logo", "art", "mark", "thumbnail":
				ctx.Zone = Rect{0, 0, 960, 540}
			}
			if node.Scene.Allocation != nil {
				ctx.Zone = *node.Scene.Allocation
			}
			_, err = r.planSceneNode(node.ID, node.Scene.Node, ctx)
			item := result{Template: key, Node: node.ID, Type: tag.Type}
			if err != nil {
				item.Error = err.Error()
			}
			results = append(results, item)
		}
	}
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "node-results.json"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	failed, unsupported := 0, 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		}
		if strings.Contains(r.Error, "scene.unsupported_node") {
			unsupported++
		}
	}
	t.Log(fmt.Sprintf("Audited %d scene nodes; %d rejected; %d unsupported types. Output %s", len(results), failed, unsupported, out))
}
