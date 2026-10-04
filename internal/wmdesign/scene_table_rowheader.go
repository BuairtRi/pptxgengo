package wmdesign

import (
	"encoding/json"
	"fmt"
)

// The source renderer treats a nonempty string rowHeader as true, styling the
// first column. It does not select the named column. Keep that exact meaning.
type sceneTableRowHeader bool

func (v *sceneTableRowHeader) UnmarshalJSON(raw []byte) error {
	var flag bool
	if err := json.Unmarshal(raw, &flag); err == nil {
		*v = sceneTableRowHeader(flag)
		return nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil || len(text) > 128 {
		return fmt.Errorf("scene.table_row_header_requires_boolean_or_string")
	}
	*v = sceneTableRowHeader(text != "")
	return nil
}
