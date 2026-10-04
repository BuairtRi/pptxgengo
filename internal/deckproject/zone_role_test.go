package deckproject

import (
	"os"
	"strings"
	"testing"

	"github.com/buairtri/pptxgengo/internal/wmdesign"
)

func TestLoadRejectsUnrenderedZoneRolesWithActionablePath(t *testing.T) {
	for _, role := range []string{"title", "slide-titel", "subhead", "unknown-purpose"} {
		t.Run(role, func(t *testing.T) {
			p := example(t)
			raw := strings.Replace(string(p.Raw), "role: slide-title", "role: "+role, 1)
			if err := os.WriteFile(p.SourcePath, []byte(raw), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(p.SourcePath)
			if err == nil || !strings.Contains(err.Error(), "/local_templates/editorial-photo/zones/title/role") || !strings.Contains(err.Error(), "slide-title") || !strings.Contains(err.Error(), "bind this zone explicitly to a node") {
				t.Fatalf("missing actionable role error: %v", err)
			}
		})
	}
}

func TestZoneRolesPreserveExplicitCustomBindingsAndFrameTitle(t *testing.T) {
	p := example(t)
	local := p.Document.LocalTemplates["editorial-photo"]
	for _, role := range []string{"body-copy", "customer-qualification", "title"} {
		z := local.Zones["summary"]
		z.Role = role
		local.Zones["summary"] = z
		p.Document.LocalTemplates["editorial-photo"] = local
		if err := p.validate(); err != nil {
			t.Fatalf("explicitly bound semantic role %q rejected: %v", role, err)
		}
	}
	compiled, err := Compile(p, bundle(t), wmdesign.CandidateEngine)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := compiled.Document.Slides[1].Title, p.Document.Slides[1].Values["title"]; got != want {
		t.Fatalf("frame title lost: got %q want %q", got, want)
	}
	// Binding consumption is recursive through groups and component arrays.
	z := local.Zones["summary"]
	z.Role = "component-caption"
	local.Zones["summary"] = z
	local.Nodes[0].Nodes[0].Kind = "component"
	local.Nodes[0].Nodes[0].Text = nil
	local.Nodes[0].Nodes[0].Definition = &Reference{Scope: "shared", ID: "card"}
	local.Nodes[0].Nodes[0].Arguments = map[string]any{"items": []any{map[string]any{"text": map[string]any{"binding": "summary"}}}}
	p.Document.LocalTemplates["editorial-photo"] = local
	if err := p.validate(); err != nil {
		t.Fatalf("nested explicit component binding rejected: %v", err)
	}
}

func TestUnboundOptionalZoneAlsoFailsAndReservedFrameRolesRemainValid(t *testing.T) {
	p := example(t)
	local := p.Document.LocalTemplates["editorial-photo"]
	local.Zones["citation"] = Zone{Role: "source", Required: false, Schema: map[string]any{"type": "string"}}
	local.Zones["tabs"] = Zone{Role: "nav", Required: false, Schema: map[string]any{"type": "object"}}
	p.Document.LocalTemplates["editorial-photo"] = local
	if err := p.validate(); err != nil {
		t.Fatalf("reserved frame roles rejected: %v", err)
	}
	local.Zones["unused"] = Zone{Role: "supporting-copy", Required: false, Schema: map[string]any{"type": "string"}}
	p.Document.LocalTemplates["editorial-photo"] = local
	if err := p.validate(); err == nil || !strings.Contains(err.Error(), "/zones/unused/role") || !strings.Contains(err.Error(), "no renderer binding") {
		t.Fatalf("unbound optional zone accepted: %v", err)
	}
}
