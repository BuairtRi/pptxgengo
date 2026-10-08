package deckproject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/buairtri/pptxgengo/internal/noreplacedir"
)

const TextReviewPacketSchema = "pptxgengo.text-review-packet.v1"

type TextReviewPacketFile struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type TextReviewPacketManifest struct {
	Schema    string                          `json:"schema"`
	ProjectID string                          `json:"project_id"`
	Files     map[string]TextReviewPacketFile `json:"files"`
}
type TextReviewPacket struct {
	Root         string
	Report       TextReconciliationReport
	ReportSHA256 string
	Edited       []byte
	files        map[string][]byte
}

// WriteTextReviewPacket retains edited bytes, immutable baseline inputs and the
// complete current authored tree. A closed manifest binds every file. Publication
// uses kernel no-replace semantics; it does not update project source or state.
func WriteTextReviewPacket(p *Project, b *TextBaseline, edited []byte, destination string) (*TextReviewPacket, error) {
	return writeReviewPacket(p, b, edited, destination, "", "")
}
func WriteGeometryReviewPacket(p *Project, b *TextBaseline, edited []byte, destination, bundle, engine string) (*TextReviewPacket, error) {
	return writeReviewPacket(p, b, edited, destination, bundle, engine)
}
func WriteMappedGeometryReviewPacket(p *Project, b *TextBaseline, edited []byte, destination, bundle, engine string, mappingRaw []byte) (*TextReviewPacket, error) {
	m, e := DecodeNativeStructureMap(mappingRaw)
	if e != nil {
		return nil, e
	}
	return writeReviewPacket(p, b, edited, destination, bundle, engine, m.Copies...)
}
func writeReviewPacket(p *Project, b *TextBaseline, edited []byte, destination, bundle, engine string, mappings ...NativeCopyMapping) (*TextReviewPacket, error) {
	analysis, e := prepareMappedNativeCopies(edited, b, mappings)
	if e != nil {
		return nil, e
	}
	report, e := ReconcileText(p, b, analysis)
	if e != nil {
		return nil, e
	}
	if bundle != "" {
		if e = addGeometryReconciliation(p, b, analysis, &report, bundle, engine, mappings...); e != nil {
			return nil, e
		}
	}
	report.EditedPPTXSHA256 = digest(edited)
	files := map[string][]byte{"report.json": canonical(report), "edited.pptx": edited, "current-source.canonical.json": p.Canonical, "current-lock.json": b.files["toolchain.lock.json"]}
	if len(mappings) > 0 {
		files["structure-map.json"] = canonical(NativeStructureMap{StructureMapSchema, mappings})
	}
	for name, raw := range b.files {
		files["baseline/"+name] = raw
	}
	for name, raw := range p.SourceFiles {
		files["current/"+name] = raw
	}
	if len(files) > 40000 {
		return nil, fmt.Errorf("reconcile.packet_too_many_files")
	}
	absolute, e := filepath.Abs(destination)
	if e != nil {
		return nil, e
	}
	parent, e := filepath.EvalSymlinks(filepath.Dir(absolute))
	if e != nil {
		return nil, e
	}
	absolute = filepath.Join(parent, filepath.Base(absolute))
	if _, e = os.Lstat(absolute); e == nil {
		return nil, fmt.Errorf("reconcile.packet_destination_exists")
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	stage, e := os.MkdirTemp(parent, ".text-review-stage-")
	if e != nil {
		return nil, e
	}
	defer os.RemoveAll(stage)
	manifest := TextReviewPacketManifest{Schema: TextReviewPacketSchema, ProjectID: p.Document.ID, Files: map[string]TextReviewPacketFile{}}
	names := []string{}
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	folded := map[string]bool{}
	var total int64
	for _, name := range names {
		if e = validReviewPath(name); e != nil {
			return nil, e
		}
		fold := strings.ToLower(name)
		if folded[fold] {
			return nil, fmt.Errorf("reconcile.packet_case_collision")
		}
		folded[fold] = true
		raw := files[name]
		max := lineageMaxPart
		if name == "edited.pptx" || name == "baseline/deck.pptx" {
			max = lineageMaxPackage
		}
		if len(raw) > max {
			return nil, fmt.Errorf("reconcile.packet_file_too_large: %s", name)
		}
		total += int64(len(raw))
		if total > 1<<30 {
			return nil, fmt.Errorf("reconcile.packet_exceeds_1GiB")
		}
		if e = writeExclusive(filepath.Join(stage, filepath.FromSlash(name)), raw, 0444); e != nil {
			return nil, e
		}
		manifest.Files[name] = TextReviewPacketFile{digest(raw), int64(len(raw))}
	}
	if e = writeExclusive(filepath.Join(stage, "manifest.json"), canonical(manifest), 0444); e != nil {
		return nil, e
	}
	if _, e = ReadTextReviewPacket(stage); e != nil {
		return nil, e
	}
	fresh, e := Load(p.SourcePath)
	if e != nil {
		return nil, e
	}
	if fresh.SourceHash() != p.SourceHash() {
		return nil, fmt.Errorf("reconcile.source_changed_during_packet_creation")
	}
	_, lock, e := ReadLock(fresh)
	if e != nil {
		return nil, e
	}
	if digest(lock) != report.LockSHA256 {
		return nil, fmt.Errorf("reconcile.lock_changed_during_packet_creation")
	}
	if e = noreplacedir.Publish(stage, absolute); e != nil {
		return nil, e
	}
	return ReadTextReviewPacket(absolute)
}

func ReadTextReviewPacket(root string) (*TextReviewPacket, error) {
	absolute, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	st, e := os.Lstat(absolute)
	if e != nil {
		return nil, e
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("reconcile.packet_root_not_regular_directory")
	}
	manifestPath, e := SafePath(absolute, "manifest.json")
	if e != nil {
		return nil, e
	}
	raw, e := readReconcileFile(manifestPath, 16<<20)
	if e != nil {
		return nil, e
	}
	var m TextReviewPacketManifest
	if e = strictInto(json.RawMessage(raw), &m); e != nil {
		return nil, e
	}
	if m.Schema != TextReviewPacketSchema || !stableID.MatchString(m.ProjectID) || len(m.Files) == 0 || len(m.Files) > 40000 {
		return nil, fmt.Errorf("reconcile.invalid_packet_manifest")
	}
	packet := &TextReviewPacket{Root: absolute, files: map[string][]byte{}}
	seen := map[string]bool{}
	folded := map[string]bool{}
	directories := map[string]bool{}
	var total int64
	for name, f := range m.Files {
		if e = validReviewPath(name); e != nil {
			return nil, e
		}
		fold := strings.ToLower(name)
		if folded[fold] || name == "manifest.json" {
			return nil, fmt.Errorf("reconcile.packet_duplicate_manifest_path")
		}
		folded[fold] = true
		if !shaPattern.MatchString(f.SHA256) || f.Size < 0 {
			return nil, fmt.Errorf("reconcile.invalid_packet_file_pin")
		}
		max := int64(lineageMaxPart)
		if name == "edited.pptx" || name == "baseline/deck.pptx" {
			max = lineageMaxPackage
		}
		if f.Size > max {
			return nil, fmt.Errorf("reconcile.packet_file_too_large")
		}
		total += f.Size
		if total > 1<<30 {
			return nil, fmt.Errorf("reconcile.packet_exceeds_1GiB")
		}
		for dir := path.Dir(name); dir != "."; dir = path.Dir(dir) {
			directories[dir] = true
		}
	}
	e = filepath.WalkDir(absolute, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absolute, file)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("reconcile.packet_symlink_forbidden")
		}
		if d.IsDir() {
			if !directories[rel] {
				return fmt.Errorf("reconcile.packet_unlisted_directory: %s", rel)
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("reconcile.packet_nonregular_file")
		}
		if rel == "manifest.json" {
			return nil
		}
		f, ok := m.Files[rel]
		if !ok {
			return fmt.Errorf("reconcile.packet_unlisted_file: %s", rel)
		}
		b, err := readReconcileFile(file, f.Size)
		if err != nil {
			return err
		}
		if int64(len(b)) != f.Size || digest(b) != f.SHA256 {
			return fmt.Errorf("reconcile.packet_file_drift: %s", rel)
		}
		packet.files[rel] = b
		seen[rel] = true
		return nil
	})
	if e != nil {
		return nil, e
	}
	if len(seen) != len(m.Files) {
		return nil, fmt.Errorf("reconcile.packet_missing_files")
	}
	for _, name := range []string{"report.json", "edited.pptx", "current-source.canonical.json", "current-lock.json", "baseline/receipt.json", "baseline/deck.pptx", "baseline/object-map.json", "baseline/source.canonical.json", "baseline/toolchain.lock.json"} {
		if _, ok := packet.files[name]; !ok {
			return nil, fmt.Errorf("reconcile.packet_missing_required_file: %s", name)
		}
	}
	if e = strictInto(json.RawMessage(packet.files["report.json"]), &packet.Report); e != nil {
		return nil, e
	}
	packet.ReportSHA256 = digest(packet.files["report.json"])
	packet.Edited = packet.files["edited.pptx"]
	r := packet.Report
	if r.Schema != TextReconciliationSchema || r.ProjectID != m.ProjectID || digest(packet.Edited) != r.EditedPPTXSHA256 || digest(packet.files["baseline/receipt.json"]) != r.BaselineReceiptSHA256 || digest(packet.files["baseline/deck.pptx"]) != r.BaselinePPTXSHA256 || digest(packet.files["current-source.canonical.json"]) != r.CurrentSemanticSHA256 || digest(packet.files["current-lock.json"]) != r.LockSHA256 {
		return nil, fmt.Errorf("reconcile.packet_identity_mismatch")
	}
	return packet, nil
}
func validReviewPath(name string) error {
	if name == "" || path.IsAbs(name) || path.Clean(name) != name || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "../") {
		return fmt.Errorf("reconcile.unsafe_packet_path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return fmt.Errorf("reconcile.nonportable_packet_path")
		}
		for _, c := range part {
			if c < 32 || strings.ContainsRune(`<>"|?*`, c) {
				return fmt.Errorf("reconcile.nonportable_packet_path")
			}
		}
		base := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) && base[3] >= '1' && base[3] <= '9' {
			return fmt.Errorf("reconcile.reserved_packet_path")
		}
	}
	return nil
}

// Compare packet inputs to a freshly authenticated baseline before replaying the
// report. A forged report cannot authorize fields the current analyzer rejects.
func verifyTextPacketBaseline(packet *TextReviewPacket, b *TextBaseline) error {
	if packet == nil || b == nil {
		return fmt.Errorf("reconcile.packet_and_baseline_required")
	}
	for name, raw := range b.files {
		if !bytes.Equal(packet.files["baseline/"+name], raw) {
			return fmt.Errorf("reconcile.packet_baseline_changed: %s", name)
		}
	}
	return nil
}
