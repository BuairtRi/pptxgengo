package deckproject

import (
	"fmt"
	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"math"
)

// This is a fixed source-frame allowance, never a caller supplied maximum.
// Caption and footer clearance takes precedence over its six-point limit.
func sourceContainerZone(f wmdesign.ResolvedFrame) wmdesign.Rect {
	bottom := f.Body.Y + f.Body.H + 6
	for _, r := range []wmdesign.Rect{f.Source, f.FooterRow, f.FooterBand} {
		if r.W > 0 && r.H > 0 {
			bottom = math.Min(bottom, r.Y-6)
		}
	}
	if f.FooterRule > 0 {
		bottom = math.Min(bottom, f.FooterRule-6)
	}
	return wmdesign.Rect{X: f.Body.X, Y: 0, W: f.Body.W, H: bottom}
}

func chooseSourceSceneZone(kind string, args map[string]any, b wmdesign.Rect, f wmdesign.ResolvedFrame) (string, wmdesign.Rect, error) {
	zone, origin, err := chooseZone(b, f)
	if err == nil {
		return zone, origin, nil
	}
	if kind != "container" && kind != "frame" {
		return zone, origin, err
	}
	if _, ok := args[wmdesign.SceneSourceGeometryArgument]; !ok {
		return zone, origin, err
	}
	r := sourceContainerZone(f)
	if b.X >= r.X-.001 && b.Y >= r.Y-.001 && b.X+b.W <= r.X+r.W+.001 && b.Y+b.H <= r.Y+r.H+.001 {
		return "source_container", r, nil
	}
	return "", wmdesign.Rect{}, fmt.Errorf("source container exceeds fixed six-point allowance or six-point caption/footer clearance (body=%+v source=%+v footer=%+v rule=%v permitted=%+v)", f.Body, f.Source, f.FooterRow, f.FooterRule, r)
}

func validateSourceContainerLocal(t LocalTemplate, n Node) error {
	v := t.Provenance
	if n.Kind != "component" && n.Kind != "composite" || n.Definition == nil || (n.Definition.ID != "wmds/component/container" && n.Definition.ID != "wmds/component/frame") || v == nil || v.Parent.Scope != "shared" || v.SourceFile == "" || v.SourceFileSHA256 == "" || v.DefinitionSHA256 == "" || n.Arguments[wmdesign.SceneSourceGeometryArgument] == nil {
		return fmt.Errorf("source_container requires an original pinned shared container/frame and captured source geometry")
	}
	return nil
}

// Recreate the actual pinned parent, not a claimed ancestor snapshot. Matching
// node identity and captured geometry permits materialized text changes while
// refusing a newly invented component or forged source extent.
func validateSourceContainerOriginal(t LocalTemplate, n Node, original TemplateScaffold) error {
	if err := validateSourceContainerLocal(t, n); err != nil {
		return err
	}
	v, actual := t.Provenance, original.Template.Provenance
	if v.Parent.ID != actual.Parent.ID || v.Parent.Revision != actual.Parent.Revision || v.DefinitionSHA256 != actual.DefinitionSHA256 || v.SourceFile != actual.SourceFile || v.SourceFileSHA256 != actual.SourceFileSHA256 {
		return fmt.Errorf("source_container actual parent/hash provenance mismatch")
	}
	for _, originalNode := range original.Template.Nodes {
		if originalNode.ID != n.ID {
			continue
		}
		var originalGeometry, candidateGeometry wmdesign.SceneSourceGeometry
		if err := strictInto(originalNode.Arguments[wmdesign.SceneSourceGeometryArgument], &originalGeometry); err != nil {
			return err
		}
		if err := strictInto(n.Arguments[wmdesign.SceneSourceGeometryArgument], &candidateGeometry); err != nil {
			return err
		}
		if originalNode.Placement.Zone != "source_container" || originalNode.Definition.ID != n.Definition.ID || string(canonical(originalGeometry)) != string(canonical(candidateGeometry)) {
			return fmt.Errorf("source_container original node geometry/type mismatch")
		}
		return nil
	}
	return fmt.Errorf("source_container original node identity missing")
}
