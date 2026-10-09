package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Imported native evidence uses the same deck-wide immutable asset pool as
// authored graphics. Complete versions retain decision-history snapshots, but
// never copy the original native package into those snapshots.
func TestNativeImportPortableVersionsShareAndRebuild(t *testing.T) {
	p, _ := compositionTransactionFixture(t)
	source := nativeImportPackage(t, nativeImportPlainShape)
	mapping := nativeImportMapping(p, source, "portable-import")
	applied, err := ImportNative(p, "local-composition", source, mapping, bundle(t), wmdesign.CandidateEngine, true)
	if err != nil || !applied.Applied {
		t.Fatal(applied, err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); err != nil {
		t.Fatal(err)
	}
	sourceHash := p.SourceHash()
	originalCompiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	originalInspection, err := InspectDiagram(p, "local-composition", bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	asset := "assets/objects/sha256/" + digest(source)
	decision, err := os.ReadFile(filepath.Join(p.Root, applied.Decision))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Evidence nativeImportEvidence `json:"evidence"`
	}
	if err = json.Unmarshal(decision, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Evidence.SourceAsset != asset || receipt.Evidence.SourcePPTXSHA256 != digest(source) || !reflect.DeepEqual(receipt.Evidence.Mapping, mapping) {
		t.Fatal("global receipt lost exact reviewed mapping and raw-source address")
	}

	versions := []DeckVersion{}
	for _, message := range []string{"Imported text source and deck", "Unchanged imported evidence retained"} {
		v, err := SaveVersion(p, "Portability operator", message)
		if err != nil {
			t.Fatal(err)
		}
		versions = append(versions, v)
		if v.SourceSHA256 != sourceHash || v.Files[applied.Decision] != digest(decision) {
			t.Fatal("version lost source or import decision", v)
		}
		if want := v.Assets[asset]; want.Object != asset || want.SHA256 != digest(source) {
			t.Fatal("version does not reference the shared original native package", want)
		}
		stored, err := os.ReadFile(filepath.Join(p.Root, "versions", v.Number, "source", applied.Decision))
		if err != nil || !bytes.Equal(stored, decision) {
			t.Fatal("complete decision history changed", err)
		}
	}
	snapshot := compositionFileSnapshot(t, p.Root)
	rawCopies := []string{}
	for path, data := range snapshot {
		if bytes.Equal(data, source) {
			rawCopies = append(rawCopies, path)
		}
	}
	if len(rawCopies) != 1 || rawCopies[0] != asset {
		t.Fatalf("native package copied per version or elsewhere: %v", rawCopies)
	}

	archive := filepath.Join(t.TempDir(), "private-project.zip")
	share, err := ShareProject(p, archive)
	if err != nil {
		t.Fatal(err)
	}
	if !share.ContainsPrivateMaterial || share.Files[asset] != digest(source) || share.Files[applied.Decision] != digest(decision) {
		t.Fatal("private share omitted import provenance", share)
	}
	zr, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	rawCopies = nil
	for _, entry := range zr.File {
		if strings.HasPrefix(entry.Name, "versions/") && strings.HasSuffix(entry.Name, "/"+asset) {
			t.Fatal("raw source asset copied into a version", entry.Name)
		}
		r, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(r)
		closeErr := r.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
		if bytes.Equal(data, source) {
			rawCopies = append(rawCopies, entry.Name)
		}
	}
	if len(rawCopies) != 1 || rawCopies[0] != asset {
		t.Fatalf("shared archive duplicated original native package: %v", rawCopies)
	}

	relocated := filepath.Join(t.TempDir(), "colleague-project")
	if _, err = ExtractShare(archive, relocated); err != nil {
		t.Fatal(err)
	}
	if _, err = VerifyShare(relocated); err != nil {
		t.Fatal(err)
	}
	extracted, err := Load(relocated)
	if err != nil {
		t.Fatal(err)
	}
	if extracted.SourceHash() != sourceHash || !bytes.Equal(extracted.Canonical, p.Canonical) {
		t.Fatal("source semantics changed during share/extraction")
	}
	for _, v := range versions {
		decisionCopy, err := os.ReadFile(filepath.Join(relocated, "versions", v.Number, "source", applied.Decision))
		if err != nil || !bytes.Equal(decisionCopy, decision) {
			t.Fatal("extracted version lost exact decision snapshot", err)
		}
	}
	sourceCopy, err := os.ReadFile(filepath.Join(relocated, asset))
	if err != nil || !bytes.Equal(sourceCopy, source) {
		t.Fatal("extracted original native package changed", err)
	}
	globalDecision, err := os.ReadFile(filepath.Join(relocated, applied.Decision))
	if err != nil || !bytes.Equal(globalDecision, decision) {
		t.Fatal("extracted global decision changed", err)
	}
	importedTemplate := extracted.Document.LocalTemplates["editorial-photo"]
	nodes, index := findDiagramNode(&importedTemplate.Nodes, "portable-import")
	if nodes == nil || (*nodes)[index].Text != "Imported source text" {
		t.Fatal("portable import no longer has its authored plain-text node")
	}
	rebuiltCompiled, err := Compile(extracted, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(originalCompiled.Document, rebuiltCompiled.Document) {
		t.Fatal("relocated compiled deck semantics changed")
	}
	rebuiltInspection, err := InspectDiagram(extracted, "local-composition", bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(originalInspection.FinalNative, rebuiltInspection.FinalNative) || !reflect.DeepEqual(originalInspection.MeasuredText, rebuiltInspection.MeasuredText) || !reflect.DeepEqual(originalInspection.Frame, rebuiltInspection.Frame) {
		t.Fatal("imported deck text, frame or native geometry changed after extraction")
	}
	rebuilt, err := Build(extracted, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if err != nil {
		t.Fatal(err)
	}
	if rebuilt.SourceSHA256 != sourceHash || rebuilt.SemanticSHA256 != digest(extracted.Canonical) {
		t.Fatal("rebuilt deck receipt differs from portable authored semantics")
	}
}
