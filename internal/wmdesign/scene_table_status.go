package wmdesign

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// The frozen source table renderer chooses a status cell's explicit label,
// then its column label map, then the canonical status label. Its mark keeps
// the canonical status color regardless of which visible label was chosen.
func sceneTableStatusLabels(column sceneTableColumn) error {
	if column.Labels == nil {
		return nil
	}
	if column.Type != "status" || len(column.Labels) > len(sceneStatuses) {
		return fmt.Errorf("scene.table_status_labels_type_or_count: %s", column.Key)
	}
	for status := range column.Labels {
		if _, ok := sceneStatuses[status]; !ok {
			return fmt.Errorf("scene.table_status_labels_enum: %s", status)
		}
	}
	return nil
}

func sceneTableStatusValue(raw json.RawMessage, labels map[string]string) (string, string, error) {
	var value struct {
		Status string `json:"status"`
		Label  string `json:"label,omitempty"`
	}
	data := bytes.TrimSpace(raw)
	if len(data) > 0 && data[0] == '{' {
		if err := sceneDecode(data, &value); err != nil {
			return "", "", err
		}
	} else if err := json.Unmarshal(data, &value.Status); err != nil {
		return "", "", fmt.Errorf("scene.table_status_string_or_object: %w", err)
	}
	status, ok := sceneStatuses[value.Status]
	if !ok {
		return "", "", fmt.Errorf("scene.table_status_enum: %s", value.Status)
	}
	label := value.Label
	if label == "" {
		label = labels[value.Status]
	}
	if label == "" {
		label = status.Label
	}
	return value.Status, label, nil
}
