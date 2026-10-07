package browsingartifact_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/browsingartifact"
	"github.com/buairtri/pptxgengo/internal/browsingfixture"
)

func TestBrowsingInventoryMutualReleasePins(t *testing.T) {
	pin := strings.Repeat("a", 64)
	files := map[string]string{}
	for _, name := range []string{"template-library.pptx", "template-library.manifest.json", "reusable-slides.pptx", "reusable-slides.manifest.json"} {
		files["browsing/"+name] = pin
	}
	inventory, e := browsingartifact.ReadInventory(browsingfixture.Inventory(t, files))
	if e != nil {
		t.Fatal(e)
	}
	expected := &browsingartifact.InventoryExpectations{Inputs: browsingfixture.Inputs(), ReleaseIdentity: "unit-test-only"}
	makeManifests := func() map[string][]byte {
		return map[string][]byte{"templates": browsingfixture.Manifest(t, "templates", pin), "reusable": browsingfixture.Manifest(t, "reusable", pin)}
	}
	if e = browsingartifact.ValidateInventory(inventory, makeManifests(), expected); e != nil {
		t.Fatal(e)
	}
	for name, change := range map[string]func(map[string]any){
		"bundle":          func(m map[string]any) { m["bundle_sha256"] = strings.Repeat("c", 64) },
		"source_revision": func(m map[string]any) { m["source_revision"] = "another-source" },
		"source_commit":   func(m map[string]any) { m["source_commit"] = strings.Repeat("c", 40) },
		"source_file_pin": func(m map[string]any) {
			m["source_files"].([]any)[0].(map[string]any)["sha256"] = strings.Repeat("c", 64)
		},
		"release":  func(m map[string]any) { m["release_identity"] = "another-release" },
		"compiler": func(m map[string]any) { m["compiler"] = "another-compiler" },
		"as_of":    func(m map[string]any) { m["as_of"] = "2026-10-08" },
		"branding_input": func(m map[string]any) {
			m["release_inputs"].(map[string]any)["branding_archive_sha256"] = strings.Repeat("c", 64)
		},
		"finished_input": func(m map[string]any) {
			m["release_inputs"].(map[string]any)["finished_library_archive_sha256"] = strings.Repeat("c", 64)
		},
		"pipeline_time": func(m map[string]any) {
			m["release_inputs"].(map[string]any)["pipeline_created_at"] = "2026-10-07T13:00:00Z"
		},
	} {
		t.Run(name, func(t *testing.T) {
			manifests := makeManifests()
			var m map[string]any
			if e := json.Unmarshal(manifests["reusable"], &m); e != nil {
				t.Fatal(e)
			}
			change(m)
			manifests["reusable"], _ = json.Marshal(m)
			if e := browsingartifact.ValidateInventory(inventory, manifests, expected); e == nil {
				t.Fatal("mixed releases or input archives accepted")
			}
		})
	}
	for name, change := range map[string]func(*browsingartifact.InventoryExpectations){
		"protected_branding": func(e *browsingartifact.InventoryExpectations) {
			e.Inputs.BrandingArchiveSHA256 = strings.Repeat("c", 64)
		},
		"protected_finished": func(e *browsingartifact.InventoryExpectations) {
			e.Inputs.FinishedLibraryArchiveSHA256 = strings.Repeat("c", 64)
		},
		"protected_pipeline": func(e *browsingartifact.InventoryExpectations) { e.Inputs.PipelineCreatedAt = "2026-10-07T13:00:00Z" },
		"protected_release":  func(e *browsingartifact.InventoryExpectations) { e.ReleaseIdentity = "another-release" },
	} {
		t.Run(name, func(t *testing.T) {
			modified := *expected
			change(&modified)
			if e := browsingartifact.ValidateInventory(inventory, makeManifests(), &modified); e == nil {
				t.Fatal("protected CI input mismatch accepted")
			}
		})
	}
	raw := browsingfixture.Inventory(t, files)
	var bad map[string]any
	json.Unmarshal(raw, &bad)
	bad["pipeline_created_at"] = "2026-10-08T12:00:00Z"
	raw, _ = json.Marshal(bad)
	if _, e = browsingartifact.ReadInventory(raw); e == nil {
		t.Fatal("pipeline UTC date and freshness date mismatch accepted")
	}
}
