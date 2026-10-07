package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/deckproject"
)

func TestProjectReconcileFlagContracts(t *testing.T) {
	for _, args := range [][]string{{}, {"bad"}, {"propose"}, {"adopt"}, {"propose", "--edited", "x", "--out", "y", "extra"}, {"adopt", "--packet", "x", "--decisions", "y", "--edited", "z"}, {"propose", "--edited", "x", "--out", "y", "--bundle", "v5"}} {
		if e := runProject(append([]string{"reconcile"}, args...)); e == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestProjectReconcileCLIProposeAdoptRepeat(t *testing.T) {
	root := projectDefaultFixture(t)
	bundle, e := filepath.Abs("../../planning/wm-design-contracts/v5/intake-20261003-587-frozen/bundle")
	if e != nil {
		t.Fatal(e)
	}
	projectCommandJSON(t, "init", "--project", root, "--bundle", bundle)
	var built deckproject.Receipt
	if e := json.Unmarshal(projectCommandJSON(t, "build", "--project", root, "--bundle", bundle), &built); e != nil {
		t.Fatal(e)
	}
	original, e := os.ReadFile(filepath.Join(root, "builds", built.BuildID, "deck.pptx"))
	if e != nil {
		t.Fatal(e)
	}
	z, e := zip.NewReader(bytes.NewReader(original), int64(len(original)))
	if e != nil {
		t.Fatal(e)
	}
	var changed bytes.Buffer
	writer := zip.NewWriter(&changed)
	replacements := 0
	for _, entry := range z.File {
		r, e := entry.Open()
		if e != nil {
			t.Fatal(e)
		}
		raw, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			t.Fatal(e)
		}
		if strings.HasPrefix(entry.Name, "ppt/slides/") && strings.HasSuffix(entry.Name, ".xml") {
			before := []byte("Exact project pins survive published defaults.")
			replacements += bytes.Count(raw, before)
			raw = bytes.ReplaceAll(raw, before, []byte("Native colleague copy."))
		}
		header := entry.FileHeader
		w, e := writer.CreateHeader(&header)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(raw); e != nil {
			t.Fatal(e)
		}
	}
	if e := writer.Close(); e != nil {
		t.Fatal(e)
	}
	if replacements != 1 {
		t.Fatalf("fixture edited %d fields", replacements)
	}
	edited := filepath.Join(t.TempDir(), "colleague copy.pptx")
	if e := os.WriteFile(edited, changed.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(t.TempDir(), "text review")
	var response struct {
		Packet       string                               `json:"packet"`
		ReportSHA256 string                               `json:"report_sha256"`
		Report       deckproject.TextReconciliationReport `json:"report"`
	}
	if e := json.Unmarshal(projectCommandJSON(t, "reconcile", "propose", "--project", root, "--edited", edited, "--out", out), &response); e != nil {
		t.Fatal(e)
	}
	resolvedOut, e := filepath.EvalSymlinks(out)
	if e != nil {
		t.Fatal(e)
	}
	if response.Report.Counts["native_only"] != 1 || response.Packet != resolvedOut {
		t.Fatalf("unexpected response %+v", response)
	}
	choices := deckproject.TextReviewDecisions{Schema: deckproject.TextReviewDecisionsSchema, Actor: "Synthetic CLI reviewer", ReportSHA256: response.ReportSHA256}
	for _, field := range response.Report.Fields {
		if field.Status == "native_only" {
			choices.Decisions = append(choices.Decisions, deckproject.TextReviewDecision{FieldID: field.ID, Action: "use_native", Reason: "Synthetic fixture decision"})
		}
	}
	raw, e := json.Marshal(choices)
	if e != nil {
		t.Fatal(e)
	}
	decisions := filepath.Join(t.TempDir(), "decisions.json")
	if e = os.WriteFile(decisions, raw, 0600); e != nil {
		t.Fatal(e)
	}
	args := []string{"reconcile", "adopt", "--project", root, "--packet", out, "--decisions", decisions, "--bundle", bundle}
	first := projectCommandJSON(t, args...)
	second := projectCommandJSON(t, args...)
	if !bytes.Equal(first, second) {
		t.Fatal("repeat command changed receipt")
	}
	p, e := deckproject.Load(root)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[0].Values["copy"] != "Native colleague copy." {
		t.Fatal("CLI did not adopt exact named copy")
	}
	if e := runProject([]string{"reconcile", "propose", "--project", root, "--edited", edited, "--out", out}); e == nil {
		t.Fatal("existing packet replaced")
	}
	projectCommandJSON(t, "build", "--project", root, "--bundle", bundle)
}

func TestProjectReconcileInputBounds(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "input")
	if e := os.WriteFile(path, []byte("exact"), 0600); e != nil {
		t.Fatal(e)
	}
	if raw, e := readReconciliationInput(path, 5); e != nil || string(raw) != "exact" {
		t.Fatal(e)
	}
	if _, e := readReconciliationInput(path, 4); e == nil {
		t.Fatal("oversize input accepted")
	}
	if _, e := readReconciliationInput(root, 100); e == nil {
		t.Fatal("directory accepted")
	}
	link := filepath.Join(root, "link")
	if e := os.Symlink(path, link); e == nil {
		if _, e = readReconciliationInput(link, 100); e == nil {
			t.Fatal("symlink accepted")
		}
	}
}
