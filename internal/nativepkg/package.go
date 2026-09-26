package nativepkg

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Slide struct {
	Number                   int       `json:"source_slide_number"`
	Part                     string    `json:"part"`
	PresentationID           *Node     `json:"presentation_id"`
	PresentationRelationship *Node     `json:"presentation_relationship"`
	Scene                    *Node     `json:"scene"`
	Relationships            *Node     `json:"relationships"`
	ReferenceOnly            bool      `json:"reference_only,omitempty"`
	Bindings                 []Binding `json:"bindings,omitempty"`
}
type Manifest struct {
	Schema                    string   `json:"schema"`
	Source                    string   `json:"source"`
	SourceSHA256              string   `json:"source_sha256"`
	Presentation              *Node    `json:"presentation"`
	PresentationRelationships *Node    `json:"presentation_relationships"`
	ContentTypes              *Node    `json:"content_types"`
	Slides                    []string `json:"slides"`
	Mode                      string   `json:"mode"`
}

func relpart(p string) string {
	if p == "" {
		return "_rels/.rels"
	}
	return path.Join(path.Dir(p), "_rels", path.Base(p)+".rels")
}
func resolve(p, t string) string {
	if strings.HasPrefix(t, "/") {
		return strings.TrimPrefix(path.Clean(t), "/")
	}
	return path.Clean(path.Join(path.Dir(p), t))
}
func readZip(file string) (map[string][]byte, error) {
	z, e := zip.OpenReader(file)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	out := map[string][]byte{}
	for _, f := range z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		b, e := io.ReadAll(r)
		r.Close()
		if e != nil {
			return nil, e
		}
		out[f.Name] = b
	}
	return out, nil
}
func writeJSON(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(p, append(b, '\n'), 0644)
}
func readJSON(p string, v any) error {
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func Extract(source, out string, numbers []int) error {
	target := out
	if e := requireAbsent(target, "output directory"); e != nil {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(target), 0755); e != nil {
		return e
	}
	stage, e := os.MkdirTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp-")
	if e != nil {
		return e
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(stage)
		}
	}()
	out = stage

	parts, e := readZip(source)
	if e != nil {
		return e
	}
	get := func(p string) *Node {
		n, err := Parse(parts[p])
		if err != nil {
			e = fmt.Errorf("%s: %w", p, err)
		}
		return n
	}
	pres := get("ppt/presentation.xml")
	rels := get("ppt/_rels/presentation.xml.rels")
	ct := get("[Content_Types].xml")
	if e != nil {
		return e
	}
	ids := pres.Child("sldIdLst")
	if ids == nil {
		return fmt.Errorf("no slide list")
	}
	var idlist []*Node
	for _, n := range ids.Children {
		if local(n.Name) == "sldId" {
			idlist = append(idlist, n)
		}
	}
	byID := map[string]*Node{}
	for _, r := range rels.Children {
		if r.Name != "" {
			byID[r.Attr("Id")] = r
		}
	}
	wanted := map[int]bool{}
	partNumbers := map[string]int{}
	for i, id := range idlist {
		if r := byID[id.Attr("r:id")]; r != nil {
			partNumbers[resolve("ppt/presentation.xml", r.Attr("Target"))] = i + 1
		}
	}
	for _, n := range numbers {
		if n < 1 || n > len(idlist) {
			return fmt.Errorf("slide %d outside range 1..%d", n, len(idlist))
		}
		wanted[n] = true
	}
	if e = os.MkdirAll(filepath.Join(out, "slides"), 0755); e != nil {
		return e
	}
	src, _ := os.ReadFile(source)
	h := sha256.Sum256(src)
	m := Manifest{Schema: "pptxgengo.native-scene.v2", Source: filepath.Base(source), SourceSHA256: hex.EncodeToString(h[:]), Presentation: pres, PresentationRelationships: rels, ContentTypes: ct, Mode: "source-derived parameterized native objects; binding-driven text/geometry/style regeneration; assets/masters/layouts retained; not semantic layout inference"}
	queue := append([]int(nil), numbers...)
	queued := map[int]bool{}
	for _, n := range queue {
		queued[n] = true
	}
	for qi := 0; qi < len(queue); qi++ {
		n := queue[qi]
		id := idlist[n-1]
		r := byID[id.Attr("r:id")]
		if r == nil {
			return fmt.Errorf("missing slide relation")
		}
		part := resolve("ppt/presentation.xml", r.Attr("Target"))
		scene := get(part)
		sr := get(relpart(part))
		if e != nil {
			return e
		}
		s := Slide{Number: n, Part: part, PresentationID: id, PresentationRelationship: r, Scene: scene, Relationships: sr, ReferenceOnly: !wanted[n]}
		s.Bindings = ExtractBindings(scene)
		for _, rel := range sr.Children {
			if strings.HasSuffix(rel.Attr("Type"), "/slide") && rel.Attr("TargetMode") != "External" {
				dest := partNumbers[resolve(part, rel.Attr("Target"))]
				if dest == 0 {
					return fmt.Errorf("unknown internal slide target")
				}
				if !queued[dest] {
					queue = append(queue, dest)
					queued[dest] = true
				}
			}
		}
		name := fmt.Sprintf("slides/uhg-%03d.json", n)
		if e = writeJSON(filepath.Join(out, name), s); e != nil {
			return e
		}
		m.Slides = append(m.Slides, name)
	}
	// Slide bodies and their relations never enter the resource store. Build
	// must serialize their JSON scenes; it cannot copy source slide XML.
	for p, b := range parts {
		if strings.HasPrefix(p, "ppt/slides/") || p == "ppt/presentation.xml" || p == "ppt/_rels/presentation.xml.rels" || p == "[Content_Types].xml" {
			continue
		}
		dst := filepath.Join(out, "resources", filepath.FromSlash(p))
		if e = os.MkdirAll(filepath.Dir(dst), 0755); e != nil {
			return e
		}
		if e = os.WriteFile(dst, b, 0644); e != nil {
			return e
		}
	}
	if e = writeJSON(filepath.Join(out, "manifest.json"), m); e != nil {
		return e
	}
	if e = requireAbsent(target, "output directory"); e != nil {
		return e
	}
	if e = os.Rename(stage, target); e != nil {
		return e
	}
	committed = true
	return nil
}

type BuildReport struct {
	Mode               string   `json:"mode"`
	Slides             []int    `json:"source_slides"`
	Regenerated        []string `json:"regenerated_slide_parts"`
	Retained           int      `json:"retained_resource_parts"`
	Frozen             bool     `json:"source_slide_numbers_frozen"`
	HiddenDependencies []int    `json:"hidden_link_destinations,omitempty"`
}

func Build(project, out string, numbers []int, freeze bool) error {
	if e := requireAbsent(out, "output"); e != nil {
		return e
	}
	if e := requireAbsent(out+".build.json", "build report"); e != nil {
		return e
	}
	var m Manifest
	if e := readJSON(filepath.Join(project, "manifest.json"), &m); e != nil {
		return e
	}
	if m.Schema != "pptxgengo.native-scene.v1" && m.Schema != "pptxgengo.native-scene.v2" {
		return fmt.Errorf("unsupported scene schema")
	}
	all := map[int]Slide{}
	for _, p := range m.Slides {
		var s Slide
		if e := readJSON(filepath.Join(project, p), &s); e != nil {
			return e
		}
		all[s.Number] = s
	}
	if len(numbers) == 0 {
		for _, p := range m.Slides {
			var s Slide
			_ = readJSON(filepath.Join(project, p), &s)
			if !s.ReferenceOnly {
				numbers = append(numbers, s.Number)
			}
		}
	}
	parts := map[string][]byte{}
	base := filepath.Join(project, "resources")
	e := filepath.WalkDir(base, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		r, e := filepath.Rel(base, p)
		if e != nil {
			return e
		}
		b, e := os.ReadFile(p)
		parts[filepath.ToSlash(r)] = b
		return e
	})
	if e != nil {
		return e
	}
	ids := m.Presentation.Child("sldIdLst")
	ids.Children = nil
	var keep []*Node
	for _, r := range m.PresentationRelationships.Children {
		if !strings.HasSuffix(r.Attr("Type"), "/slide") {
			keep = append(keep, r)
		}
	}
	m.PresentationRelationships.Children = keep
	// Source sections contain IDs of slides outside the selected output.
	var cleanSections func(*Node)
	cleanSections = func(n *Node) {
		var c []*Node
		for _, x := range n.Children {
			if local(x.Name) == "sectionLst" {
				continue
			}
			cleanSections(x)
			c = append(c, x)
		}
		n.Children = c
	}
	cleanSections(m.Presentation)
	report := BuildReport{Mode: m.Mode, Frozen: freeze}
	visible := map[int]bool{}
	byPart := map[string]int{}
	for _, n := range numbers {
		visible[n] = true
	}
	for n, s := range all {
		byPart[s.Part] = n
	}
	queue := append([]int(nil), numbers...)
	enqueued := map[int]bool{}
	for _, n := range queue {
		enqueued[n] = true
	}
	for i := 0; i < len(queue); i++ {
		s, ok := all[queue[i]]
		if !ok {
			return fmt.Errorf("slide %d not extracted", queue[i])
		}
		for _, rel := range s.Relationships.Children {
			if strings.HasSuffix(rel.Attr("Type"), "/slide") && rel.Attr("TargetMode") != "External" {
				n := byPart[resolve(s.Part, rel.Attr("Target"))]
				if n == 0 {
					return fmt.Errorf("missing linked slide; re-extract project with dependency support")
				}
				if !enqueued[n] {
					queue = append(queue, n)
					enqueued[n] = true
				}
			}
		}
	}
	seen := map[int]bool{}
	for _, number := range queue {
		if seen[number] {
			return fmt.Errorf("duplicate slide %d", number)
		}
		seen[number] = true
		s, ok := all[number]
		if !ok {
			return fmt.Errorf("slide %d not extracted", number)
		}
		if m.Schema == "pptxgengo.native-scene.v2" {
			if e := ApplyBindings(s.Scene, s.Bindings); e != nil {
				return fmt.Errorf("slide %d: %w", number, e)
			}
		}
		if !visible[number] {
			s.Scene.SetAttr("show", "0")
			report.HiddenDependencies = append(report.HiddenDependencies, number)
		}
		if freeze {
			s.Scene.Walk(func(n *Node) {
				if local(n.Name) == "fld" && n.Attr("type") == "slidenum" {
					n.Name = "a:r"
					n.Attrs = nil
					var c []*Node
					for _, x := range n.Children {
						if local(x.Name) != "pPr" {
							c = append(c, x)
						}
					}
					n.Children = c
				}
			})
		}
		ids.Children = append(ids.Children, s.PresentationID)
		m.PresentationRelationships.Children = append(m.PresentationRelationships.Children, s.PresentationRelationship)
		parts[s.Part] = s.Scene.XML()
		parts[relpart(s.Part)] = s.Relationships.XML()
		if visible[number] {
			report.Slides = append(report.Slides, number)
		}
		report.Regenerated = append(report.Regenerated, s.Part)
	}
	parts["ppt/presentation.xml"] = m.Presentation.XML()
	parts["ppt/_rels/presentation.xml.rels"] = m.PresentationRelationships.XML()
	// Traverse OPC relationships so no unused notes with dangling slide refs
	// are carried into the result. Theme/layout cycles are handled by seen.
	used := map[string]bool{}
	var visit func(string) error
	visit = func(p string) error {
		if used[p] {
			return nil
		}
		if p != "" {
			if _, ok := parts[p]; !ok {
				return fmt.Errorf("missing required package part %s", p)
			}
			used[p] = true
		}
		rp := relpart(p)
		b, ok := parts[rp]
		if !ok {
			return nil
		}
		used[rp] = true
		r, e := Parse(b)
		if e != nil {
			return e
		}
		for _, rel := range r.Children {
			if rel.Name == "" || rel.Attr("TargetMode") == "External" {
				continue
			}
			if t := rel.Attr("Target"); t != "" {
				if e = visit(resolve(p, t)); e != nil {
					return e
				}
			}
		}
		return nil
	}
	if e = visit(""); e != nil {
		return e
	}
	var types []*Node
	for _, n := range m.ContentTypes.Children {
		if local(n.Name) == "Override" && !used[strings.TrimPrefix(n.Attr("PartName"), "/")] {
			continue
		}
		types = append(types, n)
	}
	m.ContentTypes.Children = types
	parts["[Content_Types].xml"] = m.ContentTypes.XML()
	used["[Content_Types].xml"] = true
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	names := make([]string, 0, len(used))
	for p := range used {
		names = append(names, p)
	}
	sort.Strings(names)
	for _, p := range names {
		w, e := zw.Create(p)
		if e != nil {
			return e
		}
		if _, e = w.Write(parts[p]); e != nil {
			return e
		}
	}
	if e = zw.Close(); e != nil {
		return e
	}
	report.Retained = len(used) - 2*len(queue) - 3
	return writeBuildPair(out, buf.Bytes(), report)
}

func requireAbsent(p, label string) error {
	_, e := os.Stat(p)
	if e == nil {
		return fmt.Errorf("%s already exists: %s", label, p)
	}
	if !os.IsNotExist(e) {
		return e
	}
	return nil
}

// writeBuildPair stages both output files in the target directory. It commits
// the report first, then the PPTX, and removes the committed report if the
// PPTX rename fails. A process crash between the two renames is not a strict
// cross-file atomic transaction.
func writeBuildPair(out string, pptx []byte, report BuildReport) error {
	reportBytes, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		return e
	}
	reportBytes = append(reportBytes, '\n')
	if e = os.MkdirAll(filepath.Dir(out), 0755); e != nil {
		return e
	}
	if e = requireAbsent(out, "output"); e != nil {
		return e
	}
	reportPath := out + ".build.json"
	if e = requireAbsent(reportPath, "build report"); e != nil {
		return e
	}
	pptxFile, e := os.CreateTemp(filepath.Dir(out), "."+filepath.Base(out)+".tmp-")
	if e != nil {
		return e
	}
	pptxTemp := pptxFile.Name()
	if e = pptxFile.Close(); e != nil {
		_ = os.Remove(pptxTemp)
		return e
	}
	reportFile, e := os.CreateTemp(filepath.Dir(out), "."+filepath.Base(reportPath)+".tmp-")
	if e != nil {
		_ = os.Remove(pptxTemp)
		return e
	}
	reportTemp := reportFile.Name()
	if e = reportFile.Close(); e != nil {
		_ = os.Remove(pptxTemp)
		_ = os.Remove(reportTemp)
		return e
	}
	committedReport := false
	defer func() {
		_ = os.Remove(pptxTemp)
		_ = os.Remove(reportTemp)
		if committedReport {
			_ = os.Remove(reportPath)
		}
	}()
	if e = os.WriteFile(pptxTemp, pptx, 0644); e != nil {
		return e
	}
	if e = os.WriteFile(reportTemp, reportBytes, 0644); e != nil {
		return e
	}
	if e = os.Chmod(pptxTemp, 0644); e != nil {
		return e
	}
	if e = os.Chmod(reportTemp, 0644); e != nil {
		return e
	}
	if e = requireAbsent(out, "output"); e != nil {
		return e
	}
	if e = requireAbsent(reportPath, "build report"); e != nil {
		return e
	}
	if e = os.Rename(reportTemp, reportPath); e != nil {
		return e
	}
	committedReport = true
	if e = os.Rename(pptxTemp, out); e != nil {
		return e
	}
	committedReport = false
	return nil
}
func Numbers(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	var out []int
	for _, p := range strings.Split(s, ",") {
		n, e := strconv.Atoi(strings.TrimSpace(p))
		if e != nil {
			return nil, e
		}
		out = append(out, n)
	}
	return out, nil
}
