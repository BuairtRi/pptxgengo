package deckproject

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func decodeSemanticDecisions(raw []byte, file, schema, action string) (GanttSemanticDecisions, error) {
	var out GanttSemanticDecisions
	if len(raw) > 1<<20 {
		return out, fmt.Errorf("semantic decisions exceed 1 MiB")
	}
	d := yaml.NewDecoder(bytes.NewReader(raw))
	var doc, extra yaml.Node
	if e := d.Decode(&doc); e != nil {
		return out, e
	}
	if len(doc.Content) != 1 {
		return out, fmt.Errorf("empty semantic decisions")
	}
	if e := d.Decode(&extra); e != io.EOF {
		return out, fmt.Errorf("semantic decisions require one document")
	}
	p := &Project{SourcePath: file, Positions: map[string]Position{}}
	v, e := p.yamlValue(doc.Content[0], "", 0)
	if e != nil {
		return out, e
	}
	if e = p.shapeType(v, reflect.TypeOf(out), ""); e != nil {
		return out, e
	}
	if e = strictInto(v, &out); e != nil {
		return out, e
	}
	if out.Schema != schema || !shaPattern.MatchString(out.ReportSHA256) || strings.TrimSpace(out.Actor) == "" || len(out.Actor) > 256 || strings.TrimSpace(out.Reason) == "" || len(out.Reason) > 4096 || len(out.Decisions) < 1 || len(out.Decisions) > 500 {
		return out, fmt.Errorf("invalid semantic decisions")
	}
	seen := map[string]bool{}
	for _, d := range out.Decisions {
		if !shaPattern.MatchString(d.ProposalID) || seen[d.ProposalID] || d.Action != action && d.Action != "keep_source" || strings.TrimSpace(d.Reason) == "" || len(d.Reason) > 4096 {
			return out, fmt.Errorf("invalid or duplicate semantic decision")
		}
		seen[d.ProposalID] = true
	}
	return out, nil
}
func semanticEvidenceCandidate(p *Project, packet *TextReviewPacket, slideID, operation, actor, reason string, t LocalTemplate, report any, raw []byte, bundle, engine string, apply bool) (CompositionResult, error) {
	var empty CompositionResult
	fresh, b, e := verifiedGanttPacket(p, packet, bundle, engine)
	if e != nil {
		return empty, e
	}
	packet = fresh
	reportRaw := canonical(report)
	evidence := struct {
		Report                 any    `json:"report"`
		ReportSHA256           string `json:"report_sha256"`
		RetainedPPTX           string `json:"retained_pptx"`
		RetainedReport         string `json:"retained_report"`
		RetainedDecisions      string `json:"retained_decisions"`
		RetainedGeometryReport string `json:"retained_geometry_report"`
		Scope                  string `json:"scope"`
	}{report, digest(reportRaw), "assets/objects/sha256/" + digest(packet.Edited), "assets/objects/sha256/" + digest(reportRaw), "assets/objects/sha256/" + digest(raw), "assets/objects/sha256/" + packet.ReportSHA256, "partial explicit semantic adoption; unselected/native formatting/text/structure changes remain retained, not synchronized"}
	_, lock, e := ReadLock(p)
	if e != nil {
		return empty, e
	}
	guards := map[string][]byte{p.Document.Toolchain.Lockfile: lock}
	for name, data := range b.files {
		guards["builds/"+b.Receipt.BuildID+"/"+name] = data
	}
	return compositionCandidate(p, slideID, operation, actor, reason, t, bundle, engine, apply, evidence, map[string][]byte{evidence.RetainedPPTX: packet.Edited, evidence.RetainedReport: reportRaw, evidence.RetainedDecisions: raw, evidence.RetainedGeometryReport: packet.files["report.json"]}, guards)
}
