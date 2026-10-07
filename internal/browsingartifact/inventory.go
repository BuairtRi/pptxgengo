package browsingartifact

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// ReleaseInputs bind release-generated documents to immutable private archives
// and the authoritative pipeline time. Standalone CLI documents need not have
// these fields; installation archives require the complete release inventory.
type ReleaseInputs struct {
	BrandingArchiveSHA256        string `json:"branding_archive_sha256"`
	FinishedLibraryArchiveSHA256 string `json:"finished_library_archive_sha256"`
	AsOf                         string `json:"as_of"`
	PipelineCreatedAt            string `json:"pipeline_created_at"`
}

type Inventory struct {
	ReleaseInputs
	Schema          string            `json:"schema"`
	Files           map[string]string `json:"files_sha256"`
	BundleSHA256    string            `json:"bundle_sha256"`
	SourceRevision  string            `json:"source_revision"`
	SourceCommit    string            `json:"source_commit"`
	Compiler        string            `json:"compiler"`
	ReleaseIdentity string            `json:"release_identity"`
}

type InventoryExpectations struct {
	Inputs          ReleaseInputs
	ReleaseIdentity string
}

var inventoryFiles = []string{"browsing/template-library.pptx", "browsing/template-library.manifest.json", "browsing/reusable-slides.pptx", "browsing/reusable-slides.manifest.json"}

func ReadInventory(raw []byte) (Inventory, error) {
	var inventory Inventory
	if len(raw) > MaxManifestBytes {
		return inventory, fmt.Errorf("browsing.inventory_exceeds_16MiB")
	}
	if e := json.Unmarshal(raw, &inventory); e != nil {
		return inventory, e
	}
	pipeline, e := time.Parse(time.RFC3339Nano, inventory.PipelineCreatedAt)
	if e != nil || pipeline.UTC().Format(time.DateOnly) != inventory.AsOf || !hash(inventory.BrandingArchiveSHA256) || !hash(inventory.FinishedLibraryArchiveSHA256) || inventory.Schema != "pptxgengo.release-browsing-files.v1" || !hash(inventory.BundleSHA256) || inventory.SourceRevision == "" || inventory.SourceCommit == "" || inventory.Compiler == "" || inventory.ReleaseIdentity == "" || len(inventory.Files) != len(inventoryFiles) {
		return inventory, fmt.Errorf("browsing.release_inventory_pins_invalid")
	}
	for _, file := range inventoryFiles {
		if !hash(inventory.Files[file]) {
			return inventory, fmt.Errorf("browsing.release_inventory_file_missing: %s", file)
		}
	}
	return inventory, nil
}

// ValidateInventory closes both independently verified documents under one
// release. Expected protected inputs are supplied by CI; installers rely on the
// enclosing signed archive and still enforce mutual source/input coherence.
func ValidateInventory(inventory Inventory, manifests map[string][]byte, expected *InventoryExpectations) error {
	if expected != nil && (!reflect.DeepEqual(inventory.ReleaseInputs, expected.Inputs) || inventory.ReleaseIdentity != expected.ReleaseIdentity) {
		return fmt.Errorf("browsing.release_inventory_protected_inputs_mismatch")
	}
	if len(manifests) != 2 {
		return fmt.Errorf("browsing.release_inventory_both_manifests_required")
	}
	for _, kind := range []string{"templates", "reusable"} {
		raw, ok := manifests[kind]
		if !ok || len(raw) > MaxManifestBytes {
			return fmt.Errorf("browsing.release_inventory_manifest_missing: %s", kind)
		}
		var m Manifest
		if e := json.Unmarshal(raw, &m); e != nil {
			return e
		}
		if m.Kind != kind || m.BundleSHA256 != inventory.BundleSHA256 || m.SourceRevision != inventory.SourceRevision || m.SourceCommit != inventory.SourceCommit || m.Compiler != inventory.Compiler || m.ReleaseIdentity != inventory.ReleaseIdentity || m.AsOf != inventory.AsOf || !reflect.DeepEqual(m.ReleaseInputs, inventory.ReleaseInputs) {
			return fmt.Errorf("browsing.release_inventory_cross_deck_pins_mismatch: %s", kind)
		}
	}
	return nil
}
