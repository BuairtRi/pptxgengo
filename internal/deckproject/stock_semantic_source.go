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
	annotateStockSlide(node, metadata)
	return encodeSourceYAML(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}})
}

func annotateStockSlide(node *yaml.Node, metadata wmdesign.LibraryAuthoring) {
	// YAML may associate a preceding generated comment with a mapping key or
	// its parent after a read/write cycle. Remove old generated blocks from all
	// comment positions before adding the current metadata exactly once.
	descriptions := make(map[string]bool, len(metadata.Slots))
	for _, slot := range metadata.Slots {
		descriptions[slot.Description] = true
	}
	stripGeneratedStockComments(node, descriptions)
	for _, slot := range metadata.Slots {
		if slot.Classification == "decorative" {
			continue
		}
		leaf := stockYAMLAt(node, "/content"+slot.Alias)
		// Hand-authored aliases may have different names. Binding destinations
		// still identify the pinned stock slot exactly.
		if leaf == nil {
			bindings := mappingNode(node, "bindings")
			if bindings != nil {
				for i := 0; i < len(bindings.Content); i += 2 {
					target := bindings.Content[i+1].Value
					rawSlot := target == "/slots/"+escape(slot.Name)
					typedSlot := slot.SourcePointer == target || strings.HasPrefix(slot.SourcePointer, target+"/")
					if rawSlot || typedSlot {
						alias := bindings.Content[i].Value
						if !strings.HasPrefix(alias, "/") {
							alias = "/" + alias
						}
						if !rawSlot {
							alias += strings.TrimPrefix(slot.SourcePointer, target)
						}
						leaf = stockYAMLAt(node, "/content"+alias)
						break
					}
				}
			}
		}
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
		human := strings.Trim(leaf.HeadComment, "\n")
		generated := "Stock slot: " + slot.Description + "\n" + capacity + "\nMetadata: " + slot.ReviewStatus + "; native fit not evaluated."
		if human != "" {
			generated = human + "\n" + generated
		}
		leaf.HeadComment = generated
	}
}

func stockCommentLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "#") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
	}
	return line
}

// Strip generated capacity/metadata blocks wherever the YAML parser attached
// them. Other author comments retain their original text and location.
func stockHumanComments(comment string, descriptions map[string]bool) string {
	lines := strings.Split(comment, "\n")
	drop := map[int]bool{}
	for i := 0; i+1 < len(lines); i++ {
		current, next := stockCommentLine(lines[i]), stockCommentLine(lines[i+1])
		capacity := strings.HasPrefix(current, "approximate capacity:") || strings.HasPrefix(current, "capacity unavailable (")
		metadata := strings.HasPrefix(next, "Metadata: ") && strings.HasSuffix(next, "; native fit not evaluated.")
		if !capacity || !metadata {
			continue
		}
		drop[i], drop[i+1] = true, true
		if i > 0 {
			previous := stockCommentLine(lines[i-1])
			if strings.HasPrefix(previous, "Stock slot: ") || descriptions[previous] || strings.Contains(previous, "; pinned source field /") || strings.Contains(previous, " at /") || strings.HasPrefix(previous, "Phase ") {
				drop[i-1] = true
			}
		}
	}
	kept := []string{}
	for i, line := range lines {
		if !drop[i] {
			kept = append(kept, line)
		}
	}
	return strings.Trim(strings.Join(kept, "\n"), "\n")
}

func stripGeneratedStockComments(node *yaml.Node, descriptions map[string]bool) {
	if node == nil {
		return
	}
	node.HeadComment = stockHumanComments(node.HeadComment, descriptions)
	node.LineComment = stockHumanComments(node.LineComment, descriptions)
	node.FootComment = stockHumanComments(node.FootComment, descriptions)
	for _, child := range node.Content {
		stripGeneratedStockComments(child, descriptions)
	}
}

// Content replacement retains comments attached to matching authored paths.
// Regenerated stock metadata is removed separately; human notes survive edits.
func retainContentComments(previous, replacement *yaml.Node) {
	if previous == nil || replacement == nil {
		return
	}
	if replacement.HeadComment == "" {
		replacement.HeadComment = previous.HeadComment
	}
	if replacement.LineComment == "" {
		replacement.LineComment = previous.LineComment
	}
	if replacement.FootComment == "" {
		replacement.FootComment = previous.FootComment
	}
	if previous.Kind != replacement.Kind {
		return
	}
	switch previous.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(replacement.Content); i += 2 {
			for j := 0; j+1 < len(previous.Content); j += 2 {
				if replacement.Content[i].Value == previous.Content[j].Value {
					retainContentComments(previous.Content[j], replacement.Content[i])
					retainContentComments(previous.Content[j+1], replacement.Content[i+1])
					break
				}
			}
		}
	case yaml.SequenceNode:
		for i := 0; i < len(replacement.Content) && i < len(previous.Content); i++ {
			retainContentComments(previous.Content[i], replacement.Content[i])
		}
	}
}

func refreshStockComments(node *yaml.Node, bundle string) error {
	reference := mappingNode(node, "template")
	if reference == nil || mappingNode(node, "content") == nil {
		return nil
	}
	var ref Reference
	if err := reference.Decode(&ref); err != nil {
		return err
	}
	if ref.Scope != "shared" {
		return nil
	}
	catalog, err := wmdesign.LibraryCatalog(bundle, "")
	if err != nil {
		return err
	}
	for _, def := range catalog {
		if def.Key != ref.ID {
			continue
		}
		metadata, err := wmdesign.LibraryAuthoringMetadata(def)
		if err != nil {
			return err
		}
		typography, err := wmdesign.NewTypography(filepath.Join(bundle, "fonts"))
		if err != nil {
			return err
		}
		for i := range metadata.Slots {
			metadata.Slots[i].Capacity, err = wmdesign.EstimateLibrarySlotCapacity(typography, metadata.Slots[i].Capacity)
			if err != nil {
				return err
			}
		}
		annotateStockSlide(node, metadata)
		return nil
	}
	return fmt.Errorf("unknown shared template %s", ref.ID)
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
