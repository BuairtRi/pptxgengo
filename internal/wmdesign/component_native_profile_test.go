package wmdesign

import (
	"archive/zip"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

func TestNativeProfileTypedCardBulletsAndPlainCardSerializeMeasuredText(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	r.editingProfile = NativeEditingProfile
	zone := Rect{X: 0, Y: 0, W: 960, H: 540}
	bulletNode := Node{ID: "typed-list-card", Kind: "card", Surface: "light", Card: &CardSpec{Title: "A measured checklist", Body: []BodyBlock{{Key: "steps", Bullets: []BulletItem{{Key: "owner", Text: "Assign an owner to each exception."}, {Key: "date", Text: "Confirm the next review date."}}}}}}
	bulletPlan, err := r.planComponent(bulletNode, Rect{X: 57, Y: 144, W: 414, H: 216}, zone, "light")
	if err != nil {
		t.Fatal(err)
	}
	if len(bulletPlan.texts) != 2 || bulletPlan.texts[1].NativeParagraphContract != EditableListContract || len(bulletPlan.texts[1].Rich.Paragraphs) != 2 {
		t.Fatalf("typed card bullet block did not become a native list: %+v", bulletPlan)
	}
	if len(bulletPlan.record.Shapes) != 1 || len(bulletPlan.record.Parts) != 3 || bulletPlan.record.Parts[2] != "typed-list-card.body.steps" {
		t.Fatalf("component lineage retained removed marker objects: %+v", bulletPlan.record)
	}

	plainNode := Node{ID: "typed-plain-card", Kind: "card", Surface: "light", Card: &CardSpec{Title: "A paragraph card", Body: []BodyBlock{{Key: "copy", Paragraph: "One editable paragraph keeps its measured source allocation."}}}}
	plainPlan, err := r.planComponent(plainNode, Rect{X: 507, Y: 144, W: 414, H: 216}, zone, "light")
	if err != nil {
		t.Fatal(err)
	}
	if !plainPlan.record.NativeObject || len(plainPlan.record.Shapes) != 0 || len(plainPlan.record.Parts) != 1 || plainPlan.record.Parts[0] != plainNode.ID || len(plainPlan.texts) != 1 || plainPlan.texts[0].NativeShape == nil || plainPlan.texts[0].ID != plainNode.ID {
		t.Fatalf("plain typed card did not merge into the source container object: %+v", plainPlan)
	}

	pres := pptx.New()
	r.slide = pres.AddSlide()
	var records []TextRecord
	r.records = &records
	r.drawComponent(bulletPlan)
	r.drawComponent(plainPlan)
	if r.err != nil {
		t.Fatal(r.err)
	}
	deck, err := pres.Write()
	if err != nil {
		t.Fatal(err)
	}
	deck, err = candidateParagraphs(deck, map[int][]TextRecord{1: records})
	if err != nil {
		t.Fatal(err)
	}
	report := SlideReport{ID: "native-typed-cards", Texts: records, Components: []ComponentRecord{bulletPlan.record, plainPlan.record}}
	if err := validateOwnedParts(report); err != nil {
		t.Fatalf("native component ownership is ambiguous: %v", err)
	}
	deck, err = componentGroups(deck, []SlideReport{report})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(deck), int64(len(deck)))
	if err != nil {
		t.Fatal(err)
	}
	var slideXML []byte
	for _, f := range zr.File {
		if f.Name != "ppt/slides/slide1.xml" {
			continue
		}
		reader, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		_, err = b.ReadFrom(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		slideXML = b.Bytes()
	}
	text := string(slideXML)
	for _, want := range []string{`name="typed-plain-card"`, `name="typed-list-card.body.steps"`, `typeface="Wingdings" charset="2"`, `<a:buClr><a:srgbClr val="` + bulletPlan.texts[1].Rich.Paragraphs[0].BulletColor + `"/></a:buClr>`, `name="typed-list-card.container"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("serialized typed component XML missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, `name="typed-plain-card.container"`) || strings.Count(text, `name="typed-plain-card"`) != 1 {
		t.Fatalf("merged card left a phantom container object: %s", text)
	}
}

func TestNativeProfileCanonicalCardsThreeBuildRetainsDecorativeCards(t *testing.T) {
	bundle := densityTestBundle()
	input := TemplateReference(2026)
	if len(input.Slides) == 0 || input.Slides[0].Template != "cards/3" {
		t.Fatal("canonical cards/3 fixture moved")
	}
	input.Slides = input.Slides[:1]
	doc, _, err := BindTemplates(bundle, "", input)
	if err != nil {
		t.Fatal(err)
	}
	doc.EditingProfile = NativeEditingProfile
	deck, report, err := BuildWithEngine(bundle, "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if len(deck) == 0 || len(report.Slides) != 1 || len(report.Slides[0].Components) != 3 {
		t.Fatalf("canonical cards/3 did not build as one three-card slide: %+v", report.Slides)
	}
	nativeCards := 0
	for _, component := range report.Slides[0].Components {
		if !component.NativeObject {
			continue
		}
		nativeCards++
		if len(component.Parts) != 1 || component.Parts[0] != component.ID {
			t.Fatalf("native card component identity is not preserved: %+v", component)
		}
		found := false
		for _, text := range report.Slides[0].Texts {
			if text.ID == component.ID && text.NativeShape != nil && text.NativeShape.ParagraphContract == EditableCardContract {
				found = true
			}
		}
		if !found {
			t.Fatalf("native card has no matching editable text record: %s", component.ID)
		}
	}
	if nativeCards != 0 {
		t.Fatalf("decorated canonical cards/3 should retain their separate objects, converted %d", nativeCards)
	}
	for _, component := range report.Slides[0].Components {
		if component.Definition != "card.band" || len(component.Shapes) < 2 {
			t.Fatalf("canonical cards/3 lost its source number/band composition: %+v", component)
		}
	}
}
