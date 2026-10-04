package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SlideEdit replaces complete values rather than merging arrays or guessing
// semantic correspondence between unrelated template bindings.
type SlideEdit struct {
	Template *Reference        `json:"template,omitempty"`
	Values   map[string]any    `json:"values,omitempty"`
	Content  map[string]any    `json:"content,omitempty"`
	Bindings map[string]string `json:"bindings,omitempty"`
	Brief    *string           `json:"brief,omitempty"`
}

type EditReceipt struct {
	Operation    string   `json:"operation"`
	SlideIDs     []string `json:"slide_ids"`
	BeforeSHA256 string   `json:"before_sha256"`
	AfterSHA256  string   `json:"after_sha256"`
	Decision     string   `json:"decision"`
	Validation   string   `json:"validation"`
}

type EditOptions struct {
	CheckFit  bool
	Operation string
}

// DecodeSlideEdits accepts readable YAML or JSON with the same strict key,
// scalar, alias and single-document rules as the maintained deck source.
func DecodeSlideEdits(raw []byte, sourcePath string) (map[string]SlideEdit, error) {
	if len(raw) > 16<<20 {
		return nil, fmt.Errorf("slide patch exceeds 16 MiB")
	}
	if err := flowMapCommaDiagnostic(raw, sourcePath); err != nil {
		return nil, err
	}
	p := &Project{SourcePath: sourcePath, Positions: map[string]Position{}}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document, extra yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 {
		return nil, fmt.Errorf("empty slide patch")
	}
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("slide patch requires exactly one YAML or JSON document")
	}
	value, err := p.yamlValue(document.Content[0], "", 0)
	if err != nil {
		return nil, err
	}
	var edits map[string]SlideEdit
	if err = p.shapeType(value, reflect.TypeOf(edits), ""); err != nil {
		return nil, err
	}
	if err = strictInto(value, &edits); err != nil {
		return nil, err
	}
	return edits, nil
}

// EditSlides validates the complete resulting project before atomically
// replacing deck.yaml. It retains untouched YAML nodes and exact predecessors.
func EditSlides(p *Project, edits map[string]SlideEdit, bundle, engine string) (EditReceipt, error) {
	return EditSlidesWithOptions(p, edits, bundle, engine, EditOptions{})
}

func EditSlidesWithOptions(p *Project, edits map[string]SlideEdit, bundle, engine string, options EditOptions) (EditReceipt, error) {
	if options.Operation != "" && options.Operation != "swap" {
		return EditReceipt{}, fmt.Errorf("unsupported edit operation %q", options.Operation)
	}
	if p.hasExternalSources() || p.hasContentAliases() || extendedSlideEdits(edits) {
		return editSourceSlidesWithOptions(p, edits, bundle, engine, options)
	}
	r := EditReceipt{Operation: "edit-slides", SlideIDs: []string{}, BeforeSHA256: digest(p.Raw), Validation: "source_and_binding_checked_native_review_pending"}
	if options.Operation != "" {
		r.Operation = options.Operation
	}
	if len(edits) == 0 {
		return r, fmt.Errorf("slide patch must contain at least one stable slide ID")
	}
	var document yaml.Node
	if err := yaml.Unmarshal(p.Raw, &document); err != nil {
		return r, err
	}
	slides := mappingNode(document.Content[0], "slides")
	seen := map[string]bool{}
	for _, node := range slides.Content {
		id := mappingNode(node, "id").Value
		edit, ok := edits[id]
		if !ok {
			continue
		}
		if edit.Template == nil && edit.Values == nil && edit.Brief == nil {
			return r, fmt.Errorf("slide %s has no edit fields", id)
		}
		seen[id] = true
		r.SlideIDs = append(r.SlideIDs, id)
		fields := []struct {
			name  string
			value any
			used  bool
		}{{"template", edit.Template, edit.Template != nil}, {"values", edit.Values, edit.Values != nil}, {"brief", edit.Brief, edit.Brief != nil}}
		for _, field := range fields {
			if !field.used {
				continue
			}
			replacement, err := editYAMLNode(field.value)
			if err != nil {
				return r, err
			}
			quoteLeadingNewlines(replacement)
			if old := mappingNode(node, field.name); old != nil {
				replacement.HeadComment, replacement.LineComment, replacement.FootComment = old.HeadComment, old.LineComment, old.FootComment
				*old = *replacement
			} else {
				node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: field.name}, replacement)
			}
		}
	}
	for id := range edits {
		if !seen[id] {
			return r, fmt.Errorf("unknown stable slide ID %s", id)
		}
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return r, err
	}
	if err := encoder.Close(); err != nil {
		return r, err
	}
	raw := output.Bytes()
	r.AfterSHA256 = digest(raw)
	temp, err := SafePath(p.Root, ".deck-edit-"+nonce()+".yaml")
	if err != nil {
		return r, err
	}
	stat, err := os.Stat(p.SourcePath)
	if err != nil {
		return r, err
	}
	if err = writeExclusive(temp, raw, stat.Mode().Perm()); err != nil {
		return r, err
	}
	defer os.Remove(temp)
	candidate, err := Load(temp)
	if err != nil {
		return r, err
	}
	if _, err = Compile(candidate, bundle, engine); err != nil {
		return r, err
	}
	if options.CheckFit {
		if err = CheckSlideFit(candidate, r.SlideIDs, bundle, engine); err != nil {
			return r, err
		}
		r.Validation = "source_binding_and_go_fit_checked_native_review_pending"
	}
	guard, err := SafePath(p.Root, ".deck-source-mutation.lock")
	if err != nil {
		return r, err
	}
	if err = writeExclusive(guard, []byte(r.BeforeSHA256), 0600); err != nil {
		return r, fmt.Errorf("another slide edit is active: %w", err)
	}
	defer os.Remove(guard)
	current, err := os.ReadFile(p.SourcePath)
	if err != nil {
		return r, err
	}
	if !bytes.Equal(current, p.Raw) {
		return r, fmt.Errorf("source changed during slide edit")
	}
	before, err := SafePath(p.Root, "decisions/sources/"+r.BeforeSHA256+".yaml")
	if err != nil {
		return r, err
	}
	if stored, e := os.ReadFile(before); e == nil {
		if !bytes.Equal(stored, p.Raw) {
			return r, fmt.Errorf("source predecessor snapshot drift")
		}
	} else if os.IsNotExist(e) {
		if err = writeExclusive(before, p.Raw, 0444); err != nil {
			return r, err
		}
	} else {
		return r, e
	}
	r.Decision = "decisions/" + r.Operation + "-" + time.Now().UTC().Format("20060102T150405") + "-" + nonce() + ".json"
	decision, err := SafePath(p.Root, r.Decision)
	if err != nil {
		return r, err
	}
	if err = writeJSON(decision, r); err != nil {
		return r, err
	}
	if err = os.Rename(temp, p.SourcePath); err != nil {
		os.Remove(decision)
		return r, err
	}
	return r, nil
}

func quoteLeadingNewlines(n *yaml.Node) {
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && strings.HasPrefix(n.Value, "\n") {
		n.Style = yaml.DoubleQuotedStyle
	}
	for _, child := range n.Content {
		quoteLeadingNewlines(child)
	}
}

// Node.Encode serializes then reparses strings, which can discard leading
// blank lines in literal blocks. Construct scalar nodes before serialization.
func editYAMLNode(value any) (*yaml.Node, error) {
	var normalized any
	if err := json.Unmarshal(canonical(value), &normalized); err != nil {
		return nil, err
	}
	var convert func(any) *yaml.Node
	convert = func(value any) *yaml.Node {
		n := &yaml.Node{}
		switch v := value.(type) {
		case map[string]any:
			n.Kind, n.Tag = yaml.MappingNode, "!!map"
			keys := make([]string, 0, len(v))
			for key := range v {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, convert(v[key]))
			}
		case []any:
			n.Kind, n.Tag = yaml.SequenceNode, "!!seq"
			for _, item := range v {
				n.Content = append(n.Content, convert(item))
			}
		case string:
			n.Kind, n.Tag, n.Value = yaml.ScalarNode, "!!str", v
			if strings.HasPrefix(v, "\n") {
				n.Style = yaml.DoubleQuotedStyle
			} else if strings.Contains(v, "\n") {
				n.Style = yaml.LiteralStyle
			}
		case bool:
			n.Kind, n.Tag, n.Value = yaml.ScalarNode, "!!bool", strconv.FormatBool(v)
		case float64:
			n.Kind, n.Tag, n.Value = yaml.ScalarNode, "!!float", strconv.FormatFloat(v, 'g', -1, 64)
		case nil:
			n.Kind, n.Tag, n.Value = yaml.ScalarNode, "!!null", "null"
		}
		return n
	}
	return convert(normalized), nil
}
