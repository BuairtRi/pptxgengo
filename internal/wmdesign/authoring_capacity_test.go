package wmdesign

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestTypedCardAuthoringCapacityMatchesRendererBounds(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	definitions, err := LibraryCatalogFromSource(source)
	if err != nil {
		t.Fatal(err)
	}
	typography, err := NewTypographyEngine(filepath.Join(bundle, "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"cards/3", "cards/4"} {
		var definition LibraryTemplate
		for _, candidate := range definitions {
			if candidate.Key == key {
				definition = candidate
				break
			}
		}
		if definition.Authoring == nil {
			t.Fatal("missing typed card metadata", key)
		}
		count := definition.ValueSchema.ExactCounts["cards"]
		capacities := map[string]LibrarySlotCapacity{}
		for _, slot := range definition.Authoring.Slots {
			capacities[slot.Alias] = slot.Capacity
		}
		slide, _, err := compileTemplateSource(source, key)
		if err != nil {
			t.Fatal(err)
		}
		frame, err := source.ResolveFrame(slide.Frame)
		if err != nil {
			t.Fatal(err)
		}
		row := slide.Nodes[0]
		for i := range row.CardRow.Items {
			title := capacities[fmt.Sprintf("/cards/%d/title", i)]
			body := capacities[fmt.Sprintf("/cards/%d/body", i)]
			if title.Style == nil || body.Style == nil || title.LineBudget != 2 || body.LineBudget < 1 || body.ApproxCharacters < 1 || body.NativeFit != "not_evaluated" {
				t.Fatal("missing pinned typed card capacity", key, title, body)
			}
			row.CardRow.Items[i].Card.Title = "Capacity\nCapacity"
			row.CardRow.Items[i].Card.Body[0].Paragraph = strings.TrimSuffix(strings.Repeat("M\n", body.LineBudget), "\n")
		}
		r := renderer{source: source, typeEngine: typography}
		plan, err := r.planCardRow(row, row.Rect, frame.Body, row.Surface)
		if err != nil || len(plan.children) != count {
			t.Fatal("reported card budget did not fit actual renderer", key, err)
		}
		for i, child := range plan.children {
			for _, text := range child.texts {
				if !strings.HasSuffix(text.ID, ".title") && !strings.HasSuffix(text.ID, ".body.copy") {
					continue
				}
				field := "title"
				if strings.HasSuffix(text.ID, ".body.copy") {
					field = "body"
				}
				c := capacities[fmt.Sprintf("/cards/%d/%s", i, field)]
				if c.WidthPt != text.Rect.W {
					t.Fatal("capacity did not use renderer number/padding geometry", key, field, c.WidthPt, text.Rect.W)
				}
			}
		}
		row.CardRow.Items[0].Card.Body[0].Paragraph += "\nM"
		if _, err = r.planCardRow(row, row.Rect, frame.Body, row.Surface); err == nil {
			t.Fatal("card body overrun was not rejected by renderer", key)
		}
	}
	counts := map[string]int{}
	unsupported := 0
	for _, definition := range definitions {
		for _, slot := range definition.Authoring.Slots {
			if slot.Capacity.Status == "unsupported" {
				unsupported++
			} else if strings.HasPrefix(slot.Capacity.Basis, "pinned_") {
				counts[slot.Capacity.Basis]++
			}
		}
	}
	t.Logf("new planner-backed slot counts: %v; remaining unsupported slots: %d", counts, unsupported)
}

func TestFixedTableAndVerticalStepperAuthoringCapacity(t *testing.T) {
	bundle := filepath.Join("..", "..", "library", "wm-design-system", "v5")
	source, err := Load(bundle, "")
	if err != nil {
		t.Fatal(err)
	}
	obj := map[string]any{}
	if err = json.Unmarshal([]byte(`{"body":[{"type":"table","x":57,"y":144,"w":360,"rowH":54,"rowHeader":true,"cols":[{"k":"name","label":"Name","w":180},{"k":"owner","label":"Owner","w":180}],"rows":[{"name":"Item","owner":"Team"}]},{"type":"vstepper","x":57,"y":240,"w":360,"steps":[{"label":"01","title":"Review","text":"Evidence"},{"label":"02","title":"Deliver","text":"Result"}]}]}`), &obj); err != nil {
		t.Fatal(err)
	}
	capacities := fixedSceneAuthoringCapacities(LibraryTemplate{}, obj, source, authoringFonts(source))
	for _, pointer := range []string{"/body/0/cols/0/label", "/body/0/rows/0/name", "/body/0/rows/0/owner", "/body/1/steps/0/text", "/body/1/steps/1/text"} {
		c, found := capacities[pointer]
		if !found || c.Style == nil || c.LineBudget < 1 || c.NativeFit != "not_evaluated" {
			t.Fatal("fixed planner capacity absent", pointer, c)
		}
	}
	if capacities["/body/0/rows/0/name"].Style.Weight != 600 || capacities["/body/0/rows/0/owner"].Style.Weight != 400 {
		t.Fatal("native row-header weights did not reach capacity")
	}
	if _, found := capacities["/body/1/steps/0/title"]; found {
		t.Fatal("coupled inline heading incorrectly assigned fixed budget")
	}
	before, _ := json.Marshal(obj)
	_ = fixedSceneAuthoringCapacities(LibraryTemplate{}, obj, source, authoringFonts(source))
	after, _ := json.Marshal(obj)
	if string(before) != string(after) {
		t.Fatal("capacity planning changed source example")
	}
	// Typed status cells have marks and width reservations; leave them explicit.
	body := obj["body"].([]any)
	table := body[0].(map[string]any)
	table["cols"].([]any)[1].(map[string]any)["type"] = "status"
	capacities = fixedSceneAuthoringCapacities(LibraryTemplate{}, obj, source, authoringFonts(source))
	if _, found := capacities["/body/0/rows/0/owner"]; found {
		t.Fatal("marked table cell claimed plain text capacity")
	}
}
