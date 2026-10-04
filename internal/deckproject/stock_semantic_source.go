package deckproject

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
	"gopkg.in/yaml.v3"
)

// MarshalStockSlideSource adds source-pinned descriptions and drafting budgets
// as YAML comments. Comments never enter the closed values contract.
func MarshalStockSlideSource(authored map[string]any, metadata wmdesign.LibraryAuthoring) ([]byte, error) {
	node, e := editYAMLNode(authored)
	if e != nil {
		return nil, e
	}
	orderAuthoredSlide(node)
	for _, slot := range metadata.Slots {
		if slot.Classification == "decorative" {
			continue
		}
		leaf := stockYAMLAt(node, "/content"+slot.Alias)
		if leaf == nil {
			continue
		}
		c := slot.Capacity
		capacity := "capacity unavailable (" + c.Basis + ")"
		if c.LineBudget > 0 && c.Style != nil {
			capacity = fmt.Sprintf("approximate capacity: %d line(s), %.0f pt wide at %.0f pt type", c.LineBudget, c.WidthPt, c.Style.Size)
			if c.ApproxCharacters > 0 {
				capacity = fmt.Sprintf("approximate capacity: ~%d characters, %d line(s) at %.0f pt type; Latin prose estimate, not a maximum", c.ApproxCharacters, c.LineBudget, c.Style.Size)
			}
		}
		leaf.HeadComment = slot.Description + "\n" + capacity + "\nMetadata: " + slot.ReviewStatus + "; native fit not evaluated."
	}
	return encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
}
func stockYAMLAt(node *yaml.Node, pointer string) *yaml.Node {
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node.Kind {
		case yaml.MappingNode:
			var next *yaml.Node
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == part {
					next = node.Content[i+1]
					break
				}
			}
			if next == nil {
				return nil
			}
			node = next
		case yaml.SequenceNode:
			i, e := strconv.Atoi(part)
			if e != nil || i < 0 || i >= len(node.Content) {
				return nil
			}
			node = node.Content[i]
		default:
			return nil
		}
	}
	return node
}
func StockScaffoldSlideSource(bundle, key, id string, year int) ([]byte, error) {
	authored, e := StockScaffoldSlide(bundle, key, id, year)
	if e != nil {
		return nil, e
	}
	catalog, e := wmdesign.LibraryCatalog(bundle, "")
	if e != nil {
		return nil, e
	}
	for _, def := range catalog {
		if def.Key != key {
			continue
		}
		metadata, e := wmdesign.LibraryAuthoringMetadata(def)
		if e != nil {
			return nil, e
		}
		typography, e := wmdesign.NewTypography(filepath.Join(bundle, "fonts"))
		if e != nil {
			return nil, e
		}
		for i := range metadata.Slots {
			metadata.Slots[i].Capacity, e = wmdesign.EstimateLibrarySlotCapacity(typography, metadata.Slots[i].Capacity)
			if e != nil {
				return nil, e
			}
		}
		return MarshalStockSlideSource(authored, metadata)
	}
	return nil, fmt.Errorf("unknown shared template %s", key)
}
