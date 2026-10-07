package deckproject

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestPortableLayoutPreservesCanonicalAndPreimages(t *testing.T) {
	p := example(t)
	before := p.SourceHash()
	sem := append([]byte{}, p.Canonical...)
	plan, e := PortableLayout(p, false)
	if e != nil {
		t.Fatal(e)
	}
	if plan.Applied || p.SourceHash() != before {
		t.Fatal("dry run changed source")
	}
	if _, e = os.Stat(filepath.Join(p.Root, "slides")); !os.IsNotExist(e) {
		t.Fatal("dry run wrote files")
	}
	if _, e = PortableLayout(p, true); e != nil {
		t.Fatal(e)
	}
	next, e := Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(next.Canonical, sem) {
		t.Fatal("semantic mutation")
	}
	for id, rel := range next.SlideFiles {
		if rel != "slides/"+id+".yaml" {
			t.Fatal(rel)
		}
	}
	for id, rel := range next.TemplateFiles {
		if rel != "slides/templates/"+id+".yaml" {
			t.Fatal(rel)
		}
	}
	if b, e := os.ReadFile(filepath.Join(p.Root, "decisions/sources", before, "deck.yaml")); e != nil || !bytes.Equal(b, p.Raw) {
		t.Fatalf("missing preimage %v", e)
	}
	if _, e = PortableLayout(next, true); e != nil {
		t.Fatal(e)
	}
}
func TestPortableLayoutRefusesOccupiedPath(t *testing.T) {
	p := example(t)
	path := filepath.Join(p.Root, "slides", p.Document.Slides[0].ID+".yaml")
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(path, []byte("colleague content"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := PortableLayout(p, false); e == nil {
		t.Fatal("occupied file accepted")
	}
	b, _ := os.ReadFile(path)
	if string(b) != "colleague content" {
		t.Fatal("overwritten")
	}
}
func TestPortableCompleteVersionsSharedAssetsAndRelocatedZIP(t *testing.T) {
	p := example(t)
	if _, e := PortableLayout(p, true); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.Root)
	pin(t, p)
	build := func() {
		t.Helper()
		if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
			t.Fatal(e)
		}
	}
	build()
	v1, e := SaveVersion(p, "Operator One", "first source and native deck")
	if e != nil {
		t.Fatal(e)
	}
	if v1.Number != "000001" {
		t.Fatal(v1.Number)
	}
	asset := p.Document.Assets["sample-image"]
	original, e := readProjectFile(p.Root, asset.Path)
	if e != nil {
		t.Fatal(e)
	} // PNG byte mutation that stays valid image content: append harmless bytes.
	changed := append(append([]byte{}, original...), []byte("second immutable revision")...)
	if _, e = RegisterAsset(p, AssetRegistration{ID: "sample-image", Data: changed, Description: "changed original", ReplaceSHA256: digest(original)}); e != nil {
		t.Fatal(e)
	}
	p, _ = Load(p.Root)
	build()
	v2, e := SaveVersion(p, "Operator Two", "new asset revision")
	if e != nil {
		t.Fatal(e)
	}
	if v2.Number != "000002" || v2.Parent != v1.Number {
		t.Fatal(v2)
	}
	list, e := ListVersions(p.Root)
	if e != nil || len(list.Versions) != 2 {
		t.Fatalf("%+v %v", list, e)
	}
	objects, e := os.ReadDir(filepath.Join(p.Root, "assets/objects/sha256"))
	if e != nil || len(objects) != 2 {
		t.Fatalf("asset objects: %v %v", objects, e)
	} // Unchanged snapshot reuses asset objects.
	if _, e = SaveVersion(p, "Operator Two", "same bytes retained"); e != nil {
		t.Fatal(e)
	}
	objects, _ = os.ReadDir(filepath.Join(p.Root, "assets/objects/sha256"))
	if len(objects) != 2 {
		t.Fatal("unchanged asset duplicated")
	}
	out := filepath.Join(t.TempDir(), "colleague.zip")
	receipt, e := ShareProject(p, out)
	if e != nil {
		t.Fatal(e)
	}
	if !receipt.ContainsPrivateMaterial {
		t.Fatal("private source mislabeled")
	}
	z, e := zip.OpenReader(out)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	extracted := filepath.Join(t.TempDir(), "extracted")
	storedAssets := map[string]bool{}
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "assets/") {
			if !strings.HasPrefix(f.Name, "assets/objects/sha256/") {
				t.Fatalf("legacy bytes duplicated in ZIP: %s", f.Name)
			}
			if storedAssets[f.Name] {
				t.Fatal("duplicate object")
			}
			storedAssets[f.Name] = true
		}
	}
	if len(storedAssets) != 2 {
		t.Fatalf("ZIP contains %d asset objects", len(storedAssets))
	}
	if _, e = ExtractShare(out, extracted); e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyShare(extracted); e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{"000001", "000002", "000003"} {
		dest := filepath.Join(t.TempDir(), "restored")
		v, e := MaterializeVersion(extracted, n, dest)
		if e != nil {
			t.Fatal(e)
		}
		restored, e := Load(dest)
		if e != nil {
			t.Fatal(e)
		}
		if restored.SourceHash() != v.SourceSHA256 {
			t.Fatal("relocation source drift")
		}
		if _, e = Check(restored, bundle(t), wmdesign.CandidateEngine); e != nil {
			t.Fatal(e)
		}
		rebuilt, e := Build(restored, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
		if e != nil {
			t.Fatal(e)
		}
		if rebuilt.Outputs["deck.pptx"] != v.DeckSHA256 {
			t.Fatalf("relocated version%s rebuilt different native deck", n)
		}
	} // Neither the old source nor native deck was replaced.
	if _, e = VerifyVersion(p.Root, "000001"); e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyVersion(p.Root, "000002"); e != nil {
		t.Fatal(e)
	}
}
func TestPortableVersionRejectsDriftConflictAndBusy(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e := SaveVersion(p, "Operator", "first"); e != nil {
		t.Fatal(e)
	}
	lock := filepath.Join(p.Root, ".project-version.lock")
	if e := os.WriteFile(lock, []byte("another writer"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := SaveVersion(p, "Operator", "second"); e == nil {
		t.Fatal("busy lock taken over")
	}
	os.Remove(lock)
	conflict := filepath.Join(p.Root, "deck (conflicted copy).yaml")
	os.WriteFile(conflict, []byte("colleague"), 0644)
	if _, e := ListVersions(p.Root); e == nil {
		t.Fatal("conflict ignored")
	}
	os.Remove(conflict)
	pointer := filepath.Join(p.Root, "versions/current.json")
	b, _ := os.ReadFile(pointer)
	os.WriteFile(pointer, bytes.Replace(b, []byte("000001"), []byte("000002"), 1), 0644)
	if _, e := ListVersions(p.Root); e == nil {
		t.Fatal("divergent pointer accepted")
	}
	os.WriteFile(pointer, b, 0644)
	asset := p.Document.Assets["sample-image"]
	raw, _ := readProjectFile(p.Root, asset.Path)
	object := filepath.Join(p.Root, "assets/objects/sha256", digest(raw))
	os.Chmod(object, 0644)
	os.WriteFile(object, []byte("tampered"), 0644)
	if _, e := VerifyVersion(p.Root, "000001"); e == nil {
		t.Fatal("tampered asset accepted")
	}
}

func TestPortableCreateUsesStableLayoutAndExactPins(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project with spaces")
	p, e := CreateProject(CreateOptions{Out: root, ID: "stable-deck", Title: "New generic deck", Year: 2026, Bundle: bundle(t), Template: "cards/3", Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	if p.SlideFiles["first-slide"] != "slides/first-slide.yaml" {
		t.Fatal(p.SlideFiles)
	}
	if _, _, e = ReadLock(p); e != nil {
		t.Fatal(e)
	}
	if _, e = Check(p, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
	if _, e = CreateProject(CreateOptions{Out: root, ID: "stable-deck", Title: "Overwrite", Year: 2026, Bundle: bundle(t), Template: "cards/3"}); e == nil {
		t.Fatal("existing project overwritten")
	}
	p, e = Load(root)
	if e != nil || p.Document.Title != "New generic deck" {
		t.Fatal(e)
	}
}
func TestPortableExplicitRecoveryRetainsCompleteSnapshot(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e := SaveVersion(p, "Operator", "first"); e != nil {
		t.Fatal(e)
	}
	ptr := filepath.Join(p.Root, "versions/current.json")
	before, e := os.ReadFile(ptr)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = SaveVersion(p, "Operator", "second"); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(ptr, before, 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = ListVersions(p.Root); e == nil {
		t.Fatal("interrupted publication accepted")
	}
	if _, e = RecoverVersion(p.Root, "000002", strings.Repeat("0", 64)); e == nil {
		t.Fatal("wrong pointer preimage accepted")
	}
	if _, e = RecoverVersion(p.Root, "000002", digest(before)); e != nil {
		t.Fatal(e)
	}
	if list, e := ListVersions(p.Root); e != nil || list.Current.Number != "000002" {
		t.Fatalf("%+v %v", list, e)
	}
}
func TestPortableRefusesStaleSourceAndSyncPlaceholders(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	old := p.Raw
	if e := os.WriteFile(p.SourcePath, append(append([]byte{}, old...), []byte("\n# colleague edit\n")...), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := SaveVersion(p, "Operator", "stale"); e == nil {
		t.Fatal("stale loaded source snapshot accepted")
	}
	os.WriteFile(p.SourcePath, old, 0644)
	placeholder := filepath.Join(p.Root, "source.pdf.icloud")
	os.WriteFile(placeholder, []byte("pending sync"), 0644)
	if _, e := SaveVersion(p, "Operator", "unsynced"); e == nil {
		t.Fatal("sync placeholder ignored")
	}
	os.Remove(placeholder)
	if _, e := SaveVersion(p, "Operator", "valid"); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(p.Root, "case.txt"), []byte("one"), 0644)
	os.WriteFile(filepath.Join(p.Root, "CASE.txt"), []byte("two"), 0644)
	entries, _ := os.ReadDir(p.Root)
	caseCount := 0
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), "case.txt") {
			caseCount++
		}
	}
	if caseCount == 2 {
		if _, e := ShareProject(p, filepath.Join(t.TempDir(), "share.zip")); e == nil {
			t.Fatal("case conflict accepted")
		}
	} else {
		t.Log("Filesystem prevents distinct case-colliding paths; extractor rejection is covered independently")
	}
}
func TestPortablePathsAndExtractionRejectUnsafeInputs(t *testing.T) {
	for _, rel := range []string{"../escape", "a/../b", "a\\b", "CON.txt", "folder/file.", "folder/with:colon", "LPT1/image.png"} {
		if e := portableName(rel); e == nil {
			t.Errorf("accepted %q", rel)
		}
	}
	path := filepath.Join(t.TempDir(), "unsafe.zip")
	f, e := os.Create(path)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	w, _ := z.Create("../escape")
	w.Write([]byte("no"))
	z.Close()
	f.Close()
	out := filepath.Join(t.TempDir(), "new")
	if _, e = ExtractShare(path, out); e == nil {
		t.Fatal("archive traversal accepted")
	}
	if _, e = os.Stat(out); !os.IsNotExist(e) {
		t.Fatal("unsafe output created")
	}
}

func TestPortableFailedAssetMutationRetainsDeduplicatedObject(t *testing.T) {
	p := example(t)
	data, e := readProjectFile(p.Root, p.Document.Assets["sample-image"].Path)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p.SourcePath, append(append([]byte{}, p.Raw...), []byte("\n# Concurrent source edit\n")...), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = RegisterAsset(p, AssetRegistration{ID: "stale-add", Data: data, Description: "stale source"}); e == nil {
		t.Fatal("stale source mutation accepted")
	}
	object := filepath.Join(p.Root, "assets/objects/sha256", digest(data))
	if raw, e := os.ReadFile(object); e != nil || !bytes.Equal(raw, data) {
		t.Fatalf("deduplicated object removed on source failure: %v", e)
	}
	current, e := Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = RegisterAsset(current, AssetRegistration{ID: "current-add", Data: data, Description: "current source"}); e != nil {
		t.Fatal(e)
	}
}

func TestPortableLegacyAssetOutsideAssetsStillReconstructs(t *testing.T) {
	p := example(t)
	old := p.Document.Assets["sample-image"].Path
	data, e := readProjectFile(p.Root, old)
	if e != nil {
		t.Fatal(e)
	}
	legacy := "legacy-photos/reference.png"
	if e = writeExclusive(filepath.Join(p.Root, filepath.FromSlash(legacy)), data, 0644); e != nil {
		t.Fatal(e)
	}
	raw := bytes.Replace(p.Raw, []byte(old), []byte(legacy), 1)
	if e = os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	pin(t, p)
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	v, e := SaveVersion(p, "Operator", "legacy source path")
	if e != nil {
		t.Fatal(e)
	}
	if v.Files[legacy] != "" || v.Assets[legacy].SHA256 != digest(data) {
		t.Fatal("legacy asset not shared independently")
	}
	zipPath := filepath.Join(t.TempDir(), "share.zip")
	if _, e = ShareProject(p, zipPath); e != nil {
		t.Fatal(e)
	}
	extracted := filepath.Join(t.TempDir(), "extracted")
	if _, e = ExtractShare(zipPath, extracted); e != nil {
		t.Fatal(e)
	}
	restored := filepath.Join(t.TempDir(), "restored")
	if _, e = MaterializeVersion(extracted, "000001", restored); e != nil {
		t.Fatal(e)
	}
	out, e := Load(restored)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Check(out, bundle(t), wmdesign.CandidateEngine); e != nil {
		t.Fatal(e)
	}
}

func TestPortableShareLogicalExpansionIsBoundedAndClosed(t *testing.T) {
	sha := strings.Repeat("a", 64)
	object := "assets/objects/sha256/" + sha
	r := ShareReceipt{Files: map[string]string{object: sha}, AssetAliases: map[string]VersionAsset{}}
	sizes := map[string]uint64{object: uint64(portableFileLimit)}
	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("assets/legacy-%02d.png", i)
		r.Files[name] = sha
		r.AssetAliases[name] = VersionAsset{Object: object, SHA256: sha}
	}
	if e := validateShareInventory(r, sizes); e == nil {
		t.Fatal("asset aliases exceeded expanded-byte budget")
	}
	for _, paths := range [][]string{{"a", "a-other", "a/file"}, {"FILE", "file"}, {"share-manifest.json/hidden"}} {
		r := ShareReceipt{Files: map[string]string{}}
		sizes := map[string]uint64{}
		for _, p := range paths {
			r.Files[p] = sha
			sizes[p] = 1
		}
		if e := validateShareInventory(r, sizes); e == nil {
			t.Fatalf("logical path collision accepted: %v", paths)
		}
	}
}
func TestPortableVersionRejectsUnlistedSnapshotFiles(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	if _, e := SaveVersion(p, "Operator", "first"); e != nil {
		t.Fatal(e)
	}
	for _, relative := range []string{"versions/000001/source/unlisted.txt", "versions/000001/unlisted.txt"} {
		path := filepath.Join(p.Root, filepath.FromSlash(relative))
		if e := os.WriteFile(path, []byte("sync conflict"), 0644); e != nil {
			t.Fatal(e)
		}
		if _, e := VerifyVersion(p.Root, "000001"); e == nil {
			t.Fatal("unlisted immutable snapshot file accepted")
		}
		os.Remove(path)
	}
}
func TestPortableProducerRefusesOversizedFilesBeforeReading(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "oversized.bin")
	f, e := os.Create(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Truncate(portableFileLimit + 1); e != nil {
		f.Close()
		t.Skipf("sparse oversized fixture unsupported: %v", e)
	}
	f.Close()
	if _, e := projectInventory(root, nil); e == nil {
		t.Fatal("oversized producer file accepted")
	}
	if _, e := readProjectFile(root, "oversized.bin"); e == nil {
		t.Fatal("unbounded direct project read accepted")
	}
	if _, e := readProjectFileLimit(root, "oversized.bin", portableManifestLimit); e == nil {
		t.Fatal("oversized manifest read accepted")
	}
}

func TestPortableOptionalAndPointerReadsAreBounded(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "oversized-pointer.json")
	f, e := os.Create(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Truncate(int64(portableManifestLimit) + 1); e != nil {
		f.Close()
		t.Skipf("oversized fixture unsupported: %v", e)
	}
	f.Close()
	if _, e := readOptionalLimit(path, portableManifestLimit); e == nil {
		t.Fatal("unbounded pointer read accepted")
	}
	if b, e := readOptional(filepath.Join(root, "missing")); e != nil || b != nil {
		t.Fatalf("optional absent read changed: %v", e)
	}
	empty := filepath.Join(root, "empty")
	if e := os.WriteFile(empty, nil, 0600); e != nil {
		t.Fatal(e)
	}
	if b, e := readOptional(empty); e != nil || b == nil || len(b) != 0 {
		t.Fatalf("existing empty predecessor treated as absent: %v", e)
	}
	if _, e := readOptional(root); e == nil {
		t.Fatal("optional directory accepted")
	}
}

func TestPortableVersionsRetainApprovalHistoryWithoutReapprovingChanges(t *testing.T) {
	p := example(t)
	pin(t, p)
	build := func() {
		t.Helper()
		if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
			t.Fatal(e)
		}
	}
	build()
	approval, e := Approve(p, "review", "Synthetic fixture reviewer", nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = SaveVersion(p, "Snapshot author", "approved fixture history"); e != nil {
		t.Fatal(e)
	}
	original, e := readProjectFile(p.Root, p.Document.Assets["sample-image"].Path)
	if e != nil {
		t.Fatal(e)
	}
	changed := append(append([]byte{}, original...), []byte("changed fixture revision")...)
	if _, e = RegisterAsset(p, AssetRegistration{ID: "sample-image", Data: changed, Description: "changed fixture", ReplaceSHA256: digest(original)}); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	build()
	if _, e = SaveVersion(p, "Snapshot author", "changed image is not reapproved"); e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		number string
		valid  bool
	}{{"000001", true}, {"000002", false}} {
		out := filepath.Join(t.TempDir(), "version")
		if _, e = MaterializeVersion(p.Root, tc.number, out); e != nil {
			t.Fatal(e)
		}
		restored, e := Load(out)
		if e != nil {
			t.Fatal(e)
		}
		state, e := Status(restored)
		if e != nil {
			t.Fatal(e)
		}
		found := false
		for _, a := range state.Approvals {
			if a.ID == approval.ID {
				found = true
				if a.Valid != tc.valid {
					t.Fatalf("version%s approval valid=%v, expected%v", tc.number, a.Valid, tc.valid)
				}
			}
		}
		if !found {
			t.Fatal("approval history lost")
		}
	}
}

func TestPortableRetainedModesKeepBaselinesAndObjectsImmutable(t *testing.T) {
	for _, path := range []string{"builds/build-id/deck.pptx", "assets/objects/sha256/abc", "versions/000001/source/deck.yaml", "versions/000001/deck.pptx", "decisions/sources/abc/deck.yaml"} {
		if portableFileMode(path) != 0444 {
			t.Fatal(path)
		}
	}
	for _, path := range []string{"deck.yaml", "slides/stable-id.yaml", "versions/current.json", "assets/legacy.png"} {
		if portableFileMode(path) != 0644 {
			t.Fatal(path)
		}
	}
}

func TestPortableLayoutRejectsWindowsReservedSlideFilenameBeforeMutation(t *testing.T) {
	p := example(t)
	raw := bytes.Replace(p.Raw, []byte("id: maintain-the-source"), []byte("id: CON"), 1)
	if e := os.WriteFile(p.SourcePath, raw, 0644); e != nil {
		t.Fatal(e)
	}
	p, e := Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	before := p.SourceHash()
	if _, e := PortableLayout(p, false); e == nil {
		t.Fatal("reserved destination accepted by dry run")
	}
	if _, e := Split(p, SplitOptions{}); e == nil {
		t.Fatal("reserved source filename emitted")
	}
	current, e := Load(p.Root)
	if e != nil || current.SourceHash() != before {
		t.Fatal("failed portable path check changed authored source")
	}
}

func TestPortableMaterializationReadRejectsChangedEditorialEvidence(t *testing.T) {
	p := example(t)
	pin(t, p)
	if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	v, e := SaveVersion(p, "Operator", "retained source and evidence")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = VerifyVersion(p.Root, v.Number); e != nil {
		t.Fatal(e)
	}
	relative := "versions/" + v.Number + "/source/context/project.md"
	path := filepath.Join(p.Root, filepath.FromSlash(relative))
	if e = os.Chmod(path, 0644); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, []byte("concurrent changed editorial evidence"), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e = readVersionInput(p.Root, relative, v.Files["context/project.md"]); e == nil {
		t.Fatal("changed editorial evidence accepted after verification")
	}
}
