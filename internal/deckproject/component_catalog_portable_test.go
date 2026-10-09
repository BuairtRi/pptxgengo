package deckproject

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

// Opt-in checks the actual 29-family customized source deck, retaining every
// predecessor and migration receipt. No approval/native review is invented.
func TestFamilyCatalogPortableVersionsAndRelocatedZIP(t *testing.T) {
	source := os.Getenv("PPTXGENGO_FAMILY_PORTABLE_PROJECT")
	if source == "" {
		t.Skip("set explicit private combined family project to qualify portability")
	}
	owned := t.TempDir()
	if e := copyDemoProject(source, owned); e != nil {
		t.Fatal(e)
	}
	p, e := Load(owned)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Document.Slides) != 29 || len(p.Document.LocalTemplates) != 29 {
		t.Fatal("expected actual customized 29-family input")
	}
	b := journeyBundle(t)
	migration, e := Migrate(p, b, wmdesign.CandidateEngine, false)
	if e != nil {
		t.Fatal(e)
	}
	if migration.Status != "migrated" && migration.Status != "already_current" {
		t.Fatal("runtime pin migration not applied", migration)
	}
	if migration.Status == "migrated" {
		if _, e = os.Stat(filepath.Join(owned, migration.Backup)); e != nil {
			t.Fatal("old runtime pin not retained", e)
		}
	}
	if _, e = PortableLayout(p, true); e != nil {
		t.Fatal(e)
	}
	p, e = Load(owned)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.SlideFiles) != 29 || len(p.TemplateFiles) != 29 {
		t.Fatal("portable source folder layout incomplete")
	}
	if _, e = Build(p, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	one, e := SaveVersion(p, "Qualification operator", "First complete customized29-family source/deck snapshot; native review separate")
	if e != nil {
		t.Fatal(e)
	}
	two, e := SaveVersion(p, "Qualification operator", "Second complete snapshot reuses unchanged authored source and asset revisions")
	if e != nil {
		t.Fatal(e)
	}
	if one.Number != "000001" || two.Number != "000002" || two.Parent != one.Number {
		t.Fatalf("numbered lineage %+v %+v", one, two)
	}
	list, e := ListVersions(owned)
	if e != nil || list.Current.Number != two.Number || len(list.Versions) != 2 {
		t.Fatalf("current pointer %+v %v", list, e)
	}
	expected := map[string]bool{}
	for _, version := range []DeckVersion{one, two} {
		if _, e = VerifyVersion(owned, version.Number); e != nil {
			t.Fatal(e)
		}
		for _, asset := range version.Assets {
			expected[asset.Object] = true
		}
	}
	if len(expected) == 0 {
		t.Fatal("combined actual source must exercise explicit private placeholder payloads")
	}
	objects, e := os.ReadDir(filepath.Join(owned, "assets/objects/sha256"))
	if e != nil || len(objects) != len(expected) {
		t.Fatalf("asset dedup count %d expected%d %v", len(objects), len(expected), e)
	}
	archive := filepath.Join(t.TempDir(), "families29-private.zip")
	share, e := ShareProject(p, archive)
	if e != nil {
		t.Fatal(e)
	}
	if !share.ContainsPrivateMaterial || share.CurrentVersion != two.Number {
		t.Fatal("share metadata", share)
	}
	z, e := zip.OpenReader(archive)
	if e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "assets/") {
			if !expected[f.Name] || seen[f.Name] {
				t.Fatalf("duplicated/unexpected archive asset %s", f.Name)
			}
			seen[f.Name] = true
		}
		if strings.HasPrefix(f.Name, "versions/") && strings.Contains(f.Name, "/assets/") {
			t.Fatal("asset payload duplicated inside version", f.Name)
		}
	}
	z.Close()
	if len(seen) != len(expected) {
		t.Fatal("ZIP lost shared asset revisions")
	}
	extracted := filepath.Join(t.TempDir(), "extracted")
	if _, e = ExtractShare(archive, extracted); e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyShare(extracted); e != nil {
		t.Fatal(e)
	}
	for _, version := range []DeckVersion{one, two} {
		destination := filepath.Join(t.TempDir(), "restored")
		if _, e = MaterializeVersion(extracted, version.Number, destination); e != nil {
			t.Fatal(e)
		}
		restored, e := Load(destination)
		if e != nil {
			t.Fatal(e)
		}
		// Uses the genuine currently pinned test runtime. Neither a monkey-patched
		// executable hash nor a substituted lock is used after materialization.
		if _, e = Check(restored, b, wmdesign.CandidateEngine); e != nil {
			t.Fatal(e)
		}
		rebuilt, e := Build(restored, BuildOptions{Bundle: b, Engine: wmdesign.CandidateEngine})
		if e != nil {
			t.Fatal(e)
		}
		if rebuilt.Outputs["deck.pptx"] != version.DeckSHA256 {
			t.Fatal("relocated native deck byte drift", version.Number)
		}
	}
	receipt := map[string]any{"schema": "pptxgengo.family-portability-qualification.v1", "families": 29, "source_input": source, "source_sha256": p.SourceHash(), "runtime_migration": migration, "versions": []string{one.Number, two.Number}, "current": list.Current, "unique_shared_asset_objects": len(expected), "share_sha256": share.SHA256, "rebuild": "both materialized snapshots checked and rebuilt with genuine pinned runtime; identical native deck SHA256", "native": "not reviewed by this source/portability exercise"}
	if e = os.WriteFile(filepath.Join(owned, "family-portability-qualification.json"), canonical(receipt), 0600); e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(archive)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(owned, "families29-private.zip"), raw, 0600); e != nil {
		t.Fatal(e)
	}
	writeProcessJourneyQualification(t, p, "portable-families-29")
	t.Logf("29-family portability complete: 2 numbered source/deck snapshots, %d shared asset objects, exact pinned rebuilt bytes", len(expected))
}
