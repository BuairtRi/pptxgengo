package wmdesign

import "testing"

func TestLibraryV10RoadForkDiscoveryPreservesTopology(t *testing.T) {
	catalog, err := LibraryCatalog(v10IntakeBundle(), "")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"road-fork/parallel", "road-fork/decision"} {
		t.Run(key, func(t *testing.T) {
			def := discoveryPinnedTemplate(t, catalog, key)
			d := def.Discovery
			if !discoveryHas(d.ComponentTypes, "roadfork") || !discoveryHas(d.ContentRoles, "relationship") || !discoveryHas(d.ContentRoles, "sequence-item") || !discoveryHas(d.VisualForms, "diagram") {
				t.Fatalf("branching roadmap lost its visual roles: %+v", d)
			}
			for _, structure := range []string{"timeline", "sequence", "network"} {
				if !discoveryHas(d.Structures, structure) {
					t.Fatalf("missing %s", structure)
				}
			}
			mode := "parallel"
			if key == "road-fork/decision" {
				mode = "decision"
			}
			if len(d.Relationships) != 1 || d.Relationships[0].Kind != mode || len(d.Relationships[0].SourcePointers) != 3 {
				t.Fatalf("mode relationship lost: %+v", d.Relationships)
			}
			branches, trunk, milestones := 0, 0, 0
			for _, group := range d.Groups {
				switch group.SourcePointer {
				case "/body/0/branches":
					if group.Role != "relationship" || group.Scope != "primary" || group.ExactCount != 3 {
						t.Fatalf("branches confused with nested milestones: %+v", group)
					}
					branches++
				case "/body/0/trunk":
					if group.Role != "sequence-item" || group.Scope != "primary" || group.ExactCount != 2 {
						t.Fatalf("trunk sequence lost: %+v", group)
					}
					trunk++
				default:
					if group.ComponentType == "roadfork" && group.Role == "sequence-item" && group.Scope == "nested" {
						milestones++
					}
				}
			}
			if branches != 1 || trunk != 1 || milestones != 3 {
				t.Fatalf("incomplete groups: branches=%d trunk=%d milestones=%d", branches, trunk, milestones)
			}
		})
	}
}
