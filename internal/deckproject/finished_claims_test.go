package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func claimReuseProject(t *testing.T, format string) *Project {
	t.Helper()
	p := reuseProject(t)
	path := "claims.yaml"
	raw := []byte("# Existing authored registry comment\nschema: pptxgengo.claims.v1\nclaims:\n  - id: claim-1\n    text: 'A supplied finding keeps claim-1 as business copy.'\n    source: 'Interview notes; not an inferred file path.'\n  - id: unrelated\n    text: 'Unrelated authored finding.'\n")
	if format == "markdown" {
		path = "claims.md"
		raw = []byte("# Authored evidence\n\n## claim-1\nA supplied finding keeps claim-1 as business copy.\n\n## unrelated\nUnrelated authored finding.\n\n```md\n## fake-example\n```\n")
	}
	if e := os.WriteFile(filepath.Join(p.Root, path), raw, 0600); e != nil {
		t.Fatal(e)
	}
	p.Document.Context["claims"] = path
	p.Document.Slides[0].EvidenceRefs = []string{"claim-1"}
	if format == "artifact" {
		proof := []byte("Owned source evidence bytes; no external factual verification claimed.")
		if e := os.WriteFile(filepath.Join(p.Root, "proof.txt"), proof, 0600); e != nil {
			t.Fatal(e)
		}
		raw = []byte("schema: pptxgengo.claims.v1\nclaims:\n  - id: claim-1\n    text: 'A supplied finding keeps claim-1 as business copy.'\n    source: 'Interview notes; not an inferred file path.'\n    artifacts:\n      - path: proof.txt\n        sha256: " + digest(proof) + "\n")
		if e := os.WriteFile(filepath.Join(p.Root, path), raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
	tree, e := editYAMLNode(p.Document)
	if e != nil {
		t.Fatal(e)
	}
	raw, e = encodeSourceYAML(tree)
	if e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p.SourcePath, raw, 0600); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if e := ValidateEditorial(p); e != nil {
		t.Fatal(e)
	}
	return p
}

func TestFinishedSlideClaimsIndependentStructuredMarkdownAndArtifactReuse(t *testing.T) {
	for _, format := range []string{"yaml", "markdown", "artifact"} {
		t.Run(format, func(t *testing.T) {
			source := claimReuseProject(t, format)
			registryPath := source.Document.Context["claims"]
			original := mustRead(t, filepath.Join(source.Root, registryPath))
			library, m := publishReuse(t, source, 1)
			_, deps, packageFiles, e := readFinishedSource(library, m)
			if e != nil {
				t.Fatal(e)
			}
			if len(deps.Claims) != 1 || deps.Claims[0].ID != "claim-1" {
				t.Fatal("unselected claims imported", deps.Claims)
			}
			if format == "markdown" {
				ref := deps.Claims[0].Registry
				if ref == nil || ref.ClaimID != "claim-1" || !bytes.Equal(packageFiles[ref.Path], original) || deps.Claims[0].Text != "" {
					t.Fatal("Markdown inferred or not preserved")
				}
			}
			if format == "artifact" {
				if len(deps.Claims[0].Artifacts) != 1 || !bytes.Equal(packageFiles[deps.Claims[0].Artifacts[0].Path], mustRead(t, filepath.Join(source.Root, "proof.txt"))) {
					t.Fatal("proof bytes lost")
				}
			}
			ids := []string{}
			for _, targetFormat := range []string{"yaml", "markdown"} {
				p := claimReuseProject(t, targetFormat)
				oldRegistryPath := p.Document.Context["claims"]
				oldRegistry := mustRead(t, filepath.Join(p.Root, oldRegistryPath))
				r, e := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "evidenced-copy", Rationale: "Owned fixture content fits the deck", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
				if e != nil {
					t.Fatal(e)
				}
				id := r.Library.EvidenceRemaps["claim-1"]
				if id == "" || id == "claim-1" {
					t.Fatal("claim identity shared", r.Library)
				}
				ids = append(ids, id)
				p, e = Load(p.Root)
				if e != nil {
					t.Fatal(e)
				}
				if e := ValidateEditorial(p); e != nil {
					t.Fatal(e)
				}
				claims, files, e := readClaimDependencies(p)
				if e != nil {
					t.Fatal(e)
				}
				found := false
				for _, claim := range claims.Claims {
					if claim.ID == id {
						found = true
						if format == "markdown" {
							if claim.Registry == nil || claim.Registry.ClaimID != "claim-1" || !bytes.Equal(files[claim.Registry.Path], original) {
								t.Fatal("imported Markdown evidence changed")
							}
						} else if claim.Text != deps.Claims[0].Text || claim.Source != deps.Claims[0].Source {
							t.Fatal("claim copy/source rewritten", claim)
						}
						if format == "artifact" && (len(claim.Artifacts) != 1 || !bytes.Equal(files[claim.Artifacts[0].Path], mustRead(t, filepath.Join(source.Root, "proof.txt")))) {
							t.Fatal("destination proof bytes lost")
						}
					}
				}
				if !found || p.Document.Slides[len(p.Document.Slides)-1].EvidenceRefs[0] != id {
					t.Fatal("imported slide and registry disagree")
				}
				if targetFormat == "markdown" {
					if !bytes.Equal(mustRead(t, filepath.Join(p.Root, oldRegistryPath)), oldRegistry) || p.Document.Context["claims"] == oldRegistryPath {
						t.Fatal("old Markdown changed instead of being pinned")
					}
				} else if !bytes.Contains(mustRead(t, filepath.Join(p.Root, oldRegistryPath)), []byte("Existing authored registry comment")) {
					t.Fatal("registry comment lost")
				}
				if _, e := Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
					t.Fatal("evidenced copy cannot build", e)
				}
				dependencies, e := dependencies(p)
				if e != nil {
					t.Fatal(e)
				}
				if format != "yaml" {
					seen := false
					for key := range dependencies {
						if strings.HasPrefix(key, "context:claim-evidence:") {
							seen = true
						}
					}
					if !seen {
						t.Fatal("evidence not in stage dependency pins")
					}
				}
			}
			if ids[0] == ids[1] || !bytes.Equal(mustRead(t, filepath.Join(source.Root, registryPath)), original) {
				t.Fatal("claim copies share identity or source changed")
			}
		})
	}
}

func TestFinishedSlideClaimsSurviveSplitReorderAndHandoff(t *testing.T) {
	source := claimReuseProject(t, "artifact")
	library, manifest := publishReuse(t, source, 1)
	p := claimReuseProject(t, "markdown")
	receipt, e := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "evidenced-copy", Rationale: "Owned fixture for maintained handoff", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine})
	if e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Split(p, SplitOptions{Bundle: bundle(t), StockEditor: StockEditableSlide}); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = OperateSlide(p, SlideOperation{Action: "move", ID: "evidenced-copy", Before: p.Document.Slides[0].ID}); e != nil {
		t.Fatal(e)
	}
	p, e = Load(p.Root)
	if e != nil {
		t.Fatal(e)
	}
	if p.Document.Slides[0].ID != "evidenced-copy" || p.Document.Slides[0].EvidenceRefs[0] != receipt.Library.EvidenceRemaps["claim-1"] || p.SlideFiles["evidenced-copy"] == "" {
		t.Fatal("split/reorder lost independent identity")
	}
	log, e := ReadCompositionLog(p)
	if e != nil || log["evidenced-copy"].Library.RevisionSHA256 != manifest.RevisionSHA256 || log["evidenced-copy"].Library.EvidenceRemaps["claim-1"] != p.Document.Slides[0].EvidenceRefs[0] {
		t.Fatal("library provenance lost", e)
	}
	if _, e = Build(p, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
		t.Fatal(e)
	}
	_, evidence, e := readClaimDependencies(p)
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"client", "maintainer"} {
		out := filepath.Join(t.TempDir(), mode+".zip")
		if _, e = Export(p, ExportOptions{Mode: mode, Out: out, Bundle: bundle(t)}); e != nil {
			t.Fatal(e)
		}
		z, e := zip.OpenReader(out)
		if e != nil {
			t.Fatal(e)
		}
		members := map[string][]byte{}
		for _, f := range z.File {
			input, e := f.Open()
			if e != nil {
				t.Fatal(e)
			}
			raw, e := io.ReadAll(input)
			input.Close()
			if e != nil {
				t.Fatal(e)
			}
			members[f.Name] = raw
		}
		z.Close()
		if mode == "client" {
			if len(members["deck.pptx"]) == 0 || members["deck.yaml"] != nil {
				t.Fatal("invalid client export")
			}
			for name := range evidence {
				if members[name] != nil {
					t.Fatal("private claim/evidence file leaked", name)
				}
			}
			continue
		}
		for name, raw := range evidence {
			if !bytes.Equal(members[name], raw) {
				t.Fatal("handoff evidence changed", name)
			}
		}
		if !bytes.Equal(members[p.SlideFiles["evidenced-copy"]], p.SourceFiles[p.SlideFiles["evidenced-copy"]]) {
			t.Fatal("handoff lost slide source")
		}
		handoff := t.TempDir()
		for name, raw := range members {
			path, e := SafePath(handoff, name)
			if e != nil {
				t.Fatal(e)
			}
			if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(path, raw, 0600); e != nil {
				t.Fatal(e)
			}
		}
		restored, e := Load(handoff)
		if e != nil {
			t.Fatal(e)
		}
		if e := ValidateEditorial(restored); e != nil {
			t.Fatal("relocated handoff invalid", e)
		}
		if _, e := Build(restored, BuildOptions{Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e != nil {
			t.Fatal("relocated handoff cannot rebuild", e)
		}
	}
}

func TestFinishedSlideClaimsDependencyChangesInvalidateApproval(t *testing.T) {
	p := claimReuseProject(t, "artifact")
	id := p.Document.Slides[0].ID
	a, e := Approve(p, "content", "Owned test reviewer", []string{id})
	if e != nil {
		t.Fatal(e)
	}
	registry := filepath.Join(p.Root, p.Document.Context["claims"])
	changed := []byte("New owned evidence bytes")
	old := mustRead(t, filepath.Join(p.Root, "proof.txt"))
	if e := os.WriteFile(filepath.Join(p.Root, "proof.txt"), changed, 0600); e != nil {
		t.Fatal(e)
	}
	raw := bytes.ReplaceAll(mustRead(t, registry), []byte(digest(old)), []byte(digest(changed)))
	if e := os.WriteFile(registry, raw, 0600); e != nil {
		t.Fatal(e)
	}
	state, e := Status(p)
	if e != nil {
		t.Fatal(e)
	}
	if len(state.Approvals) != 1 || state.Approvals[0].ID != a.ID || state.Approvals[0].Valid {
		t.Fatal("changed declared proof kept approval")
	}
	if _, e = Approve(p, "content", "Owned test reviewer", []string{id}); e != nil {
		t.Fatal(e)
	}
	p.Document.Slides[0].EvidenceRefs = nil
	state, e = Status(p)
	if e != nil {
		t.Fatal(e)
	}
	if state.Approvals[len(state.Approvals)-1].Valid {
		t.Fatal("removing claim reference kept approval")
	}
}

func TestFinishedSlideClaimsObservedDependencyDriftRefusesWrites(t *testing.T) {
	for _, relative := range []string{"claims.yaml", "proof.txt"} {
		t.Run(relative, func(t *testing.T) {
			p := claimReuseProject(t, "artifact")
			_, observed, e := readClaimDependencies(p)
			if e != nil {
				t.Fatal(e)
			}
			before := append([]byte(nil), p.Raw...)
			changes := map[string][]byte{"deck.yaml": bytes.Replace(p.Raw, []byte("year: 2026"), []byte("year: 2027"), 1), "evidence/new-proof": []byte("New evidence must not be committed")}
			_, e = commitSourceChangesObserved(p, changes, observed, func(candidate *Project) error {
				if e := ValidateEditorial(candidate); e != nil {
					return e
				}
				return os.WriteFile(filepath.Join(p.Root, relative), append(observed[relative], []byte("\nOther writer")...), 0600)
			})
			if e == nil || !strings.Contains(e.Error(), "dependency changed") {
				t.Fatal("evidence drift ignored", e)
			}
			if !bytes.Equal(mustRead(t, p.SourcePath), before) {
				t.Fatal("guarded source changed")
			}
			if _, e := os.Stat(filepath.Join(p.Root, "evidence/new-proof")); !os.IsNotExist(e) {
				t.Fatal("new evidence written on refusal", e)
			}
		})
	}
}

func TestFinishedSlideClaimsRejectChangedMissingAndUnregisteredEvidence(t *testing.T) {
	p := claimReuseProject(t, "artifact")
	if e := os.WriteFile(filepath.Join(p.Root, "proof.txt"), []byte("Changed proof"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := ValidateEditorial(p); e == nil {
		t.Fatal("changed proof accepted")
	}
	if _, e := dependencies(p); e == nil {
		t.Fatal("changed proof not detected in dependency pins")
	}
	if _, _, e := closeFinishedClaims(p, p.Document.Slides[0]); e == nil {
		t.Fatal("changed proof published")
	}
	p = claimReuseProject(t, "markdown")
	p.Document.Slides[0].EvidenceRefs = []string{"fake-example"}
	if e := ValidateEditorial(p); e == nil {
		t.Fatal("code fence treated as evidence")
	}
	p.Document.Context["claims"] = ""
	if e := ValidateEditorial(p); e == nil {
		t.Fatal("evidence accepted without registry")
	}
}

func TestFinishedSlideClaimsMalformedPackageCannotWriteSource(t *testing.T) {
	for _, kind := range []string{"missing claim", "wrong selector", "wrong hash", "wrong role", "empty text"} {
		t.Run(kind, func(t *testing.T) {
			p := claimReuseProject(t, "markdown")
			library, m := publishReuse(t, p, 1)
			_, deps, files, e := readFinishedSource(library, m)
			if e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "missing claim":
				deps.Claims = nil
			case "wrong selector":
				deps.Claims[0].Registry.ClaimID = "missing"
			case "wrong hash":
				deps.Claims[0].Registry.SHA256 = digest([]byte("wrong"))
			case "wrong role":
				for i := range m.Files {
					if m.Files[i].Role == "evidence" {
						m.Files[i].Role = "documentation"
					}
				}
			case "empty text":
				deps.Claims[0].Registry = nil
			}
			files["dependencies.json"] = canonical(deps)
			// Re-sealing cannot make a semantically false dependency graph valid.
			for i := range m.Files {
				m.Files[i].SHA256 = digest(files[m.Files[i].Path])
				m.Files[i].Bytes = int64(len(files[m.Files[i].Path]))
				if e := os.WriteFile(filepath.Join(library, filepath.FromSlash(m.Files[i].Path)), files[m.Files[i].Path], 0600); e != nil {
					t.Fatal(e)
				}
			}
			if e := m.Seal(); e != nil {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(m)
			if e := os.WriteFile(filepath.Join(library, "manifest.json"), raw, 0600); e != nil {
				t.Fatal(e)
			}
			before := append([]byte(nil), p.Raw...)
			if _, e := InsertFinishedSlide(p, FinishedSlideInsertOptions{Package: library, ID: "bad-copy", Rationale: "Owned fixture", AllowDraft: true, Bundle: bundle(t), Engine: wmdesign.CandidateEngine}); e == nil {
				t.Fatal("false dependencies inserted")
			}
			if !bytes.Equal(mustRead(t, p.SourcePath), before) {
				t.Fatal("rejected dependencies wrote source")
			}
		})
	}
}
