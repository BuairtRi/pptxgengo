package deckproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditorialEvidenceAndComposition(t *testing.T) {
	p := example(t)
	pin(t, p)
	p.Document.Context["claims"] = "claims.yaml"
	p.Document.Slides[0].EvidenceRefs = []string{"claim-1"}
	write := func(name, value string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(p.Root, name), []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("claims.yaml", "schema: pptxgengo.claims.v1\nclaims:\n - id: claim-1\n   text: Supplied client evidence.\n   source: Interview notes.\n")
	if err := ValidateEditorial(p); err != nil {
		t.Fatal(err)
	}
	p.Document.Slides[0].EvidenceRefs = []string{"missing"}
	if err := ValidateEditorial(p); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Fatal("missing evidence passed", err)
	}
	p.Document.Slides[0].EvidenceRefs = nil
	write("composition-log.yaml", "schema: pptxgengo.composition-log.v1\nslides: {}\n")
	if err := ValidateEditorial(p); err == nil {
		t.Fatal("missing composition passed")
	}
	write("composition-log.yaml", "schema: pptxgengo.composition-log.v1\nslides:\n maintain-the-source:\n  purpose: Delivery controls\n  chosen_template: cards/3\n  rationale: Three parallel controls\n local-composition:\n  purpose: Explain customization\n  chosen_template: editorial-photo\n  rationale: Paired editorial content and photograph\n")
	if err := ValidateEditorial(p); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadCompositionLog(p)
	if err != nil || len(entries) != 2 {
		t.Fatal(entries, err)
	}
	deps, err := dependencies(p)
	if err != nil || deps["context:composition_log"] == "" {
		t.Fatal("untracked log", err)
	}
	p.Document.Context["claims"] = "claims.md"
	p.Document.Slides[0].EvidenceRefs = []string{"claim-1"}
	write("claims.md", "# Evidence\n\n## claim-1\nInterview finding.\n")
	if err := ValidateEditorial(p); err != nil {
		t.Fatal(err)
	}
	write("claims.md", "Claim identity is not inferred from this prose.\n")
	if err := ValidateEditorial(p); err == nil {
		t.Fatal("plain prose verified ID")
	}
}

func TestPageSpecStrictTypesAndAliases(t *testing.T) {
	for _, raw := range []string{"title: 42\n", "title: Test\nother: bad\n", "title: &copy Test\neyebrow: *copy\n", "title: One\n---\ntitle: Two\n", "title: One\ntitle: Two\n"} {
		file := filepath.Join(t.TempDir(), "page.yaml")
		if err := os.WriteFile(file, []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadPageSpec(file); err == nil {
			t.Fatal("accepted unsafe page spec", raw)
		}
	}
}

func TestEditorialMarkdownExamplesAreNotClaims(t *testing.T) {
	p := example(t)
	p.Document.Context["claims"] = "claims.md"
	p.Document.Slides[0].EvidenceRefs = []string{"example-id"}
	for _, raw := range []string{"```markdown\n## example-id\n```\n", "~~~\nVisible example {#example-id}\n~~~\n", "    ## example-id\n", "An inline example `{#example-id}`.\n"} {
		if err := os.WriteFile(filepath.Join(p.Root, "claims.md"), []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		if err := ValidateEditorial(p); err == nil {
			t.Fatal("code example registered as evidence", raw)
		}
	}
	if err := os.WriteFile(filepath.Join(p.Root, "claims.md"), []byte("## Evidence {#example-id}\nSupplied finding.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEditorial(p); err != nil {
		t.Fatal(err)
	}
}

func TestEditorialStrictRegistryTypesAndDocuments(t *testing.T) {
	p := example(t)
	p.Document.Context["claims"] = "claims.yaml"
	for _, raw := range []string{"schema: pptxgengo.claims.v1\nclaims: [{id: claim, text: 42}]\n", "schema: pptxgengo.claims.v1\nclaims: []\nunknown: value\n", "schema: pptxgengo.claims.v1\nclaims: []\n---\nclaims: []\n", "schema: pptxgengo.claims.v1\nclaims: &list []\nother: *list\n"} {
		if err := os.WriteFile(filepath.Join(p.Root, "claims.yaml"), []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		if err := ValidateEditorial(p); err == nil {
			t.Fatal("invalid editorial registry accepted", raw)
		}
	}
}
