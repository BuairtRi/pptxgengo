package wmdesign

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCandidateCalibrationCurrentPathDoesNotMaskDrift(t *testing.T) {
	calibration, err := os.ReadFile(filepath.Join("..", "..", "planning", "wm-design-contracts", "v5", "intake-20261003-587-frozen", "bundle", "typography", "calibration.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"valid", "changed", "directory", "absent"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			fonts := filepath.Join(root, "v5", "fonts")
			current := filepath.Join(root, "v5", "typography", "calibration.json")
			legacy := filepath.Join(root, "typography-v2-candidate", "calibration.json")
			// Font resolution is checked elsewhere. An empty font directory
			// isolates calibration path selection and pinned-byte validation.
			for _, directory := range []string{fonts, filepath.Dir(current), filepath.Dir(legacy)} {
				if err := os.MkdirAll(directory, 0755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(legacy, calibration, 0644); err != nil {
				t.Fatal(err)
			}
			switch state {
			case "valid", "changed":
				data := calibration
				if state == "changed" {
					data = append(append([]byte(nil), calibration...), '\n')
				}
				if err := os.WriteFile(current, data, 0644); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(current, 0755); err != nil {
					t.Fatal(err)
				}
			}
			path, err := CandidateCalibrationPath(fonts)
			if err != nil {
				t.Fatal(err)
			}
			want := current
			if state == "absent" {
				want = legacy
			}
			if filepath.Clean(path) != want {
				t.Fatalf("selected %s, want %s", path, want)
			}
			_, err = NewTypographyEngine(fonts, CandidateEngine)
			switch state {
			case "valid", "absent":
				if err != nil {
					t.Fatal(err)
				}
			case "changed":
				if err == nil || !strings.Contains(err.Error(), "text.calibration_drift") {
					t.Fatalf("present current drift masked by valid legacy copy: %v", err)
				}
			case "directory":
				if err == nil {
					t.Fatal("unreadable current calibration masked by valid legacy copy")
				}
			}
		})
	}
}
