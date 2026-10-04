package deckproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPageSpecExplainsUnquotedFlowMapCommas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.yaml")
	for _, raw := range []string{
		"title: Decision\nitems: [{lead: Owner, text: Confirm the owner, date, and next step}]\nrelationship: parallel\n",
		"title: Decision\ncallout: {value: 3, label: Reviews, source: Client interview, Sept 2026}\n",
	} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadPageSpec(path)
		if err == nil || !strings.Contains(err.Error(), "possible unquoted comma") || !strings.Contains(err.Error(), "quote the complete comma-containing value") {
			t.Fatalf("missing actionable flow-map diagnostic: %v", err)
		}
	}
	if err := os.WriteFile(path, []byte("title: Decision\nitems: [{lead: Owner, text: \"Confirm the owner, date, and next step\"}]\nrelationship: parallel\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPageSpec(path); err != nil {
		t.Fatalf("quoted flow-map value rejected: %v", err)
	}
}

func TestSlidePatchExplainsUnquotedFlowMapCommas(t *testing.T) {
	bad := []byte("slide-id:\n  bindings: {/title: /title, /body: /body, the owner and date}\n")
	if _, err := DecodeSlideEdits(bad, "patch.yaml"); err == nil || !strings.Contains(err.Error(), "possible unquoted comma") {
		t.Fatalf("missing actionable patch diagnostic: %v", err)
	}
	good := []byte("slide-id:\n  content: {headline: \"Confirm the owner, date, and next step\"}\n")
	if _, err := DecodeSlideEdits(good, "patch.yaml"); err != nil {
		t.Fatalf("quoted patch content rejected: %v", err)
	}
}

func TestFlowMapDiagnosticAllowsExplicitNullPhraseValue(t *testing.T) {
	if err := flowMapCommaDiagnostic([]byte(`values: {"optional note": null}`), "page.yaml"); err != nil {
		t.Fatalf("explicit null was misdiagnosed as an unquoted comma: %v", err)
	}
}
