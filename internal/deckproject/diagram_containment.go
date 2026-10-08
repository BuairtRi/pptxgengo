package deckproject

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type DiagramOverlap struct {
	Container string        `json:"container"`
	A         string        `json:"a"`
	B         string        `json:"b"`
	Bounds    wmdesign.Rect `json:"bounds_in_container_axes"`
}
type DiagramContainmentResult struct {
	Applied      bool                    `json:"applied"`
	BeforeSHA256 string                  `json:"before_sha256"`
	AfterSHA256  string                  `json:"after_sha256"`
	Actor        string                  `json:"actor"`
	Reason       string                  `json:"reason"`
	Members      []string                `json:"members"`
	Container    string                  `json:"container,omitempty"`
	Padding      wmdesign.DiagramPadding `json:"padding"`
	Decision     string                  `json:"decision,omitempty"`
	Inspection   DiagramInspection       `json:"inspection"`
}

func validateContainmentDeclarations(rules map[string]wmdesign.DiagramContainment) error {
	if len(rules) > 1000 {
		return fmt.Errorf("containment exceeds 1000 members")
	}
	for member, r := range rules {
		if member == "" || len(member) > 512 || r.Container == "" || len(r.Container) > 512 || member == r.Container {
			return fmt.Errorf("containment requires distinct bounded object names")
		}
		for _, p := range []float64{r.Padding.Top, r.Padding.Right, r.Padding.Bottom, r.Padding.Left} {
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1e6 {
				return fmt.Errorf("containment padding must be finite and nonnegative")
			}
		}
	}
	visited := map[string]uint8{}
	var visit func(string) error
	visit = func(name string) error {
		if visited[name] == 1 {
			return fmt.Errorf("logical containment cycle")
		}
		if visited[name] == 2 {
			return nil
		}
		visited[name] = 1
		if r, ok := rules[name]; ok {
			if e := visit(r.Container); e != nil {
				return e
			}
		}
		visited[name] = 2
		return nil
	}
	for member := range rules {
		if e := visit(member); e != nil {
			return e
		}
	}
	return nil
}
func inverseGeometryMatrix(m geometryMatrix) (geometryMatrix, error) {
	det := m[0]*m[3] - m[1]*m[2]
	if math.IsNaN(det) || math.IsInf(det, 0) || math.Abs(det) < 1e-12 {
		return geometryMatrix{}, fmt.Errorf("singular native containment coordinate space")
	}
	a, b, c, d := m[3]/det, -m[1]/det, -m[2]/det, m[0]/det
	return geometryMatrix{a, b, c, d, -a*m[4] - c*m[5], -b*m[4] - d*m[5]}, nil
}
func nativeOuterMatrix(objects map[string]*geometryObject, name string) (geometryMatrix, error) {
	o := objects[name]
	if o == nil {
		return geometryMatrix{}, fmt.Errorf("containment native object missing: %s", name)
	}
	m, e := nativeCoordinateMatrix(objects, o.geometry.Parent)
	if e != nil {
		return m, e
	}
	return m.mul(geometryOriented(o.geometry)), nil
}
func checkDiagramContainment(objects map[string]*geometryObject, rules map[string]wmdesign.DiagramContainment) ([]wmdesign.DiagramContainmentObservation, []DiagramOverlap, error) {
	if e := validateContainmentDeclarations(rules); e != nil {
		return nil, nil, e
	}
	members := []string{}
	for name := range rules {
		members = append(members, name)
	}
	sort.Strings(members)
	children := map[string][]string{}
	for name, o := range objects {
		children[o.geometry.Parent] = append(children[o.geometry.Parent], name)
	}
	observations := []wmdesign.DiagramContainmentObservation{}
	for _, name := range members {
		rule := rules[name]
		container := objects[rule.Container]
		member := objects[name]
		if container == nil || member == nil {
			return nil, nil, fmt.Errorf("containment %s -> %s references a missing object; revise the declared relationship", name, rule.Container)
		}
		g := container.geometry
		if g.W <= 0 || g.H <= 0 {
			return nil, nil, fmt.Errorf("containment container %s has no rectangular area", rule.Container)
		}
		// A member cannot contain its physical ancestor. Logical ownership is
		// independent of grouping, but cannot contradict that native hierarchy.
		for p := g.Parent; p != ""; {
			if p == name {
				return nil, nil, fmt.Errorf("containment contradicts native group ownership")
			}
			o := objects[p]
			if o == nil {
				return nil, nil, fmt.Errorf("native containment parent missing")
			}
			p = o.geometry.Parent
		}
		outer, e := nativeOuterMatrix(objects, rule.Container)
		if e != nil {
			return nil, nil, e
		}
		inv, e := inverseGeometryMatrix(outer)
		if e != nil {
			return nil, nil, e
		}
		x0, y0, x1, y1 := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
		var include func(string) error
		include = func(n string) error {
			o := objects[n]
			m, e := nativeOuterMatrix(objects, n)
			if e != nil {
				return e
			}
			r := inv.mul(m)
			b := r.bounds(nativeGeometryAllocation(o.geometry))
			x0 = math.Min(x0, b.X)
			y0 = math.Min(y0, b.Y)
			x1 = math.Max(x1, b.X+b.W)
			y1 = math.Max(y1, b.Y+b.H)
			for _, child := range children[n] {
				if e = include(child); e != nil {
					return e
				}
			}
			return nil
		}
		if e = include(name); e != nil {
			return nil, nil, e
		}
		p := rule.Padding
		clearance := wmdesign.DiagramPadding{Top: y0 - g.Y, Right: g.X + g.W - x1, Bottom: g.Y + g.H - y1, Left: x0 - g.X}
		if p.Top+p.Bottom >= g.H || p.Left+p.Right >= g.W {
			return nil, nil, fmt.Errorf("containment padding consumes container %s", rule.Container)
		}
		if clearance.Top < p.Top-.02 || clearance.Right < p.Right-.02 || clearance.Bottom < p.Bottom-.02 || clearance.Left < p.Left-.02 {
			return nil, nil, fmt.Errorf("containment %s exceeds %s padded allocation (clearance top=%.3f right=%.3f bottom=%.3f left=%.3f pt)", name, rule.Container, clearance.Top, clearance.Right, clearance.Bottom, clearance.Left)
		}
		observations = append(observations, wmdesign.DiagramContainmentObservation{Member: name, Container: rule.Container, Padding: p, Clearance: clearance, ContainerRect: wmdesign.Rect{X: g.X, Y: g.Y, W: g.W, H: g.H}, MemberBounds: wmdesign.Rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}})
	}
	overlaps := []DiagramOverlap{}
	for i, a := range observations {
		for _, b := range observations[i+1:] {
			if a.Container != b.Container {
				continue
			}
			x0, y0 := math.Max(a.MemberBounds.X, b.MemberBounds.X), math.Max(a.MemberBounds.Y, b.MemberBounds.Y)
			x1, y1 := math.Min(a.MemberBounds.X+a.MemberBounds.W, b.MemberBounds.X+b.MemberBounds.W), math.Min(a.MemberBounds.Y+a.MemberBounds.H, b.MemberBounds.Y+b.MemberBounds.H)
			if x1-x0 > .02 && y1-y0 > .02 {
				overlaps = append(overlaps, DiagramOverlap{a.Container, a.Member, b.Member, wmdesign.Rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}})
			}
		}
	}
	return observations, overlaps, nil
}

func ContainDiagram(p *Project, slideID string, members []string, container string, padding wmdesign.DiagramPadding, actor, reason, bundle, engine string, apply bool) (DiagramContainmentResult, error) {
	out := DiagramContainmentResult{BeforeSHA256: p.SourceHash(), Members: members, Container: container, Padding: padding, Actor: actor, Reason: reason}
	if strings.TrimSpace(actor) == "" || strings.TrimSpace(reason) == "" || len(actor) > 256 || len(reason) > 4096 || len(members) == 0 || len(members) > 1000 {
		return out, fmt.Errorf("containment requires actor, reason and 1..1000 member names")
	}
	idx, _, e := diagramSlide(p, slideID)
	if e != nil {
		return out, e
	}
	s := p.Document.Slides[idx]
	rules := map[string]wmdesign.DiagramContainment{}
	for name, r := range s.DiagramContainment {
		rules[name] = r
	}
	seen := map[string]bool{}
	for _, name := range members {
		if name == "" || len(name) > 512 || seen[name] {
			return out, fmt.Errorf("invalid or duplicate containment member")
		}
		seen[name] = true
		if container == "" {
			if _, ok := rules[name]; !ok {
				return out, fmt.Errorf("member has no containment declaration: %s", name)
			}
			delete(rules, name)
		} else {
			rules[name] = wmdesign.DiagramContainment{Container: container, Padding: padding}
		}
	}
	if e = validateContainmentDeclarations(rules); e != nil {
		return out, e
	}
	main, e := sourceYAML(p.Raw)
	if e != nil {
		return out, e
	}
	slides, documents, e := authoredSlides(p, main)
	if e != nil {
		return out, e
	}
	node := slides[slideID]
	if len(rules) > 0 {
		v, e := editYAMLNode(rules)
		if e != nil {
			return out, e
		}
		preserveDiagramComments(mappingNode(node, "diagram_containment"), v)
		replaceMappingField(node, "diagram_containment", v)
		pin, _ := editYAMLNode(s.Template)
		replaceMappingField(node, "native_geometry_template", pin)
	} else {
		removeMappingField(node, "diagram_containment")
		if len(s.NativeGeometry) == 0 && len(s.NativeOrder) == 0 {
			removeMappingField(node, "native_geometry_template")
		}
	}
	file := p.SlideFiles[slideID]
	if file == "" {
		file = filepath.Base(p.SourcePath)
	}
	raw, e := encodeSourceYAML(documents[file])
	if e != nil {
		return out, e
	}
	changes := map[string][]byte{file: raw}
	candidate, e := loadProject(p.SourcePath, mergeTextOverrides(p.SourceFiles, changes))
	if e != nil {
		return out, e
	}
	out.AfterSHA256 = candidate.SourceHash()
	out.Inspection, e = InspectDiagram(candidate, slideID, bundle, engine)
	if e != nil {
		return out, e
	}
	if !apply {
		return out, nil
	}
	out.Applied = true
	out.Decision = "decisions/containment-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	changes[out.Decision] = canonical(out)
	_, e = commitSourceChanges(p, changes, func(c *Project) error { _, e := InspectDiagram(c, slideID, bundle, engine); return e })
	return out, e
}
