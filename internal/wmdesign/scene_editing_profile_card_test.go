package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/pptx"
)

func TestNativeProfileCardMultiParagraphPreservesStylesAndPositions(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ density, raw string }{
		{"comfortable", `{"type":"card","x":60,"y":100,"w":360,"h":230,"surface":"subtle","title":"Multi paragraph card","body":[{"p":"First paragraph with a measured wrap in the available space."},{"p":"Second paragraph keeps its own measured style."}]}`},
		{"compact", `{"type":"card","x":60,"y":100,"w":360,"h":230,"surface":"subtle","bodySize":"small","title":"Small body card","body":[{"p":"First small paragraph."},{"p":"Second small paragraph."}]}`},
	} {
		r.bodyDensity = tc.density
		r.editingProfile = ""
		before, err := r.planSceneNode("example", json.RawMessage(tc.raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		after := nativeProfileCardMulti(before, json.RawMessage(tc.raw))
		if after == nil || len(after.Items) != 1 || after.Items[0].Text == nil || len(after.Items[0].Text.Rich.Paragraphs) != 3 {
			t.Fatalf("card did not convert: %+v", after)
		}
		native := after.Items[0].Text
		paragraph := 0
		line := 0
		for _, item := range before.Items {
			if item.Text == nil {
				continue
			}
			got := native.Rich.Paragraphs[paragraph]
			if len(got.Runs) != 1 || got.Runs[0].Style != item.Text.Layout.Style || !reflect.DeepEqual(got.Runs[0].Font, item.Text.Layout.Font) || got.Runs[0].Color != item.Text.Color {
				t.Fatalf("paragraph style changed: %d", paragraph)
			}
			for _, sourceLine := range item.Text.Layout.Lines {
				want := item.Text.Rect.Y - native.Rect.Y + sourceLine.Baseline
				if native.Layout.Lines[line].Text != sourceLine.Text || math.Abs(native.Layout.Lines[line].Baseline-want) > .02 {
					t.Fatalf("line position changed: %d", line)
				}
				line++
			}
			paragraph++
		}
		if line != len(native.Layout.Lines) {
			t.Fatal("extra lines in combined card")
		}
	}
}

func TestNativeProfileCardFootnoteRunsKeepPerRunFormatting(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"card","x":60,"y":100,"w":360,"h":230,"title":"Evidence[^1]","body":[{"p":"Source note[^1]"}]}`)
	ctx := SceneContext{Surface: "light", Notes: []string{"Source"}}
	p, err := r.planSceneNode("example", raw, ctx)
	if err != nil {
		t.Fatal(err)
	}
	native := nativeProfileCardMulti(p, raw)
	if native == nil || len(native.Items) != 1 || native.Items[0].Text == nil {
		t.Fatal("plain native footnote paragraphs did not convert")
	}
	tr := *native.Items[0].Text
	if len(tr.Rich.Paragraphs) != 2 || len(tr.Rich.Paragraphs[0].Runs) < 2 || len(tr.Rich.Paragraphs[1].Runs) < 2 {
		t.Fatal("inline footnote runs were flattened")
	}
	serialized := string(editableCardParagraphXML(tr))
	if !strings.Contains(serialized, "baseline=") || !strings.Contains(serialized, "sz=\"840\"") {
		t.Fatal("footnote run baseline or reduced font size was not emitted", serialized)
	}
}

func TestNativeProfileCardBodyOnlyLabelAndOutline(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		raw        string
		paragraphs int
		bordered   bool
	}{
		{`{"type":"card","x":60,"y":100,"w":300,"h":180,"surface":"subtle","body":[{"p":"Body starts at the source content origin."},{"p":"Second body paragraph."}]}`, 2, false},
		{`{"type":"card","x":60,"y":100,"w":300,"h":180,"surface":"subtle","label":"Context","title":"Title","body":[{"p":"Body copy."}]}`, 3, false},
		{`{"type":"card","x":60,"y":100,"w":300,"h":180,"surface":"outline","body":[{"p":"Outlined body copy."}]}`, 1, true},
	} {
		r.editingProfile = ""
		before, err := r.planSceneNode("example", json.RawMessage(tc.raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		r.editingProfile = NativeEditingProfile
		after, err := r.planSceneNode("example", json.RawMessage(tc.raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		if len(after.Items) != 1 || after.Items[0].Text == nil || len(after.Items[0].Text.Rich.Paragraphs) != tc.paragraphs {
			t.Fatalf("native card not converted: %s: %v", tc.raw, after.Warnings)
		}
		native := after.Items[0].Text
		if native.NativeShape == nil || native.NativeShape.Rect != before.Items[0].Shape.Record.Rect || native.NativeShape.Fill != before.Items[0].Shape.Props.Fill.Color {
			t.Fatal("card outer geometry or fill changed", tc.raw)
		}
		if tc.bordered {
			if !reflect.DeepEqual(native.NativeShape.Line, before.Items[0].Shape.Props.Line) {
				t.Fatal("outline border changed", native.NativeShape.Line, before.Items[0].Shape.Props.Line)
			}
			got, want := nativeTextShapeBounds(*native.NativeShape), before.Items[0].Shape.Props.PositionProps
			if math.Abs(got.X/72-want.X.Val) > .0001 || math.Abs(got.Y/72-want.Y.Val) > .0001 || math.Abs(got.W/72-want.W.Val) > .0001 || math.Abs(got.H/72-want.H.Val) > .0001 {
				t.Fatal("outline path bounds differ from source centered-stroke geometry", got, want)
			}
		} else if native.NativeShape.Line != nil {
			t.Fatal("unexpected border on filled card")
		}
	}
}

func TestNativeProfileOutlinedCardDrawingML(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"type":"card","x":60,"y":100,"w":300,"h":180,"surface":"outline","body":[{"p":"Outlined body copy."}]}`)
	pres := pptx.New()
	rslide := pres.AddSlide()
	r.slide = rslide
	textRecords := []TextRecord{}
	r.records = &textRecords
	sourcePlan, err := r.planSceneNode("outline", raw, SceneContext{Surface: "light"})
	if err != nil {
		t.Fatal(err)
	}
	r.editingProfile = NativeEditingProfile
	native, err := r.planSceneNode("outline", raw, SceneContext{Surface: "light"})
	if err != nil || len(native.Items) != 1 || native.Items[0].Text == nil {
		t.Fatal("outline card did not convert", err)
	}
	if err = r.drawScene(native, &SlideReport{}, "/body/0"); err != nil {
		t.Fatal(err)
	}
	objects := rslide.PresSlide().SlideObjects
	if len(objects) != 1 || objects[0].Options == nil || objects[0].Options.Line == nil {
		t.Fatalf("outline was not emitted on one native text shape: %+v", objects)
	}
	if !reflect.DeepEqual(objects[0].Options.Line, sourcePlan.Items[0].Shape.Props.Line) {
		t.Fatal("emitted text shape line differs from source outline")
	}
	shapeBounds := nativeTextShapeBounds(*native.Items[0].Text.NativeShape)
	if math.Abs(objects[0].Options.X.Val-shapeBounds.X/72) > .0001 || math.Abs(objects[0].Options.Y.Val-shapeBounds.Y/72) > .0001 || math.Abs(objects[0].Options.W.Val-shapeBounds.W/72) > .0001 || math.Abs(objects[0].Options.H.Val-shapeBounds.H/72) > .0001 {
		t.Fatal("native outline path has altered source bounds", objects[0].Options.PositionProps)
	}
	var archive bytes.Buffer
	if _, err = pres.WriteTo(&archive); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(archive.Bytes()), int64(archive.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var slideXML []byte
	for _, file := range zr.File {
		if file.Name == "ppt/slides/slide1.xml" {
			reader, openErr := file.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			slideXML, err = io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if len(slideXML) == 0 || xml.Unmarshal(slideXML, new(any)) != nil {
		t.Fatal("slide XML is absent or invalid")
	}
	xmlText := string(slideXML)
	for _, want := range []string{`<a:ln w="12700">`, `val="CED7E6"`, `Outlined body copy.`, `typeface="IBM Plex Sans"`, `lIns="222250"`, `tIns="222250"`} {
		if !strings.Contains(xmlText, want) {
			t.Fatalf("outlined native card XML missing %q", want)
		}
	}
}

func TestNativeProfileCardMultiRejectsNonPlainOrUnsafeCards(t *testing.T) {
	r := intakeTestRenderer(t)
	r.source = densityTestSource(t)
	var err error
	r.typeEngine, err = NewSourceTypographyEngine(r.source, filepath.Join(densityTestBundle(), "fonts"), CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{
		`{"type":"card","x":60,"y":100,"w":360,"h":230,"title":"Title","body":[{"p":"Plain"},{"label":"Not a paragraph"}]}`,
		`{"type":"card","x":60,"y":100,"w":360,"h":230,"title":"Title","body":[{"p":"Contains\nnewline"},{"p":"Plain"}]}`,
	} {
		p, err := r.planSceneNode("example", json.RawMessage(raw), SceneContext{Surface: "light"})
		if err != nil {
			t.Fatal(err)
		}
		if nativeProfileCardMulti(p, json.RawMessage(raw)) != nil {
			t.Fatalf("unsafe card converted: %s", raw)
		}
	}
}
