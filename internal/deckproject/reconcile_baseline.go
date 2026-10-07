package deckproject

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// TextBaseline is loaded only through receipt-bound verification. Retained bytes
// are used for review packets; the authored baseline is never rewritten.
type TextBaseline struct {
	Receipt       Receipt
	ReceiptSHA256 string
	Objects       Objects
	files         map[string][]byte
	source        map[string]any
	inspection    NativeLineageInspection
	root          string
}

// ReadTextBaseline selects an explicit immutable build, never the latest mtime.
// Empty build selects state's current build and its pinned receipt. Historical
// builds require an explicit trusted receipt hash (for example from a prior packet).
func ReadTextBaseline(p *Project, buildID, receiptSHA string) (*TextBaseline, error) {
	state, e := readState(p)
	if e != nil {
		return nil, e
	}
	if buildID == "" {
		buildID = state.CurrentBuild
	}
	if !stableID.MatchString(buildID) {
		return nil, fmt.Errorf("reconcile.baseline_missing: select a generated build")
	}
	if buildID == state.CurrentBuild {
		if receiptSHA != "" && receiptSHA != state.ReceiptSHA256 {
			return nil, fmt.Errorf("reconcile.receipt_pin_conflicts_with_state")
		}
		receiptSHA = state.ReceiptSHA256
	}
	if !shaPattern.MatchString(receiptSHA) {
		return nil, fmt.Errorf("reconcile.receipt_pin_required: historical baselines require an explicit receipt SHA256")
	}
	root, e := SafePath(p.Root, "builds/"+buildID)
	if e != nil {
		return nil, e
	}
	receiptPath, e := SafePath(root, "receipt.json")
	if e != nil {
		return nil, e
	}
	raw, e := readReconcileFile(receiptPath, 16<<20)
	if e != nil {
		return nil, e
	}
	if digest(raw) != receiptSHA {
		return nil, fmt.Errorf("reconcile.baseline_divergence: receipt changed")
	}
	b := &TextBaseline{ReceiptSHA256: receiptSHA, files: map[string][]byte{"receipt.json": raw}, root: root}
	if e = strictInto(json.RawMessage(raw), &b.Receipt); e != nil {
		return nil, e
	}
	r := b.Receipt
	if r.Schema != "pptxgengo.deck-build-receipt.v1" || r.BuildID != buildID || r.ProjectID != p.Document.ID {
		return nil, fmt.Errorf("reconcile.baseline_identity_mismatch")
	}
	required := map[string]bool{"deck.pptx": true, "deck.yaml": true, "source.canonical.json": true, "object-map.json": true, "toolchain.lock.json": true, "scene.json": true, "layout-report.json": true}
	if len(r.Outputs) > 40000 {
		return nil, fmt.Errorf("reconcile.baseline_too_many_outputs")
	}
	keys := []string{}
	for rel := range r.Outputs {
		keys = append(keys, rel)
	}
	sort.Strings(keys)
	var aggregate int64
	for _, rel := range keys {
		want := r.Outputs[rel]
		if !shaPattern.MatchString(want) {
			return nil, fmt.Errorf("reconcile.invalid_baseline_output_hash: %s", rel)
		}
		file, e := SafePath(root, rel)
		if e != nil {
			return nil, e
		}
		max := int64(lineageMaxPart)
		if rel == "deck.pptx" {
			max = lineageMaxPackage
		}
		// Verify every receipt output; retain only the inputs needed by reconciliation.
		sum, size, data, e := hashReconcileFile(file, max, required[rel])
		if e != nil {
			return nil, e
		}
		aggregate += size
		if aggregate > 1<<30 {
			return nil, fmt.Errorf("reconcile.baseline_exceeds_1GiB")
		}
		if sum != want {
			return nil, fmt.Errorf("reconcile.baseline_divergence: %s changed", rel)
		}
		if required[rel] {
			b.files[rel] = data
		}
	}
	for rel := range required {
		if _, ok := b.files[rel]; !ok {
			return nil, fmt.Errorf("reconcile.baseline_missing_output: %s", rel)
		}
	}
	if digest(b.files["source.canonical.json"]) != r.SemanticSHA256 || digest(b.files["toolchain.lock.json"]) != r.LockSHA256 {
		return nil, fmt.Errorf("reconcile.baseline_source_or_lock_pin_mismatch")
	}
	// Canonical source must be exactly one JSON object. The complete object map is
	// receipt-pinned and strictly decoded before any source pointer is followed.
	decoder := json.NewDecoder(bytes.NewReader(b.files["source.canonical.json"]))
	decoder.UseNumber()
	if e = decoder.Decode(&b.source); e != nil {
		return nil, e
	}
	var extra any
	if e = decoder.Decode(&extra); e != io.EOF {
		return nil, fmt.Errorf("reconcile.baseline_source_trailing_data")
	}
	if b.source["id"] != r.ProjectID || b.source["schema"] != Schema {
		return nil, fmt.Errorf("reconcile.baseline_source_identity_mismatch")
	}
	if e = strictInto(json.RawMessage(b.files["object-map.json"]), &b.Objects); e != nil {
		return nil, e
	}
	if b.Objects.DeckID != r.ProjectID || b.Objects.SourceSHA256 != r.SemanticSHA256 || b.Objects.Schema != "pptxgengo.deck-object-map.v1" || b.Objects.TextModelSchema != NativeTextModelSchema || b.Objects.Lineage == nil || b.Objects.Lineage.BuildToken != r.NativeLineageBuildToken || b.Objects.Lineage.LockSHA256 != r.LockSHA256 {
		return nil, fmt.Errorf("reconcile.baseline_object_map_mismatch")
	}
	b.inspection, e = InspectNativeLineage(b.files["deck.pptx"], b.Objects)
	if e != nil {
		return nil, e
	}
	if len(b.inspection.Issues) != 0 {
		return nil, fmt.Errorf("reconcile.baseline_native_identity_invalid: %s", b.inspection.Issues[0].Kind)
	}
	native := map[string]NativeLineageObject{}
	for _, o := range b.inspection.Objects {
		native[o.ShapeToken] = o
	}
	for _, object := range b.Objects.Objects {
		shape := native[object.ShapeToken].shape
		if shape == nil || nativeParagraphText(nativeParagraphs(shape)) != object.NativeText {
			return nil, fmt.Errorf("reconcile.baseline_native_text_mismatch")
		}
		if object.TextMappingContract != "" && !editableCardBaselineParagraphsValid(object, nativeParagraphs(shape)) {
			return nil, fmt.Errorf("reconcile.baseline_paragraph_contract_invalid")
		}
		for _, field := range object.Fields {
			if field.Status != "plain_text_baseline" {
				continue
			}
			value, e := lookupPointer(b.source, field.SourcePointer)
			if e != nil {
				return nil, e
			}
			text, ok := value.(string)
			nativeText, addressOK := nativeObjectFieldText(object, field, object.Paragraphs)
			if !ok || !addressOK || digest([]byte(text)) != field.SourceValueSHA256 || digest([]byte(nativeText)) != field.BaselineNativeSHA256 || text != nativeText {
				return nil, fmt.Errorf("reconcile.baseline_field_text_mismatch")
			}
			if !nativeFieldIdentityValid(object, field) || field.SourceSlot == "" || object.SourceSlots[field.SourcePointer] != field.SourceSlot {
				return nil, fmt.Errorf("reconcile.baseline_field_identity_mismatch")
			}
		}
	}
	return b, nil
}

func readReconcileFile(path string, max int64) ([]byte, error) {
	_, _, data, e := hashReconcileFile(path, max, true)
	return data, e
}
func hashReconcileFile(path string, max int64, retain bool) (string, int64, []byte, error) {
	before, e := os.Lstat(path)
	if e != nil {
		return "", 0, nil, e
	}
	if !before.Mode().IsRegular() || before.Size() > max {
		return "", 0, nil, fmt.Errorf("reconcile.invalid_or_oversized_file: %s", filepath.Base(path))
	}
	f, e := os.Open(path)
	if e != nil {
		return "", 0, nil, e
	}
	defer f.Close()
	opened, e := f.Stat()
	if e != nil {
		return "", 0, nil, e
	}
	if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		return "", 0, nil, fmt.Errorf("reconcile.file_identity_changed")
	}
	h := sha256.New()
	var b bytes.Buffer
	var writer io.Writer = h
	if retain {
		writer = io.MultiWriter(h, &b)
	}
	size, e := io.Copy(writer, io.LimitReader(f, max+1))
	if e != nil {
		return "", 0, nil, e
	}
	after, e := f.Stat()
	if e != nil {
		return "", 0, nil, e
	}
	if size > max || size != opened.Size() || after.Size() != opened.Size() || !after.ModTime().Equal(opened.ModTime()) {
		return "", 0, nil, fmt.Errorf("reconcile.file_changed_during_read")
	}
	return hex.EncodeToString(h.Sum(nil)), size, b.Bytes(), nil
}
