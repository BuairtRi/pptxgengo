package assetregistry

import "testing"

func TestCatalogReturnsIndependentCanonicalPins(t *testing.T) {
	first := Catalog()
	canonical := first["arrow-connecting"]
	if canonical.SHA256 != "646a5a89f976631f409f2aff52223cc21f54ba077b6552537613a45d956307ed" {
		t.Fatal("known canonical identity changed")
	}
	delete(first, "arrow-connecting")
	first["icon/fixture"] = Reference{Path: "fake.svg", SHA256: "fake"}
	second := Catalog()
	if second["arrow-connecting"] != canonical {
		t.Fatal("catalog caller mutated trusted registry")
	}
	if _, exists := second["icon/fixture"]; exists {
		t.Fatal("catalog caller inserted trusted registry identity")
	}
}
