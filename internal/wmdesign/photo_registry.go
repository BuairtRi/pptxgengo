package wmdesign

import (
	"bufio"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

const PhotoRegistrySchema = "pptxgengo.photo-registry.v1"
const PhotoMetadataAuthority = "existing-branding-sidecar"

//go:embed photo_registry.json
var embeddedPhotoRegistry []byte

// PhotoMetadata preserves the existing descriptive sidecar exactly. These
// descriptions are source material, not a new visual review by this program.
type PhotoMetadata struct {
	Authority         string            `json:"authority"`
	Path              string            `json:"path"`
	SHA256            string            `json:"sha256"`
	Text              string            `json:"text,omitempty"`
	Fields            map[string]string `json:"fields,omitempty"`
	Title             string            `json:"title"`
	Caption           string            `json:"caption"`
	AltText           string            `json:"alt_text,omitempty"`
	Keywords          []string          `json:"keywords"`
	Topics            []string          `json:"topics,omitempty"`
	Style             string            `json:"style,omitempty"`
	Placement         string            `json:"placement,omitempty"`
	Practice          string            `json:"practice,omitempty"`
	PeopleDescription string            `json:"people_description,omitempty"`
	Location          string            `json:"location,omitempty"`
}

// RegisteredPhoto pins a full original and its descriptive source. Dimensions
// are read from the actual JPEG header during registration, without a decode.
type RegisteredPhoto struct {
	ID       string        `json:"id"`
	Path     string        `json:"path"`
	SHA256   string        `json:"sha256"`
	Bytes    int64         `json:"bytes"`
	Width    int           `json:"width_px"`
	Height   int           `json:"height_px"`
	MIME     string        `json:"mime"`
	Metadata PhotoMetadata `json:"metadata"`
}

// PhotoOriginalFacts are verified at snapshot registration. Search serves
// these immutable facts without claiming that a local file was read again.
type PhotoOriginalFacts struct {
	Bytes  int64  `json:"bytes"`
	Width  int    `json:"width_px"`
	Height int    `json:"height_px"`
	MIME   string `json:"mime"`
	Basis  string `json:"basis"`
}

func photoOriginalFacts(photo RegisteredPhoto) *PhotoOriginalFacts {
	return &PhotoOriginalFacts{photo.Bytes, photo.Width, photo.Height, photo.MIME, "original_sha256_and_jpeg_header_verified_at_registration"}
}

type PhotoRegistrySnapshot struct {
	Schema          string            `json:"schema"`
	SourceDirectory string            `json:"source_directory"`
	Photos          []RegisteredPhoto `json:"photos"`
}

type PhotoRegistrationReport struct {
	Schema         string `json:"schema"`
	Photos         int    `json:"photos"`
	PreservedIDs   int    `json:"preserved_registered_ids"`
	Bytes          int64  `json:"original_bytes"`
	SnapshotSHA256 string `json:"snapshot_sha256"`
}

var photoRegistryMetadata map[string]RegisteredPhoto

func init() {
	snapshot, err := decodePhotoRegistry(embeddedPhotoRegistry)
	if err != nil {
		panic(fmt.Errorf("embedded photo registry: %w", err))
	}
	photoRegistryMetadata, err = mergePhotoRegistry(primitiveAssetRegistry, snapshot)
	if err != nil {
		panic(fmt.Errorf("embedded photo registry: %w", err))
	}
}

// StablePhotoID uses the case-preserving, root-relative source path. Content
// updates keep the ID; two paths with identical image bytes remain discoverable.
func StablePhotoID(relative string) string {
	digest := sha256.Sum256([]byte(filepath.ToSlash(filepath.Clean(relative))))
	return "photo/library/" + fmt.Sprintf("%x", digest[:12])
}

func validPhotoHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func decodePhotoRegistry(raw []byte) (PhotoRegistrySnapshot, error) {
	var snapshot PhotoRegistrySnapshot
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return snapshot, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return snapshot, fmt.Errorf("photo.trailing_snapshot_json")
	}
	if snapshot.Schema != PhotoRegistrySchema || snapshot.SourceDirectory != "West Monroe Photos" {
		return snapshot, fmt.Errorf("photo.invalid_snapshot_schema_or_scope")
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	for _, photo := range snapshot.Photos {
		ext := strings.ToLower(filepath.Ext(photo.Path))
		if !filepath.IsLocal(photo.Path) || filepath.ToSlash(filepath.Clean(photo.Path)) != photo.Path || !strings.HasPrefix(photo.Path, "West Monroe Photos/") || (ext != ".jpg" && ext != ".jpeg") {
			return snapshot, fmt.Errorf("photo.invalid_original_path: %s", photo.Path)
		}
		if photo.ID == "" || ids[photo.ID] || paths[photo.Path] {
			return snapshot, fmt.Errorf("photo.duplicate_identity_or_path: %s", photo.Path)
		}
		if !(strings.HasPrefix(photo.ID, "photo-") || photo.ID == StablePhotoID(photo.Path)) {
			return snapshot, fmt.Errorf("photo.invalid_stable_id: %s", photo.ID)
		}
		ids[photo.ID], paths[photo.Path] = true, true
		if !validPhotoHash(photo.SHA256) || photo.Bytes <= 0 || photo.Width <= 0 || photo.Height <= 0 || photo.MIME != "image/jpeg" {
			return snapshot, fmt.Errorf("photo.invalid_original_facts: %s", photo.ID)
		}
		meta := photo.Metadata
		if meta.Authority != PhotoMetadataAuthority || meta.Path != photo.Path+".md" || !validPhotoHash(meta.SHA256) || meta.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(meta.Text))) {
			return snapshot, fmt.Errorf("photo.invalid_sidecar_provenance: %s", photo.ID)
		}
		parsed, err := ParsePhotoMetadata([]byte(meta.Text), meta.Path)
		if err != nil {
			return snapshot, err
		}
		if !reflectPhotoMetadataEqual(parsed, meta) {
			return snapshot, fmt.Errorf("photo.sidecar_projection_drift: %s", photo.ID)
		}
	}
	return snapshot, nil
}

func reflectPhotoMetadataEqual(a, b PhotoMetadata) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return string(left) == string(right)
}

func mergePhotoRegistry(registry map[string]primitiveAsset, snapshot PhotoRegistrySnapshot) (map[string]RegisteredPhoto, error) {
	metadata := map[string]RegisteredPhoto{}
	for _, photo := range snapshot.Photos {
		existing, exists := registry[photo.ID]
		if exists && (existing.Path != photo.Path || existing.SHA256 != photo.SHA256 || existing.Crop != [4]int{}) {
			return nil, fmt.Errorf("photo.existing_registration_drift: %s", photo.ID)
		}
		if !exists {
			for key, asset := range registry {
				if asset.Path == photo.Path {
					return nil, fmt.Errorf("photo.existing_path_id_changed: %s: %s", photo.ID, key)
				}
			}
		}
		registry[photo.ID] = primitiveAsset{Path: photo.Path, SHA256: photo.SHA256}
		metadata[photo.ID] = photo
	}
	return metadata, nil
}

// ParsePhotoMetadata reads the narrowly scoped sidecar convention used by the
// local photo collection: H1, caption paragraphs and bold searchable fields.
// It retains unknown field labels and the verbatim UTF-8 text for inspection.
func ParsePhotoMetadata(raw []byte, relative string) (PhotoMetadata, error) {
	meta := PhotoMetadata{Authority: PhotoMetadataAuthority, Path: relative, SHA256: fmt.Sprintf("%x", sha256.Sum256(raw)), Text: string(raw), Fields: map[string]string{}, Keywords: []string{}}
	if !utf8.Valid(raw) {
		return meta, fmt.Errorf("photo.sidecar_not_utf8: %s", relative)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	var caption []string
	var heading string
	inMetadata := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") && heading == "" {
			heading = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			inMetadata = true
			continue
		}
		if strings.HasPrefix(line, "- **") || strings.HasPrefix(line, "* **") {
			start := strings.Index(line, "**") + 2
			end := strings.Index(line[start:], "**")
			if end >= 0 {
				key := strings.TrimSuffix(strings.TrimSpace(line[start:start+end]), ":")
				value := strings.TrimSpace(strings.TrimPrefix(line[start+end+2:], ":"))
				if key != "" {
					if _, exists := meta.Fields[key]; exists {
						return meta, fmt.Errorf("photo.duplicate_sidecar_field: %s: %s", relative, key)
					}
					meta.Fields[key] = strings.Trim(value, "`")
				}
			}
			continue
		}
		if !inMetadata && line != "" && !strings.HasPrefix(line, "**") {
			caption = append(caption, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return meta, err
	}
	field := func(names ...string) string {
		for _, name := range names {
			for label, value := range meta.Fields {
				if strings.EqualFold(name, label) {
					return value
				}
			}
		}
		return ""
	}
	meta.Title = field("Title")
	if meta.Title == "" {
		meta.Title = heading
	}
	meta.Caption = strings.Join(caption, "\n\n")
	meta.AltText = field("Alt text")
	meta.Keywords = splitPhotoMetadata(field("Keywords"))
	meta.Topics = splitPhotoMetadata(field("Primary trending topic") + ";" + field("Secondary topics"))
	meta.Style = field("Official Style", "Style")
	meta.Placement = field("Recommended West Monroe placement", "Recommended placement")
	meta.Practice = field("Primary practice area")
	meta.PeopleDescription = field("People")
	meta.Location = field("Location")
	if meta.Title == "" || meta.Caption == "" || len(meta.Keywords) == 0 {
		return meta, fmt.Errorf("photo.sidecar_semantics_required: %s", relative)
	}
	return meta, nil
}

func splitPhotoMetadata(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ';' })
	return cleanAssetTags(fields)
}

// RegisterPhotoLibrary creates a reproducible snapshot from every JPEG in the
// requested collection. Originals and descriptive sidecars are read only.
func RegisterPhotoLibrary(brandingRoot, out string) (PhotoRegistrationReport, error) {
	report := PhotoRegistrationReport{Schema: "pptxgengo.photo-registration.v1"}
	if brandingRoot == "" || out == "" {
		return report, fmt.Errorf("photo.root_and_new_output_required")
	}
	if _, err := os.Lstat(out); !os.IsNotExist(err) {
		return report, fmt.Errorf("photo.new_output_required: %s", out)
	}
	root := filepath.Join(brandingRoot, "West Monroe Photos")
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("photo.symlink_original_forbidden: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".jpeg" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return report, err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return report, fmt.Errorf("photo.empty_collection")
	}
	registeredPaths := map[string]PrimitiveAssetReference{}
	for _, ref := range PrimitiveAssetCatalog() {
		if strings.HasPrefix(ref.Path, "West Monroe Photos/") {
			if _, exists := registeredPaths[ref.Path]; exists {
				return report, fmt.Errorf("photo.ambiguous_existing_path: %s", ref.Path)
			}
			registeredPaths[ref.Path] = ref
		}
	}
	snapshot := PhotoRegistrySnapshot{Schema: PhotoRegistrySchema, SourceDirectory: "West Monroe Photos", Photos: []RegisteredPhoto{}}
	for _, path := range paths {
		relative, err := filepath.Rel(brandingRoot, path)
		if err != nil {
			return report, err
		}
		relative = filepath.ToSlash(relative)
		file, err := os.Open(path)
		if err != nil {
			return report, err
		}
		digest := sha256.New()
		size, readErr := io.Copy(digest, file)
		closeErr := file.Close()
		if readErr != nil {
			return report, readErr
		}
		if closeErr != nil {
			return report, closeErr
		}
		hash := fmt.Sprintf("%x", digest.Sum(nil))
		file, err = os.Open(path)
		if err != nil {
			return report, err
		}
		dimensions, format, readErr := image.DecodeConfig(file)
		closeErr = file.Close()
		if readErr != nil {
			return report, fmt.Errorf("photo.image_header: %s: %w", relative, readErr)
		}
		if closeErr != nil {
			return report, closeErr
		}
		if format != "jpeg" {
			return report, fmt.Errorf("photo.jpeg_format_required: %s", relative)
		}
		sidecar := path + ".md"
		info, err := os.Lstat(sidecar)
		if err != nil {
			return report, err
		}
		if !info.Mode().IsRegular() {
			return report, fmt.Errorf("photo.regular_sidecar_required: %s", sidecar)
		}
		raw, err := os.ReadFile(sidecar)
		if err != nil {
			return report, err
		}
		meta, err := ParsePhotoMetadata(raw, relative+".md")
		if err != nil {
			return report, err
		}
		id := StablePhotoID(relative)
		if prior, exists := registeredPaths[relative]; exists {
			id = prior.Key
			report.PreservedIDs++
			if strings.HasPrefix(id, "photo-") && prior.SHA256 != hash {
				return report, fmt.Errorf("photo.curated_original_drift: %s", id)
			}
		}
		snapshot.Photos = append(snapshot.Photos, RegisteredPhoto{ID: id, Path: relative, SHA256: hash, Bytes: size, Width: dimensions.Width, Height: dimensions.Height, MIME: "image/jpeg", Metadata: meta})
		report.Bytes += size
	}
	encoded, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return report, err
	}
	encoded = append(encoded, '\n')
	if _, err = decodePhotoRegistry(encoded); err != nil {
		return report, err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return report, err
	}
	file, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return report, err
	}
	_, writeErr := file.Write(encoded)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(out)
		if writeErr != nil {
			return report, writeErr
		}
		return report, closeErr
	}
	report.Photos = len(snapshot.Photos)
	report.SnapshotSHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded))
	return report, nil
}

func registeredPhoto(key, path string) (RegisteredPhoto, bool) {
	photo, ok := photoRegistryMetadata[key]
	asset, registered := primitiveAssetRegistry[key]
	return photo, ok && registered && asset.Path == photo.Path && asset.SHA256 == photo.SHA256 && (path == "" || photo.Path == path)
}

func registeredPhotoOriginalFacts(key, path string) *PhotoOriginalFacts {
	photo, ok := registeredPhoto(key, path)
	if !ok {
		return nil
	}
	return photoOriginalFacts(photo)
}

func clonePhotoMetadata(meta PhotoMetadata) *PhotoMetadata {
	clone := meta
	clone.Fields = catalogCopyMap(meta.Fields)
	clone.Keywords = catalogCopySlice(meta.Keywords)
	clone.Topics = catalogCopySlice(meta.Topics)
	return &clone
}

func photoOrientation(width, height int) string {
	if width > height {
		return "landscape"
	}
	if height > width {
		return "portrait"
	}
	return "square"
}

func photoPeople(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, prefix := range []string{"no people", "none visible", "none discernible", "none as subjects", "no visible people"} {
		if strings.HasPrefix(lower, prefix) {
			return "no"
		}
	}
	for _, prefix := range []string{"one ", "two ", "three ", "four ", "five ", "several ", "multiple ", "at least ", "approximately ", "a group ", "a person ", "a woman ", "a man ", "adults ", "people "} {
		if strings.HasPrefix(lower, prefix) {
			return "yes"
		}
	}
	return "unknown"
}

func enrichPhotoAssetMetadata(key, path string, meta assetMetadata, curated bool) assetMetadata {
	photo, ok := registeredPhoto(key, path)
	if !ok {
		return meta
	}
	source := photo.Metadata
	if !curated {
		meta.name = source.Title
		meta.description = source.Caption
		meta.people = photoPeople(source.PeopleDescription)
	}
	meta.kind = "photo"
	meta.orientation = photoOrientation(photo.Width, photo.Height)
	tags := append(catalogCopySlice(meta.tags), source.Keywords...)
	tags = append(tags, source.Topics...)
	tags = append(tags, splitPhotoMetadata(source.Style)...)
	meta.tags = cleanAssetTags(tags)
	if source.Practice != "" {
		meta.industry = cleanAssetTags(append(catalogCopySlice(meta.industry), source.Practice))
	}
	if len(meta.industry) == 0 {
		parts := strings.Split(photo.Path, "/")
		for i, part := range parts {
			if (part == "Practices" || part == "Other") && i+1 < len(parts) {
				meta.industry = []string{strings.ToLower(parts[i+1])}
				break
			}
		}
	}
	if source.Location != "" {
		meta.setting = cleanAssetTags(append(catalogCopySlice(meta.setting), source.Location))
	}
	meta.photoMetadata = clonePhotoMetadata(source)
	meta.searchText = source.Text
	return meta
}

// AssetRegistryFingerprint binds both registered original facts and semantic
// source metadata. SQLite built before an asset/sidecar update is rejected.
func AssetRegistryFingerprint() string {
	type entry struct {
		Ref      PrimitiveAssetReference
		Metadata assetMetadataFingerprint
		Photo    *PhotoMetadata
		Original *PhotoOriginalFacts
	}
	entries := []entry{}
	for _, ref := range PrimitiveAssetCatalog() {
		meta := assetMetadataFor(ref.Key, ref.Path)
		entries = append(entries, entry{ref, assetMetadataFingerprint{meta.name, meta.description, meta.kind, meta.people, meta.orientation, meta.tags, meta.industry, meta.setting}, meta.photoMetadata, registeredPhotoOriginalFacts(ref.Key, ref.Path)})
	}
	raw, _ := json.Marshal(entries)
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

type assetMetadataFingerprint struct {
	Name, Description, Kind, People, Orientation string
	Tags, Industry, Setting                      []string
}
