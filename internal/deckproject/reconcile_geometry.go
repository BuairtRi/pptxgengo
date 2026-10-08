package deckproject

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

type GeometryReconciliationField struct {
	SourceGeometrySHA256 string          `json:"source_geometry_sha256,omitempty"`
	ID                   string          `json:"id"`
	SlideID              string          `json:"slide_id"`
	Name                 string          `json:"source_object"`
	ShapeToken           string          `json:"shape_token,omitempty"`
	Property             string          `json:"property"`
	Baseline             *NativeGeometry `json:"baseline,omitempty"`
	CurrentYAML          *NativeGeometry `json:"current_yaml,omitempty"`
	EditedNative         *NativeGeometry `json:"edited_native,omitempty"`
	BaselineOrder        []string        `json:"baseline_order,omitempty"`
	CurrentOrder         []string        `json:"current_order,omitempty"`
	EditedOrder          []string        `json:"edited_order,omitempty"`
	Status               string          `json:"status"`
	Reason               string          `json:"reason,omitempty"`
}

func threeWayStatus(a, b, c any) string {
	base, yaml, native := canonical(a), canonical(b), canonical(c)
	switch {
	case bytes.Equal(base, yaml) && bytes.Equal(base, native):
		return "no_op"
	case bytes.Equal(base, native):
		return "yaml_only"
	case bytes.Equal(yaml, native):
		return "matching_changes"
	case bytes.Equal(base, yaml):
		return "native_only"
	default:
		return "conflict"
	}
}
func geometryOnlyChange(a, b *xmlNode) bool {
	var clone func(*xmlNode) *xmlNode
	clone = func(n *xmlNode) *xmlNode {
		c := *n
		c.Children = nil
		for _, x := range n.Children {
			if x.Name.Local == "xfrm" && (x.Name.Space == drawingML || x.Name.Space == lineagePML) {
				continue
			}
			c.Children = append(c.Children, clone(x))
		}
		return &c
	}
	return reconcileStructureHash(clone(a)) == reconcileStructureHash(clone(b))
}

// Geometry proposals use exact tagged identity. Current YAML is rendered with
// the pinned compiler to detect simultaneous placement and template changes.
// Native coordinates are retained in their parent space; only final validation
// resolves group transforms into frame-space bounds.
func addGeometryReconciliation(p *Project, b *TextBaseline, edited []byte, report *TextReconciliationReport, bundle, engine string) error {
	c, e := Check(p, bundle, engine)
	if e != nil {
		return e
	}
	current, layout, e := wmdesign.BuildWithEngineAndAssets(bundle, "", c.Document, engine, c.Assets)
	if e != nil {
		return e
	}
	sourcePackage, e := openLineagePackage(current)
	if e != nil {
		return e
	}
	sourceBasis := map[string]map[string]string{}
	for i, s := range c.Document.Slides {
		raw, e := sourcePackage.read(fmt.Sprintf("ppt/slides/slide%d.xml", i+1))
		if e != nil {
			return e
		}
		inv, _, e := geometryInventory(raw)
		if e != nil {
			return e
		}
		sourceBasis[s.ID] = map[string]string{}
		for name := range inv {
			sourceBasis[s.ID][name] = nativeGeometryBasis(inv, name)
		}
	}
	current, e = applyNativeGeometry(current, c.Document, &layout)
	if e != nil {
		return e
	}
	currentPkg, e := openLineagePackage(current)
	if e != nil {
		return e
	}
	native, e := InspectNativeLineage(edited, b.Objects)
	if e != nil {
		return e
	}
	originals := map[string]NativeLineageObject{}
	matched := map[string]NativeLineageObject{}
	bindings := map[string]ObjectRecord{}
	blocked := map[string]bool{}
	blockedSlides := map[string]bool{}
	for _, i := range native.Issues {
		if i.ShapeToken != "" {
			blocked[i.ShapeToken] = true
		}
		if len(i.Kind) >= 6 && i.Kind[:6] == "slide_" {
			blockedSlides[i.SlideToken] = true
		}
	}
	for _, o := range b.inspection.Objects {
		originals[o.ShapeToken] = o
	}
	for _, o := range native.Objects {
		if !blocked[o.ShapeToken] && !blockedSlides[o.SlideToken] {
			matched[o.ShapeToken] = o
		}
	}
	for _, o := range b.Objects.Objects {
		bindings[o.ShapeToken] = o
	}
	inventories := map[string]map[string]*geometryObject{}
	orders := map[string]map[string][]string{}
	for i, s := range c.Document.Slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		raw, e := currentPkg.read(part)
		if e != nil {
			return e
		}
		inv, order, e := geometryInventory(raw)
		if e != nil {
			return e
		}
		inventories[s.ID], orders[s.ID] = inv, order
	}
	removedIssues := map[string]bool{}
	for _, o := range b.Objects.Objects {
		original := originals[o.ShapeToken]
		n, ok := matched[o.ShapeToken]
		if !ok || o.NodeID == "" || strings.HasPrefix(o.NodeID, "native-") {
			continue
		}
		parent := ""
		if original.ParentToken != "" {
			parent = bindings[original.ParentToken].NativeName
		}
		base, e := readNativeGeometry(original.shape, parent)
		if e != nil {
			continue
		}
		editedParent := ""
		if n.ParentToken != "" {
			editedParent = bindings[n.ParentToken].NativeName
		}
		after, e := readNativeGeometry(n.shape, editedParent)
		if e != nil {
			continue
		}
		f := GeometryReconciliationField{SourceGeometrySHA256: sourceBasis[o.SlideID][o.NativeName], SlideID: o.SlideID, Name: o.NativeName, ShapeToken: o.ShapeToken, Property: "transform", Baseline: &base, EditedNative: &after, Status: "manual_review"}
		if now := inventories[o.SlideID][o.NativeName]; now != nil && now.geometry.Kind == base.Kind && now.geometry.Parent == base.Parent && editedParent == parent {
			g := now.geometry
			f.CurrentYAML = &g
			f.Status = threeWayStatus(base, g, after)
			if (f.Status == "native_only" || f.Status == "conflict") && base.Child != nil && !bytes.Equal(canonical(base.Child), canonical(g.Child)) {
				f.Status = "manual_review"
				f.Reason = "The authored group coordinate space changed; rebuild a baseline before adopting its native transform."
			}
		} else {
			f.Reason = "The source object or parent changed; geometry ownership must be reviewed."
		}
		f.ID = digest(canonical(f))
		report.Geometry = append(report.Geometry, f)
		if f.Status != "manual_review" && geometryOnlyChange(original.shape, n.shape) {
			removedIssues[o.ShapeToken] = true
		}
	}
	// Compare paint order by lineage token, ignoring edited names and numeric IDs.
	// Any addition, deletion or ambiguous identity prevents automatic order adoption.
	baseOrders := map[string]map[string][]string{}
	nativeOrders := map[string]map[string][]string{}
	validOrders := map[string]bool{}
	collect := func(view NativeLineageInspection, out map[string]map[string][]string) {
		for _, n := range view.Objects {
			o, ok := bindings[n.ShapeToken]
			if !ok || blocked[n.ShapeToken] || blockedSlides[n.SlideToken] {
				continue
			}
			parent := ""
			if n.ParentToken != "" {
				parent = bindings[n.ParentToken].NativeName
			}
			if out[o.SlideID] == nil {
				out[o.SlideID] = map[string][]string{}
			}
			out[o.SlideID][parent] = append(out[o.SlideID][parent], o.NativeName)
		}
	}
	collect(b.inspection, baseOrders)
	collect(native, nativeOrders)
	for slide, parents := range baseOrders {
		for parent, base := range parents {
			after := nativeOrders[slide][parent]
			now := orders[slide][parent]
			if !sameNameSet(base, after) || !sameNameSet(base, now) {
				continue
			}
			if bytes.Equal(canonical(base), canonical(after)) && bytes.Equal(canonical(base), canonical(now)) {
				continue
			}
			f := GeometryReconciliationField{SlideID: slide, Name: parent, Property: "paint_order", BaselineOrder: base, CurrentOrder: now, EditedOrder: after, Status: threeWayStatus(base, now, after)}
			f.ID = digest(canonical(f))
			report.Geometry = append(report.Geometry, f)
			validOrders[slide+"\x00"+parent] = true
		}
	}
	issues := report.ManualReview[:0]
	for _, i := range report.ManualReview {
		if i.Kind == "native_structure_or_format_changed" && removedIssues[i.ShapeToken] {
			continue
		}
		if i.Kind == "native_paint_order_changed" && validOrders[i.SlideID+"\x00"+strings.TrimPrefix(i.Detail, "Paint order changed under source parent: ")] {
			continue
		}
		issues = append(issues, i)
	}
	report.ManualReview = issues
	sort.Slice(report.Geometry, func(i, j int) bool { return report.Geometry[i].ID < report.Geometry[j].ID })
	report.GeometryScope = "tagged_transforms_and_existing_object_paint_order_v1"
	report.AdoptionScope = "reviewed_plain_text_and_native_transforms; topology_routes_styles_remain_explicit_review"
	for _, f := range report.Geometry {
		report.Counts["geometry_"+f.Status]++
	}
	report.Counts["manual_review_items"] = len(report.ManualReview)
	if len(canonical(report)) > 60<<20 {
		return fmt.Errorf("reconcile.report_context_too_large")
	}
	return nil
}
func sameNameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	s := map[string]int{}
	for _, n := range a {
		s[n]++
	}
	for _, n := range b {
		s[n]--
	}
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

// Detect an order edit even in legacy text-only proposals. Geometry proposals
// turn a supported order change into an independently reviewed field.
func reconcileNativeOrder(b *TextBaseline, native NativeLineageInspection, add func(string, string, string, string, string)) {
	names := map[string]ObjectRecord{}
	for _, o := range b.Objects.Objects {
		names[o.ShapeToken] = o
	}
	a, z := map[string][]string{}, map[string][]string{}
	collect := func(view NativeLineageInspection, out map[string][]string) {
		for _, o := range view.Objects {
			b, ok := names[o.ShapeToken]
			if !ok {
				continue
			}
			parent := ""
			if o.ParentToken != "" {
				parent = names[o.ParentToken].NativeName
			}
			key := b.SlideID + "\x00" + parent
			out[key] = append(out[key], o.ShapeToken)
		}
	}
	collect(b.inspection, a)
	collect(native, z)
	keys := []string{}
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if sameNameSet(a[k], z[k]) && !bytes.Equal(canonical(a[k]), canonical(z[k])) {
			var slide, parent string
			for _, o := range b.Objects.Objects {
				p := ""
				if o.NativeParentToken != "" {
					p = names[o.NativeParentToken].NativeName
				}
				if o.SlideID+"\x00"+p == k {
					slide, parent = o.SlideID, p
					break
				}
			}
			add("native_paint_order_changed", "", slide, "", "Paint order changed under source parent: "+parent)
		}
	}
}
