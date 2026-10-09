package deckproject

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func commentNode(t *testing.T, source string) *yaml.Node {
	t.Helper()
	var node yaml.Node
	if e := yaml.Unmarshal([]byte(source), &node); e != nil {
		t.Fatal(e)
	}
	return node.Content[0]
}
func TestDiagramKeyedCommentsNestedRecordReorder(t *testing.T) {
	old := commentNode(t, `keys:
  /groups/0/tasks: [alpha, beta]
arguments:
  groups:
    - tasks:
        - name: Alpha
          value: 1
        - name: Beta
          value: 2
`)
	next := commentNode(t, `keys:
  /groups/0/tasks: [beta, new, alpha]
arguments:
  groups:
    - tasks:
        - name: Updated beta
          value: 3
        - name: New record
          value: 4
        - name: Updated alpha
          value: 5
`)
	before := diagramCommentPointer(mappingNode(old, "arguments"), "/groups/0/tasks")
	before.Content[0].HeadComment = "Alpha belongs to explicit authored alpha key"
	mappingNode(before.Content[0], "value").LineComment = "Retain measurement rationale"
	preserveDiagramKeyedArgumentComments(old, next)
	after := diagramCommentPointer(mappingNode(next, "arguments"), "/groups/0/tasks")
	if after.Content[2].HeadComment != before.Content[0].HeadComment || mappingNode(after.Content[2], "value").LineComment != "Retain measurement rationale" {
		t.Fatal("nested identity comments lost")
	}
	if after.Content[0].HeadComment != "" || after.Content[1].HeadComment != "" {
		t.Fatal("comment assigned to unrelated records")
	}
}
func TestDiagramKeyedCommentsStaffingObservationMatrix(t *testing.T) {
	old := commentNode(t, `definition: {scope: shared, id: wmds/component/teamcurve}
keys: {series: [delivery, advisory], at: [start, review, close]}
arguments:
  at: [0, 0.5, 1]
  series:
    - {name: Delivery, values: [1, 2, 3]}
    - {name: Advisory, values: [4, 5, 6]}
`)
	next := commentNode(t, `definition: {scope: shared, id: wmds/component/teamcurve}
keys: {series: [advisory, delivery], at: [close, start, review]}
arguments:
  at: [0, 0.5, 1]
  series:
    - {name: Advisory, values: [60, 40, 50]}
    - {name: Delivery, values: [30, 10, 20]}
`)
	series := mappingNode(mappingNode(old, "arguments"), "series")
	mappingNode(series.Content[0], "values").Content[1].LineComment = "Delivery review observation evidence"
	at := mappingNode(mappingNode(old, "arguments"), "at")
	at.Content[1].LineComment = "Review phase boundary rationale"
	preserveDiagramKeyedArgumentComments(old, next)
	series = mappingNode(mappingNode(next, "arguments"), "series")
	values := mappingNode(series.Content[1], "values")
	if values.Content[2].LineComment != "Delivery review observation evidence" {
		t.Fatal("observation did not follow both identities")
	}
	if mappingNode(series.Content[0], "values").Content[2].LineComment != "" {
		t.Fatal("observation evidence assigned to other series")
	}
	if mappingNode(mappingNode(next, "arguments"), "at").Content[2].LineComment != "Review phase boundary rationale" {
		t.Fatal("point math metadata lost")
	}
}
func TestDiagramKeyedCommentsAmbiguousKeysAndMissingOldKeysRefuseGuess(t *testing.T) {
	for _, keySource := range []string{"keys: {items: [same, same]}\n", ""} {
		old := commentNode(t, keySource+"arguments: {items: [{label: Old}, {label: Other}]}\n")
		next := commentNode(t, "keys: {items: [same, next]}\narguments: {items: [{label: Changed}, {label: New}]}\n")
		mappingNode(mappingNode(old, "arguments"), "items").Content[0].HeadComment = "Only exact identity may retain this"
		preserveDiagramKeyedArgumentComments(old, next)
		for _, record := range mappingNode(mappingNode(next, "arguments"), "items").Content {
			if record.HeadComment != "" {
				t.Fatal("ambiguous or missing identity guessed")
			}
		}
	}
}

func TestDiagramKeyedCommentsDescendantsFollowAncestorIdentity(t *testing.T) {
	old := commentNode(t, `keys: {groups: [delivery, advisory], /groups/0/tasks: [implement], /groups/1/tasks: [review]}
arguments:
 groups:
  - {tasks: [{label: Implement, value: 1}]}
  - {tasks: [{label: Review, value: 2}]}
`)
	next := commentNode(t, `keys: {groups: [advisory, delivery], /groups/0/tasks: [review], /groups/1/tasks: [implement]}
arguments:
 groups:
  - {tasks: [{label: Review updated, value: 20}]}
  - {tasks: [{label: Implement updated, value: 10}]}
`)
	before := diagramCommentPointer(mappingNode(old, "arguments"), "groups/0/tasks")
	before.Content[0].HeadComment = "Delivery implementation only"
	preserveDiagramKeyedArgumentComments(old, next)
	after := diagramCommentPointer(mappingNode(next, "arguments"), "groups/1/tasks")
	if after.Content[0].HeadComment != before.Content[0].HeadComment {
		t.Fatal("comment did not follow parent and child identities")
	}
	if diagramCommentPointer(mappingNode(next, "arguments"), "groups/0/tasks").Content[0].HeadComment != "" {
		t.Fatal("comment assigned to former parent index")
	}
}
