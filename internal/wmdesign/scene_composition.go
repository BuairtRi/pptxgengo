package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
)

const SceneSourceGeometryArgument = "_source_geometry"
const SceneSourceGeometrySchema = "pptxgengo.scene-source-geometry.v1"

// SceneSourceGeometry retains source anchors separately from measured native ink.
// Fractions refer to the measured allocation, so moving/resizing remains local.
type SceneSourceGeometry struct {
	Schema         string   `json:"schema"`
	X              float64  `json:"x_fraction"`
	Y              float64  `json:"y_fraction"`
	Width          *float64 `json:"width_fraction,omitempty"`
	Height         *float64 `json:"height_fraction,omitempty"`
	VennSetCount   *int     `json:"venn_set_count,omitempty"`
	VennAreaHeight *float64 `json:"venn_area_height_fraction,omitempty"`
}

// CaptureSceneSourceGeometry must run before stripping source x/y/w/h. Cardrow
// width denotes the whole row here, while the strict source renderer takes item w.
func CaptureSceneSourceGeometry(kind string, arguments map[string]any, measured Rect, sourceZone ...Rect) error {
	if kind == "connector" {
		return nil
	}
	if !sceneGeometryFinite(measured.X, measured.Y, measured.W, measured.H) || measured.W <= 0 || measured.H <= 0 {
		return fmt.Errorf("scene.source_geometry_positive_allocation_required")
	}
	if _, exists := arguments[SceneSourceGeometryArgument]; exists {
		return fmt.Errorf("scene.source_geometry_already_captured")
	}
	x, e := sceneGeometryNumber(arguments, "x", 0)
	if e != nil {
		return e
	}
	y, e := sceneGeometryNumber(arguments, "y", 0)
	if e != nil {
		return e
	}
	g := SceneSourceGeometry{Schema: SceneSourceGeometrySchema, X: (x - measured.X) / measured.W, Y: (y - measured.Y) / measured.H}
	if kind == "venn" {
		count, err := sceneGeometryItemCount(arguments["sets"])
		if err != nil || count < 2 || count > 4 {
			return fmt.Errorf("scene.source_geometry_invalid_venn_set_count")
		}
		g.VennSetCount = &count
		if len(sourceZone) > 1 {
			return fmt.Errorf("scene.source_geometry_single_source_zone_required")
		}
		if len(sourceZone) == 1 && count > 2 {
			height, err := sceneGeometryNumber(arguments, "h", 0)
			width, errW := sceneGeometryNumber(arguments, "w", 0)
			zone := sourceZone[0]
			if err != nil || errW != nil || height <= 0 || !sceneGeometryFinite(zone.X, zone.Y, zone.W, zone.H) || zone.W <= 0 || zone.H <= 0 {
				return fmt.Errorf("scene.source_geometry_invalid_venn_source_zone")
			}
			if zone.Y == y && zone.W >= width && zone.W <= width+36 {
				height = math.Max(height, zone.Y+zone.H-y)
			}
			fraction := height / measured.H
			g.VennAreaHeight = &fraction
		}
	}
	widthField := "w"
	if kind == "square" {
		widthField = "size"
	}
	if _, exists := arguments[widthField]; exists {
		width, e := sceneGeometryNumber(arguments, widthField, 0)
		if e != nil || width <= 0 {
			return fmt.Errorf("scene.source_geometry_positive_source_width_required")
		}
		if kind == "cardrow" {
			count, e := sceneGeometryItemCount(arguments["items"])
			if e != nil {
				return e
			}
			gap, e := sceneGeometryNumber(arguments, "gap", 0)
			if e != nil || gap < 0 {
				return fmt.Errorf("scene.source_geometry_invalid_cardrow_gap")
			}
			width = width*float64(count) + gap*float64(count-1)
		}
		fraction := width / measured.W
		g.Width = &fraction
	}
	if _, exists := arguments["h"]; exists {
		height, e := sceneGeometryNumber(arguments, "h", 0)
		if e != nil || height <= 0 {
			return fmt.Errorf("scene.source_geometry_positive_source_height_required")
		}
		fraction := height / measured.H
		g.Height = &fraction
	}
	arguments[SceneSourceGeometryArgument] = g
	return nil
}

func sceneGeometryFinite(numbers ...float64) bool {
	for _, number := range numbers {
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return false
		}
	}
	return true
}
func sceneGeometryNumber(arguments map[string]any, key string, fallback float64) (float64, error) {
	value, exists := arguments[key]
	if !exists {
		return fallback, nil
	}
	raw, e := json.Marshal(value)
	if e != nil {
		return 0, e
	}
	var number float64
	if e = json.Unmarshal(raw, &number); e != nil || !sceneGeometryFinite(number) {
		return 0, fmt.Errorf("scene.source_geometry_invalid_number: %s", key)
	}
	return number, nil
}
func sceneGeometryItemCount(value any) (int, error) {
	raw, e := json.Marshal(value)
	if e != nil {
		return 0, e
	}
	var items []json.RawMessage
	if e = json.Unmarshal(raw, &items); e != nil || len(items) == 0 {
		return 0, fmt.Errorf("scene.source_geometry_cardrow_items_required")
	}
	return len(items), nil
}
func sceneGeometryDecode(value any) (SceneSourceGeometry, error) {
	var g SceneSourceGeometry
	raw, e := json.Marshal(value)
	if e != nil {
		return g, e
	}
	var fields map[string]json.RawMessage
	if e = json.Unmarshal(raw, &fields); e != nil {
		return g, e
	}
	for _, required := range []string{"x_fraction", "y_fraction"} {
		value, exists := fields[required]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return g, fmt.Errorf("scene.source_geometry_requires_numeric_origin")
		}
	}
	for _, optional := range []string{"width_fraction", "height_fraction", "venn_set_count", "venn_area_height_fraction"} {
		if value, exists := fields[optional]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return g, fmt.Errorf("scene.source_geometry_null_extent")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if e = decoder.Decode(&g); e != nil {
		return g, fmt.Errorf("scene.source_geometry_invalid_metadata: %w", e)
	}
	var extra any
	if e = decoder.Decode(&extra); e != io.EOF {
		return g, fmt.Errorf("scene.source_geometry_requires_single_value")
	}
	if g.Schema != SceneSourceGeometrySchema || !sceneGeometryFinite(g.X, g.Y) || g.Width != nil && (!sceneGeometryFinite(*g.Width) || *g.Width <= 0) || g.Height != nil && (!sceneGeometryFinite(*g.Height) || *g.Height <= 0) {
		return g, fmt.Errorf("scene.source_geometry_invalid_metadata")
	}
	if g.VennSetCount != nil && (*g.VennSetCount < 2 || *g.VennSetCount > 4) {
		return g, fmt.Errorf("scene.source_geometry_invalid_venn_set_count")
	}
	if g.VennAreaHeight != nil && (g.VennSetCount == nil || *g.VennSetCount < 3 || !sceneGeometryFinite(*g.VennAreaHeight) || *g.VennAreaHeight <= 0) {
		return g, fmt.Errorf("scene.source_geometry_invalid_venn_area_height")
	}
	return g, nil
}

// ComposeSceneNode lowers a local component's frame-relative allocation into
// the same strict source-node planners used by shared templates. Nested argument
// geometry remains explicit. Source fields are validated by the selected planner.
func ComposeSceneNode(kind string, arguments map[string]any, allocation Rect) (json.RawMessage, error) {
	noHeight := map[string]bool{}
	for _, k := range strings.Fields("rule editable-table bullets ol list schedule grouplabel numhead colhead strongnum pullquote logo art square table metric feesummary matrix beforeafter stepper vstepper phasehead phases timeaxis gauge gantt swimlane legend pod person orgchart governance device dotmap") {
		noHeight[k] = true
	}
	known := map[string]bool{}
	for _, k := range strings.Fields("rule portfolio commercial process assessment editable-table editable-list editable-card editable-block attached-connector text textblock bullets ol list schedule grouplabel numhead colhead strongnum pullquote imageframe logo art square mark thumbnail table chart card cardrow metric callout feesummary block frame chevron textarrow connector container cylinder node layerrow matrix beforeafter stepper vstepper phasehead phases timeaxis pyramid funnel cycle road roadfork gauge bracket scorelegend gantt swimlane legend pod role person orgchart governance logoslot device plane dotmap teamcurve venn maturity") {
		known[k] = true
	}
	if !known[kind] {
		return nil, fmt.Errorf("scene.unknown_local_definition: %s", kind)
	}
	n := make(map[string]any, len(arguments)+5)
	for k, v := range arguments {
		n[k] = v
	}
	n["type"] = kind
	sourceAllocation := allocation
	sourceGeometry := false
	if metadata, exists := n[SceneSourceGeometryArgument]; exists {
		if kind == "connector" {
			return nil, fmt.Errorf("scene.connector_uses_allocation_points")
		}
		g, e := sceneGeometryDecode(metadata)
		if e != nil {
			return nil, e
		}
		if !sceneGeometryFinite(allocation.X, allocation.Y, allocation.W, allocation.H) || allocation.W <= 0 || allocation.H <= 0 {
			return nil, fmt.Errorf("scene.source_geometry_positive_allocation_required")
		}
		sourceAllocation.X = allocation.X + g.X*allocation.W
		sourceAllocation.Y = allocation.Y + g.Y*allocation.H
		if g.Width != nil {
			sourceAllocation.W = allocation.W * (*g.Width)
		}
		if g.Height != nil {
			sourceAllocation.H = allocation.H * (*g.Height)
		}
		if g.VennSetCount != nil {
			if kind != "venn" {
				return nil, fmt.Errorf("scene.source_geometry_venn_count_wrong_kind")
			}
			count, err := sceneGeometryItemCount(n["sets"])
			if err != nil {
				return nil, err
			}
			if count != *g.VennSetCount {
				// The captured source envelope belongs to its original topology.
				// A different count reflows into the owned allocation, rather than
				// applying the old source-to-ink ratio to different circle geometry.
				sourceAllocation = allocation
				if count == 2 {
					if allocation.W <= 1 || allocation.H <= 1 {
						return nil, fmt.Errorf("scene.local_venn_allocation_too_small")
					}
					sourceAllocation = Rect{allocation.X + .5, allocation.Y + .5, allocation.W - 1, allocation.H - 1}
				}
			} else if g.VennAreaHeight != nil {
				n["_composition_area_h"] = allocation.H * *g.VennAreaHeight
			}
		}
		if kind == "cardrow" {
			count, e := sceneGeometryItemCount(n["items"])
			if e != nil {
				return nil, e
			}
			gap, e := sceneGeometryNumber(n, "gap", 0)
			if e != nil || gap < 0 {
				return nil, fmt.Errorf("scene.source_geometry_invalid_cardrow_gap")
			}
			sourceAllocation.W = (sourceAllocation.W - gap*float64(count-1)) / float64(count)
			if sourceAllocation.W <= 0 {
				return nil, fmt.Errorf("scene.source_geometry_cardrow_gap_consumes_allocation")
			}
		}
		sourceGeometry = true
		delete(n, SceneSourceGeometryArgument)
	}
	if space, exists := n["point_space"]; exists {
		if kind != "connector" || space != "allocation" {
			return nil, fmt.Errorf("point_space allocation is only supported for connector")
		}
		var points [][2]float64
		raw, err := json.Marshal(n["points"])
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &points); err != nil || len(points) < 2 {
			return nil, fmt.Errorf("connector allocation points require at least two coordinate pairs")
		}
		for i := range points {
			points[i][0] = allocation.X + points[i][0]*allocation.W
			points[i][1] = allocation.Y + points[i][1]*allocation.H
		}
		n["points"] = points
		delete(n, "point_space")
	}
	if kind != "connector" {
		allocation = sourceAllocation
		if kind == "cardrow" && !sourceGeometry {
			count, e := sceneGeometryItemCount(n["items"])
			if e != nil {
				return nil, e
			}
			gap, e := sceneGeometryNumber(n, "gap", 0)
			if e != nil || gap < 0 {
				return nil, fmt.Errorf("scene.source_geometry_invalid_cardrow_gap")
			}
			allocation.W = (allocation.W - gap*float64(count-1)) / float64(count)
			if allocation.W <= 0 {
				return nil, fmt.Errorf("scene.source_geometry_cardrow_gap_consumes_allocation")
			}
		}
		if kind == "venn" && !sourceGeometry {
			count, e := sceneGeometryItemCount(n["sets"])
			if e != nil {
				return nil, e
			}
			if count == 2 {
				// Two sets can use the full source height; their 1pt ellipse outline
				// extends half a point beyond it. Captured originals own this ink already.
				if allocation.W <= 1 || allocation.H <= 1 {
					return nil, fmt.Errorf("scene.local_venn_allocation_too_small")
				}
				allocation = Rect{allocation.X + .5, allocation.Y + .5, allocation.W - 1, allocation.H - 1}
			}
		}
		if kind == "maturity" && !sourceGeometry {
			// The source curve has2.5pt stroked endpoints. A local allocation
			// owns the complete ink envelope, so reserve their half-width.
			if allocation.W <= 3 || allocation.H <= 3 {
				return nil, fmt.Errorf("scene.local_maturity_allocation_too_small")
			}
			allocation = Rect{allocation.X + 1.5, allocation.Y + 1.5, allocation.W - 3, allocation.H - 3}
		}
		if kind == "phases" && !sourceGeometry {
			if current, marked := n["current"]; marked && current != nil {
				if allocation.H <= 24 {
					return nil, fmt.Errorf("scene.local_phases_allocation_too_small")
				}
				allocation.Y += 24
				allocation.H -= 24
			}
		}
		if kind == "teamcurve" && !sourceGeometry {
			// Native line series use a 2.5 pt stroke. Local placement owns the
			// complete ink envelope, including the stroked curve endpoints.
			if allocation.W <= 2.5 || allocation.H <= 2.5 {
				return nil, fmt.Errorf("scene.local_teamcurve_allocation_too_small")
			}
			allocation = Rect{allocation.X + 1.25, allocation.Y + 1.25, allocation.W - 2.5, allocation.H - 2.5}
		}
		n["x"], n["y"], n["w"] = allocation.X, allocation.Y, allocation.W
		if !noHeight[kind] {
			n["h"] = allocation.H
		}
		if kind == "square" {
			delete(n, "w")
			n["size"] = allocation.W
		}
	}
	return json.Marshal(n)
}

// SceneSourcePlanningZone captures the original frame context for a source
// component before its measured ink allocation becomes its local ownership.
func SceneSourcePlanningZone(arguments map[string]any, frame ResolvedFrame) (Rect, error) {
	x, err := sceneGeometryNumber(arguments, "x", 0)
	if err != nil {
		return Rect{}, err
	}
	w, err := sceneGeometryNumber(arguments, "w", 0)
	if err != nil {
		return Rect{}, err
	}
	zone := Rect{frame.Body.X, 0, frame.Body.W, frame.Body.Y + frame.Body.H}
	if frame.Rail.W > 0 && x >= frame.Rail.X-.02 && x < frame.Rail.X+frame.Rail.W {
		zone = Rect{frame.Rail.X, 0, frame.Rail.W, frame.Rail.Y + frame.Rail.H}
	}
	if frame.Request.Split != "" {
		if x >= frame.TallBody.X-.02 && x+w <= frame.TallBody.X+frame.TallBody.W+.02 {
			zone = frame.TallBody
		} else {
			zone = frame.ShortBody
		}
	}
	return zone, nil
}
