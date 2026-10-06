package wmdesign

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const syntheticPhotoSidecar = "# Existing photo title\n\nTwo colleagues compare a chart near a window. There is negative space on the right.\n\nThis photo supports accessible digital customer experiences.\n\n## Searchable metadata\n\n- **Title:** Existing photo title\n- **Alt text:** Two colleagues near a window.\n- **Keywords:** colleagues, customer experience, negative space, accessibility\n- **Style:** Industry; Authentic workplace\n- **Recommended West Monroe placement:** `Industry / Practices / Banking`\n- **Primary trending topic:** Digital inclusion\n- **Secondary topics:** Customer experience; Accessibility\n- **People:** Two adults visible\n- **Location:** Bright office\n- **Custom field:** Preserve me\n"

func syntheticPhotoRecord(t *testing.T, path string) (RegisteredPhoto, []byte) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 24, 16))
	img.Set(3, 3, color.RGBA{R: 120, B: 50, A: 255})
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, nil); err != nil {
		t.Fatal(err)
	}
	raw := encoded.Bytes()
	meta, err := ParsePhotoMetadata([]byte(syntheticPhotoSidecar), path+".md")
	if err != nil {
		t.Fatal(err)
	}
	return RegisteredPhoto{ID: StablePhotoID(path), Path: path, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), Bytes: int64(len(raw)), Width: 24, Height: 16, MIME: "image/jpeg", Metadata: meta}, raw
}

func withPhotoSnapshot(t *testing.T, records ...RegisteredPhoto) {
	t.Helper()
	old := photoRegistryMetadata
	photoRegistryMetadata = map[string]RegisteredPhoto{}
	for _, record := range records {
		photoRegistryMetadata[record.ID] = record
	}
	t.Cleanup(func() { photoRegistryMetadata = old })
}

func TestPhotoMetadataParserPreservesSidecarAndActualSemantics(t *testing.T) {
	meta, err := ParsePhotoMetadata([]byte(syntheticPhotoSidecar), "West Monroe Photos/example.jpg.md")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Text != syntheticPhotoSidecar || meta.Authority != PhotoMetadataAuthority || meta.Fields["Custom field"] != "Preserve me" {
		t.Fatal("source text, authority or unknown field lost")
	}
	if meta.Title != "Existing photo title" || !strings.Contains(meta.Caption, "negative space on the right") || meta.Placement != "Industry / Practices / Banking" || meta.PeopleDescription != "Two adults visible" {
		t.Fatalf("actual semantics missing: %+v", meta)
	}
	if meta.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(syntheticPhotoSidecar))) {
		t.Fatal("sidecar hash mismatch")
	}
	if !reflect.DeepEqual(meta.Topics, []string{"accessibility", "customer experience", "digital inclusion"}) {
		t.Fatal(meta.Topics)
	}
	if _, err = ParsePhotoMetadata([]byte(syntheticPhotoSidecar+"- **Title:** duplicate\n"), meta.Path); err == nil {
		t.Fatal("duplicate sidecar field accepted")
	}
}

func TestPhotoStableIDsAndSnapshotValidation(t *testing.T) {
	first, _ := syntheticPhotoRecord(t, "West Monroe Photos/Industry/Banking/photo.jpg")
	second := first
	second.Path = "West Monroe Photos/Industry/Healthcare/photo.jpg"
	second.ID = StablePhotoID(second.Path)
	second.Metadata.Path = second.Path + ".md"
	if first.ID == second.ID || first.ID != StablePhotoID(first.Path) {
		t.Fatal("path IDs unstable or collision")
	}
	snapshot := PhotoRegistrySnapshot{Schema: PhotoRegistrySchema, SourceDirectory: "West Monroe Photos", Photos: []RegisteredPhoto{first, second}}
	raw, _ := json.Marshal(snapshot)
	if _, err := decodePhotoRegistry(raw); err != nil {
		t.Fatal("identical bytes at two paths must remain distinct", err)
	}
	for _, suffix := range []string{"{}", "garbage"} {
		if _, err := decodePhotoRegistry(append(append([]byte(nil), raw...), []byte(suffix)...)); err == nil {
			t.Fatal("trailing JSON accepted")
		}
	}
	snapshot.Photos = append(snapshot.Photos, first)
	raw, _ = json.Marshal(snapshot)
	if _, err := decodePhotoRegistry(raw); err == nil {
		t.Fatal("duplicate path accepted")
	}
	snapshot.Photos = []RegisteredPhoto{first}
	snapshot.Photos[0].Metadata.Caption = "Invented caption"
	raw, _ = json.Marshal(snapshot)
	if _, err := decodePhotoRegistry(raw); err == nil {
		t.Fatal("sidecar projection tampering accepted")
	}
}

func TestPhotoRegistrationReproducibleAndPreservesCuratedID(t *testing.T) {
	record, raw := syntheticPhotoRecord(t, "West Monroe Photos/Practices/Banking/example.jpg")
	record.ID = "photo-working-session"
	withSyntheticAssetRegistry(t, map[string]primitiveAsset{record.ID: {Path: record.Path, SHA256: record.SHA256}}, map[string][]byte{record.Path: raw, record.Metadata.Path: []byte(record.Metadata.Text)})
	root := os.Getenv("WMDS_BRANDING_ROOT")
	var previous []byte
	for i := 0; i < 2; i++ {
		out := filepath.Join(t.TempDir(), "photos.json")
		report, err := RegisterPhotoLibrary(root, out)
		if err != nil {
			t.Fatal(err)
		}
		if report.Photos != 1 || report.PreservedIDs != 1 {
			t.Fatal(report)
		}
		encoded, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		if previous != nil && !bytes.Equal(previous, encoded) {
			t.Fatal("registration snapshot is not reproducible")
		}
		previous = encoded
		snapshot, err := decodePhotoRegistry(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Photos[0].ID != record.ID || snapshot.Photos[0].Width != 24 || snapshot.Photos[0].Height != 16 {
			t.Fatal(snapshot)
		}
		if _, err = RegisterPhotoLibrary(root, out); err == nil {
			t.Fatal("existing snapshot output overwritten")
		}
	}
	if err := os.WriteFile(filepath.Join(root, record.Path), append(raw, 1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := RegisterPhotoLibrary(root, filepath.Join(t.TempDir(), "photos.json")); err == nil {
		t.Fatal("curated image hash silently changed")
	}
}

func TestPhotoDiscoveryDoesNotReadOriginalsAndPreviewVerifiesSelectedFile(t *testing.T) {
	record, raw := syntheticPhotoRecord(t, "West Monroe Photos/Industry/Banking/missing.jpg")
	withSyntheticAssetRegistry(t, map[string]primitiveAsset{record.ID: {Path: record.Path, SHA256: record.SHA256}}, nil)
	withPhotoSnapshot(t, record)
	results, err := AssetSelections("negative space accessibility", "photo", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != record.Metadata.Title || results[0].People != "yes" || results[0].Orientation != "landscape" || results[0].Variants[0].ThumbnailState != "registered_original_not_verified_in_query" {
		t.Fatalf("snapshot discovery failed: %+v", results)
	}
	if _, _, err = PrimitiveAssetOriginal(record.ID); err == nil {
		t.Fatal("preview accepted missing original")
	}
	path := filepath.Join(os.Getenv("WMDS_BRANDING_ROOT"), record.Path)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err = PrimitiveAssetOriginal(record.ID); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, append(raw, 1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err = PrimitiveAssetOriginal(record.ID); err == nil {
		t.Fatal("preview accepted changed selected original")
	}
}

func TestPhotoRegistryFingerprintIncludesSidecarSemantics(t *testing.T) {
	record, _ := syntheticPhotoRecord(t, "West Monroe Photos/Industry/Banking/example.jpg")
	withSyntheticAssetRegistry(t, map[string]primitiveAsset{record.ID: {Path: record.Path, SHA256: record.SHA256}}, nil)
	withPhotoSnapshot(t, record)
	before := AssetRegistryFingerprint()
	changed := record
	changed.Metadata.Caption += " additional source context"
	photoRegistryMetadata[record.ID] = changed
	if before == AssetRegistryFingerprint() {
		t.Fatal("metadata changes left registry fingerprint unchanged")
	}
}

func TestPhotoRegistryCompleteEmbeddedCoverage(t *testing.T) {
	snapshot, err := decodePhotoRegistry(embeddedPhotoRegistry)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Photos) != 521 {
		t.Fatalf("embedded snapshot has%d photos; want521", len(snapshot.Photos))
	}
	if len(PrimitiveAssetCatalog()) != 1204 {
		t.Fatalf("registry variants%d; want1204", len(PrimitiveAssetCatalog()))
	}
	core := 0
	for _, photo := range snapshot.Photos {
		if strings.HasPrefix(photo.ID, "photo-") {
			core++
			curated := curatedAssetMetadata[photo.ID]
			meta := assetMetadataFor(photo.ID, photo.Path)
			if meta.name != curated.name || meta.description != curated.description {
				t.Fatal("curated name/description changed", photo.ID)
			}
		}
		if _, ok := primitiveAssetRegistry[photo.ID]; !ok {
			t.Fatal("photo not registered", photo.ID)
		}
	}
	if core != 20 {
		t.Fatalf("retained core photos%d; want20", core)
	}
}

func TestPhotoLibraryPrivateSourceCoverage(t *testing.T) {
	if testing.Short() || os.Getenv("PPTXGENGO_PRIVATE_PHOTO_TESTS") != "1" {
		t.Skip("set PPTXGENGO_PRIVATE_PHOTO_TESTS=1 without -short for private exhaustive registration")
	}
	root := os.Getenv("WMDS_BRANDING_ROOT")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		root = filepath.Join(home, "Documents/branding")
	}
	out := filepath.Join(t.TempDir(), "photos.json")
	report, err := RegisterPhotoLibrary(root, out)
	if err != nil {
		t.Fatal(err)
	}
	if report.Photos != 521 {
		t.Fatal(report)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, embeddedPhotoRegistry) {
		t.Fatal("embedded registry differs from current original/sidecar collection")
	}
}

func TestPhotoRegisteredRenderingAndOfflineProvenance(t *testing.T) {
	record, raw := syntheticPhotoRecord(t, "West Monroe Photos/Industry/Banking/example.jpg")
	withSyntheticAssetRegistry(t, map[string]primitiveAsset{record.ID: {Path: record.Path, SHA256: record.SHA256}}, map[string][]byte{record.Path: raw})
	withPhotoSnapshot(t, record)
	scene, _ := json.Marshal(map[string]any{"type": "imageframe", "photo": record.ID, "x": 72, "y": 126, "w": 240, "h": 180})
	doc := Document{Schema: "pptxgengo.wmds-foundation.v1", Year: 2026, BuildIdentity: &BuildIdentity{Timestamp: "2000-01-01T00:00:00Z", Seed: "registered-photo-pipeline"}, Slides: []SlideSpec{{ID: "new-photo", Frame: FrameRequest{NoHeader: true, NoPage: true}, Nodes: []Node{{ID: "photo", Kind: "scene", Scene: &SceneSpec{Node: scene}}}}}}
	deck, _, err := BuildWithEngine(filepath.Join("..", "..", "library/wm-design-system/v10"), "", doc, CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := UsedPrimitiveAssets(deck)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].Key != record.ID || refs[0].SHA256 != record.SHA256 || refs[0].Path != record.Path {
		t.Fatalf("offline packaging provenance lost dynamic registered ID: %+v", refs)
	}
}

func TestPhotoCompactSummaryPreservesSourceAndSemanticDetails(t *testing.T) {
	photo, _ := syntheticPhotoRecord(t, "West Monroe Photos/Industry/Banking/example.jpg")
	full := []AssetSelection{{ID: photo.ID, SourceMetadata: clonePhotoMetadata(photo.Metadata), OriginalFacts: photoOriginalFacts(photo)}}
	compact := CompactAssetSelections(full)
	raw, err := json.Marshal(compact)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"text":`) || strings.Contains(string(raw), `"fields":`) {
		t.Fatal("summary includes full raw Markdown/fields")
	}
	if full[0].SourceMetadata.Text != syntheticPhotoSidecar || len(full[0].SourceMetadata.Fields) == 0 {
		t.Fatal("summary mutated full source")
	}
	if compact[0].SourceMetadata.Caption != photo.Metadata.Caption || compact[0].SourceMetadata.SHA256 != photo.Metadata.SHA256 || len(compact[0].SourceMetadata.Keywords) == 0 || compact[0].OriginalFacts.Width != 24 {
		t.Fatal("summary lost semantic/source/dimension facts")
	}
}
