package wmdesign

import (
	"encoding/json"
	"fmt"
)

// The eight-character Capacity label needs 48pt at the pinned native 10pt Mono
// size. A 50% hole provides a 53.75pt band in this source specimen, retaining
// its outer circle, center label, data and fonts while clearing both arcs.
func applyV5NativeChartRefinement(key string, doc *SlideSpec) error {
	if key != "value-types/hard-soft-right" {
		return nil
	}
	if len(doc.Nodes) == 0 || doc.Nodes[0].ID != "node01" || doc.Nodes[0].Scene == nil {
		return fmt.Errorf("library.v5_native_chart_target_mismatch")
	}
	n, err := libraryObject(doc.Nodes[0].Scene.Node)
	if err != nil {
		return err
	}
	if n["type"] != "chart" || n["kind"] != "doughnut" {
		return fmt.Errorf("library.v5_native_chart_target_mismatch")
	}
	n["holeSize"] = 50
	raw, err := json.Marshal(n)
	if err != nil {
		return err
	}
	doc.Nodes[0].Scene.Node = raw
	return nil
}
