package wmdesign

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ComposeSceneNode lowers a local component's frame-relative allocation into
// the same strict source-node planners used by shared templates. Nested argument
// geometry remains explicit. Source fields are validated by the selected planner.
func ComposeSceneNode(kind string, arguments map[string]any, allocation Rect) (json.RawMessage, error) {
	noHeight := map[string]bool{}
	for _, k := range strings.Fields("bullets ol list schedule grouplabel numhead colhead strongnum pullquote logo art square table metric feesummary matrix beforeafter stepper vstepper phasehead phases timeaxis gauge gantt swimlane legend pod person orgchart governance device dotmap") {
		noHeight[k] = true
	}
	known := map[string]bool{}
	for _, k := range strings.Fields("text textblock bullets ol list schedule grouplabel numhead colhead strongnum pullquote imageframe logo art square mark thumbnail table chart card cardrow metric callout feesummary block frame chevron textarrow connector container cylinder node layerrow matrix beforeafter stepper vstepper phasehead phases timeaxis pyramid funnel cycle road gauge bracket scorelegend gantt swimlane legend pod role person orgchart governance logoslot device plane dotmap teamcurve venn maturity") {
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
	if kind != "connector" {
		if kind == "maturity" {
			// The source curve has2.5pt stroked endpoints. A local allocation
			// owns the complete ink envelope, so reserve their half-width.
			if allocation.W <= 3 || allocation.H <= 3 {
				return nil, fmt.Errorf("scene.local_maturity_allocation_too_small")
			}
			allocation = Rect{allocation.X + 1.5, allocation.Y + 1.5, allocation.W - 3, allocation.H - 3}
		}
		if kind == "teamcurve" {
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
