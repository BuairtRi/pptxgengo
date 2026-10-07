package deckproject

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"
)

const NativeLineageSchema = "pptxgengo.native-lineage.v1"
const lineagePML = "http://schemas.openxmlformats.org/presentationml/2006/main"
const lineageRML = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const lineageOPC = "http://schemas.openxmlformats.org/package/2006/relationships"
const lineageCT = "http://schemas.openxmlformats.org/package/2006/content-types"
const lineageTagType = lineageRML + "/tags"
const lineageTagContent = "application/vnd.openxmlformats-officedocument.presentationml.tags+xml"
const lineageMaxPart = 64 << 20
const lineageMaxPackage = 512 << 20

type NativeLineage struct {
	Schema                string            `json:"schema"`
	DeckToken             string            `json:"deck_token"`
	BuildToken            string            `json:"build_token"`
	SourceSemanticSHA256  string            `json:"source_semantic_sha256"`
	LockSHA256            string            `json:"lock_sha256"`
	UnstampedNativeSHA256 string            `json:"unstamped_native_sha256"`
	Slides                map[string]string `json:"slide_tokens"` // token -> authored slide ID
}

type NativeLineageObject struct {
	ShapeToken  string `json:"shape_token,omitempty"`
	SlideToken  string `json:"slide_token,omitempty"`
	NativePart  string `json:"native_part"`
	NativeID    string `json:"native_id"`
	NativeName  string `json:"native_name"`
	Kind        string `json:"kind"`
	ParentToken string `json:"parent_token,omitempty"`
	shape       *xmlNode
}
type NativeLineageIssue struct {
	Kind       string `json:"kind"`
	NativePart string `json:"native_part,omitempty"`
	ShapeToken string `json:"shape_token,omitempty"`
	SlideToken string `json:"slide_token,omitempty"`
	Detail     string `json:"detail"`
}
type NativeLineageInspection struct {
	Schema     string                `json:"schema"`
	DeckToken  string                `json:"deck_token"`
	BuildToken string                `json:"build_token"`
	Objects    []NativeLineageObject `json:"objects"`
	Issues     []NativeLineageIssue  `json:"issues"`
}

func lineageToken(v any) string { return strings.ToUpper(digest(canonical(v))) }
func validLineageToken(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// A bounded, non-extracting package reader. Reject duplicate/unsafe entries and
// enforce actual decompressed limits, including parts unrelated to our tags.
type lineagePackage struct {
	files             map[string]*zip.File
	order             []*zip.File
	total             int64
	cache             map[string][]byte
	contentTypes      map[string]string
	contentDefaults   map[string]string
	relationshipCache map[string]map[string]lineageRelationship
}

func openLineagePackage(data []byte) (*lineagePackage, error) {
	if len(data) > lineageMaxPackage {
		return nil, fmt.Errorf("native package exceeds 512 MiB")
	}
	z, e := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if e != nil {
		return nil, e
	}
	if len(z.File) > 40000 {
		return nil, fmt.Errorf("native package has too many entries")
	}
	p := &lineagePackage{files: map[string]*zip.File{}, order: z.File, cache: map[string][]byte{}, relationshipCache: map[string]map[string]lineageRelationship{}}
	var size uint64
	for _, f := range z.File {
		n := f.Name
		if n == "" || strings.ContainsAny(n, "\\\x00") || strings.HasPrefix(n, "/") || path.Clean(n) != strings.TrimSuffix(n, "/") || strings.HasPrefix(n, "../") || strings.Contains(n, ":") {
			return nil, fmt.Errorf("unsafe native package part %q", n)
		}
		if _, ok := p.files[n]; ok {
			return nil, fmt.Errorf("duplicate native package part %q", n)
		}
		if !f.Mode().IsRegular() && !f.FileInfo().IsDir() {
			return nil, fmt.Errorf("nonregular native package part %q", n)
		}
		if f.Flags&1 != 0 || f.UncompressedSize64 > lineageMaxPart {
			return nil, fmt.Errorf("encrypted or oversized native part %q", n)
		}
		size += f.UncompressedSize64
		if size > lineageMaxPackage {
			return nil, fmt.Errorf("expanded native package exceeds 512 MiB")
		}
		p.files[n] = f
	}
	return p, nil
}
func (p *lineagePackage) read(name string) ([]byte, error) {
	if b, ok := p.cache[name]; ok {
		return b, nil
	}
	f := p.files[name]
	if f == nil {
		return nil, fmt.Errorf("missing native part %s", name)
	}
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, lineageMaxPart+1))
	if e != nil {
		return nil, e
	}
	if len(b) > lineageMaxPart {
		return nil, fmt.Errorf("native part exceeds 64 MiB: %s", name)
	}
	p.total += int64(len(b))
	if p.total > lineageMaxPackage {
		return nil, fmt.Errorf("expanded native package exceeds 512 MiB")
	}
	p.cache[name] = b
	return b, nil
}
func (p *lineagePackage) tree(name string) (*xmlNode, error) {
	b, e := p.read(name)
	if e != nil {
		return nil, e
	}
	return readLineageXML(b)
}
func readLineageXML(b []byte) (*xmlNode, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	depth, nodes := 0, 0
	for {
		t, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		switch t.(type) {
		case xml.StartElement:
			start := t.(xml.StartElement)
			seen := map[xml.Name]bool{}
			for _, a := range start.Attr {
				if seen[a.Name] {
					return nil, fmt.Errorf("duplicate native XML attribute")
				}
				seen[a.Name] = true
			}
			depth++
			nodes++
			if depth > 100 || nodes > 200000 {
				return nil, fmt.Errorf("native XML exceeds structural bounds")
			}
		case xml.EndElement:
			depth--
		case xml.Directive:
			return nil, fmt.Errorf("native XML directives forbidden")
		}
	}
	tree, e := readXML(bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	if len(tree.Children) != 1 || strings.TrimSpace(tree.Text) != "" {
		return nil, fmt.Errorf("native XML requires one root")
	}
	return tree.Children[0], nil
}
func lineageAttr(n *xmlNode, space, local string) string {
	for _, a := range n.Attrs {
		if a.Name.Space == space && a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}
func lineageChild(n *xmlNode, space, local string) *xmlNode {
	var found *xmlNode
	for _, c := range n.Children {
		if c.Name.Space == space && c.Name.Local == local {
			if found != nil {
				return nil
			}
			found = c
		}
	}
	return found
}
func lineageRelPart(part string) string {
	return path.Join(path.Dir(part), "_rels", path.Base(part)+".rels")
}
func lineageTarget(part, target string) (string, error) {
	if target == "" || strings.ContainsAny(target, "\\\x00?#:") || strings.HasPrefix(target, "/") {
		return "", fmt.Errorf("unsafe native relationship target %q", target)
	}
	p := path.Clean(path.Join(path.Dir(part), target))
	if p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return "", fmt.Errorf("native relationship escapes package")
	}
	return p, nil
}

type lineageRelationship struct{ ID, Type, Target, Mode string }

func (p *lineagePackage) relationships(part string) (map[string]lineageRelationship, error) {
	relpart := lineageRelPart(part)
	if part == "" {
		relpart = "_rels/.rels"
	}
	// Every shape on a slide shares this immutable relationship part. Parse
	// once per package, rather than once per shape (quadratic work on big slides).
	if cached, ok := p.relationshipCache[relpart]; ok {
		return cached, nil
	}
	tree, e := p.tree(relpart)
	if e != nil {
		return nil, e
	}
	if tree.Name != (xml.Name{Space: lineageOPC, Local: "Relationships"}) {
		return nil, fmt.Errorf("invalid relationships root")
	}
	out := map[string]lineageRelationship{}
	for _, n := range tree.Children {
		if n.Name != (xml.Name{Space: lineageOPC, Local: "Relationship"}) {
			return nil, fmt.Errorf("invalid relationship child")
		}
		r := lineageRelationship{ID: lineageAttr(n, "", "Id"), Type: lineageAttr(n, "", "Type"), Target: lineageAttr(n, "", "Target"), Mode: lineageAttr(n, "", "TargetMode")}
		if r.ID == "" || out[r.ID].ID != "" {
			return nil, fmt.Errorf("missing or duplicate native relationship ID")
		}
		out[r.ID] = r
	}
	p.relationshipCache[relpart] = out
	return out, nil
}
func (p *lineagePackage) tags(part string, owner *xmlNode) (map[string]string, error) {
	var list *xmlNode
	for _, c := range owner.Children {
		if c.Name == (xml.Name{Space: lineagePML, Local: "custDataLst"}) {
			if list != nil {
				return nil, fmt.Errorf("duplicate native customer data list")
			}
			list = c
		}
	}
	if list == nil {
		return map[string]string{}, nil
	}
	var ref string
	for _, c := range list.Children {
		if c.Name == (xml.Name{Space: lineagePML, Local: "tags"}) {
			if ref != "" {
				return nil, fmt.Errorf("multiple native tag references")
			}
			ref = lineageAttr(c, lineageRML, "id")
			if ref == "" {
				return nil, fmt.Errorf("missing tag relationship ID")
			}
		}
	}
	if ref == "" {
		return map[string]string{}, nil
	}
	rels, e := p.relationships(part)
	if e != nil {
		return nil, e
	}
	r := rels[ref]
	if r.Type != lineageTagType || (r.Mode != "" && r.Mode != "Internal") {
		return nil, fmt.Errorf("invalid/external native tag relationship")
	}
	target, e := lineageTarget(part, r.Target)
	if e != nil {
		return nil, e
	}
	contentType, e := p.contentType(target)
	if e != nil {
		return nil, e
	}
	if contentType != lineageTagContent {
		return nil, fmt.Errorf("invalid native tag part content type")
	}
	t, e := p.tree(target)
	if e != nil {
		return nil, e
	}
	if t.Name != (xml.Name{Space: lineagePML, Local: "tagLst"}) {
		return nil, fmt.Errorf("invalid native tag list")
	}
	out := map[string]string{}
	for _, c := range t.Children {
		if c.Name != (xml.Name{Space: lineagePML, Local: "tag"}) {
			return nil, fmt.Errorf("invalid native tag child")
		}
		name, val := lineageAttr(c, "", "name"), lineageAttr(c, "", "val")
		if name == "" || len(name) > 256 || len(val) > 4096 {
			return nil, fmt.Errorf("invalid native tag size")
		}
		name = strings.ToUpper(name)
		if _, ok := out[name]; ok {
			return nil, fmt.Errorf("duplicate native tag name %s", name)
		}
		out[name] = val
	}
	return out, nil
}

// XML spans support insertion without reserializing visible content, namespaces,
// formatting, or relationship attributes. Only generated baseline packages are stamped.
type lineageSpan struct {
	node                           *xmlNode
	start, startEnd, endStart, end int
	children                       []*lineageSpan
}

func lineageSpans(b []byte) (*lineageSpan, error) {
	if _, e := readLineageXML(b); e != nil {
		return nil, e
	}
	d := xml.NewDecoder(bytes.NewReader(b))
	var root *lineageSpan
	stack := []*lineageSpan{}
	for {
		before := int(d.InputOffset())
		t, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		after := int(d.InputOffset())
		switch t := t.(type) {
		case xml.StartElement:
			s := &lineageSpan{node: &xmlNode{Name: t.Name, Attrs: t.Attr}, start: before, startEnd: after}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, s)
			} else {
				root = s
			}
			stack = append(stack, s)
		case xml.EndElement:
			s := stack[len(stack)-1]
			s.endStart = before
			s.end = after
			stack = stack[:len(stack)-1]
		}
	}
	return root, nil
}

type lineagePatch struct {
	start, end int
	text       string
}

func lineageInsert(s *lineageSpan, child string, before map[string]bool, b []byte) (lineagePatch, error) {
	for _, c := range s.children {
		if c.node.Name.Space == lineagePML && c.node.Name.Local == "custDataLst" {
			return lineagePatch{}, fmt.Errorf("baseline already carries custom data")
		}
	}
	startTag := string(b[s.start:s.startEnd])
	if strings.HasSuffix(startTag, "/>") {
		i := strings.LastIndex(startTag, "/>")
		name := strings.Fields(strings.TrimPrefix(startTag, "<"))[0]
		name = strings.TrimSuffix(name, "/>")
		return lineagePatch{s.start, s.startEnd, startTag[:i] + ">" + child + "</" + name + ">"}, nil
	}
	offset := s.endStart
	for _, c := range s.children {
		if c.node.Name.Space == lineagePML && before[c.node.Name.Local] {
			offset = c.start
			break
		}
	}
	return lineagePatch{offset, offset, child}, nil
}
func lineageApply(b []byte, patches []lineagePatch) ([]byte, error) {
	sort.Slice(patches, func(i, j int) bool { return patches[i].start < patches[j].start })
	var out bytes.Buffer
	pos := 0
	for _, p := range patches {
		if p.start < pos || p.end < p.start || p.end > len(b) {
			return nil, fmt.Errorf("overlapping native lineage insertions")
		}
		out.Write(b[pos:p.start])
		out.WriteString(p.text)
		pos = p.end
	}
	out.Write(b[pos:])
	return out.Bytes(), nil
}
func lineageAppendRoot(b []byte, child string) ([]byte, error) {
	s, e := lineageSpans(b)
	if e != nil {
		return nil, e
	}
	if strings.HasSuffix(string(b[s.start:s.startEnd]), "/>") {
		return nil, fmt.Errorf("unexpected empty native package root")
	}
	return lineageApply(b, []lineagePatch{{s.endStart, s.endStart, child}})
}
func lineageCustomerData(relID string) string {
	return `<p:custDataLst xmlns:p="` + lineagePML + `" xmlns:r="` + lineageRML + `"><p:tags r:id="` + relID + `"/></p:custDataLst>`
}

// StampNativeLineage stamps project-generated bytes and returns an immutable
// sidecar map. Identity is deterministic for the source/lock/native baseline;
// execution BuildIDs remain in receipts. Tokens are uppercase SHA256 strings
// because PowerPoint's Tags API uppercases values.
func StampNativeLineage(data []byte, objects Objects, lockSHA string) ([]byte, Objects, error) {
	p, e := openLineagePackage(data)
	if e != nil {
		return nil, objects, e
	}
	if len(objects.SourceSHA256) != 64 || len(lockSHA) != 64 || objects.DeckID == "" {
		return nil, objects, fmt.Errorf("native lineage requires source, lock and deck identity")
	}
	l := &NativeLineage{Schema: NativeLineageSchema, DeckToken: lineageToken([]string{"deck", objects.DeckID}), SourceSemanticSHA256: objects.SourceSHA256, LockSHA256: lockSHA, Slides: map[string]string{}}
	l.UnstampedNativeSHA256 = digest(data)
	l.BuildToken = lineageToken([]string{NativeLineageSchema, l.DeckToken, objects.SourceSHA256, lockSHA, l.UnstampedNativeSHA256})
	if len(objects.Objects) > 10000 {
		return nil, objects, fmt.Errorf("native lineage supports at most 10000 shapes")
	}
	objects.Objects = append([]ObjectRecord(nil), objects.Objects...)
	byPart := map[string]map[string]string{}
	slideTokens := map[string]string{}
	seen := map[string]bool{}
	for i := range objects.Objects {
		r := &objects.Objects[i]
		if r.NativePart == "" || r.NativeID == "" || r.SlideID == "" {
			return nil, objects, fmt.Errorf("incomplete native baseline identity")
		}
		if byPart[r.NativePart] == nil {
			byPart[r.NativePart] = map[string]string{}
		}
		if byPart[r.NativePart][r.NativeID] != "" {
			return nil, objects, fmt.Errorf("duplicate native baseline shape ID in %s", r.NativePart)
		}
		st := lineageToken([]string{"slide", l.BuildToken, r.SlideID})
		if old := slideTokens[r.NativePart]; old != "" && old != st {
			return nil, objects, fmt.Errorf("multiple authored slides in one native part")
		}
		slideTokens[r.NativePart] = st
		l.Slides[st] = r.SlideID
		token := lineageToken([]string{"shape", l.BuildToken, r.SlideID, r.LogicalID, r.NativeID})
		if seen[token] {
			return nil, objects, fmt.Errorf("duplicate native shape token")
		}
		seen[token] = true
		r.ShapeToken = token
		byPart[r.NativePart][r.NativeID] = token
	}
	objects.Lineage = l
	objects.Reconciliation = "generation and native object tags recorded; edited-PPTX three-way proposals and source adoption remain pending"
	changed := map[string][]byte{}
	counter := 0
	var contentOverrides strings.Builder
	relAdditions := map[string]*strings.Builder{}
	relIDs := map[string]map[string]bool{}
	ct, e := p.read("[Content_Types].xml")
	if e != nil {
		return nil, objects, e
	}
	addTags := func(part string, tags map[string]string) (string, error) {
		counter++
		tagPart := fmt.Sprintf("ppt/tags/pptxgengo%d.xml", counter)
		if p.files[tagPart] != nil {
			return "", fmt.Errorf("native tag part collision")
		}
		keys := []string{}
		for k := range tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var t strings.Builder
		t.WriteString(`<p:tagLst xmlns:p="` + lineagePML + `">`)
		for _, k := range keys {
			if !validLineageToken(tags[k]) && !(k == "PPTXGENGO_SCHEMA" && tags[k] == "1") {
				return "", fmt.Errorf("invalid lineage tag value")
			}
			t.WriteString(`<p:tag name="` + k + `" val="` + tags[k] + `"/>`)
		}
		t.WriteString(`</p:tagLst>`)
		changed[tagPart] = []byte(t.String())
		rp := lineageRelPart(part)
		if relAdditions[rp] == nil {
			relAdditions[rp] = &strings.Builder{}
			relIDs[rp] = map[string]bool{}
			if p.files[rp] != nil {
				rels, err := p.relationships(part)
				if err != nil {
					return "", err
				}
				for id := range rels {
					relIDs[rp][id] = true
				}
			}
		}
		relID := fmt.Sprintf("pptxgengoTag%d", counter)
		if relIDs[rp][relID] {
			return "", fmt.Errorf("native tag relationship collision")
		}
		relIDs[rp][relID] = true
		target := path.Base(tagPart)
		if strings.HasPrefix(part, "ppt/slides/") {
			target = "../tags/" + target
		} else {
			target = "tags/" + target
		}
		relAdditions[rp].WriteString(`<Relationship xmlns="` + lineageOPC + `" Id="` + relID + `" Type="` + lineageTagType + `" Target="` + target + `"/>`)
		contentOverrides.WriteString(`<Override xmlns="` + lineageCT + `" PartName="/` + tagPart + `" ContentType="` + lineageTagContent + `"/>`)
		return relID, nil
	}
	presentation := "ppt/presentation.xml"
	pb, e := p.read(presentation)
	if e != nil {
		return nil, objects, e
	}
	ps, e := lineageSpans(pb)
	if e != nil {
		return nil, objects, e
	}
	relID, e := addTags(presentation, map[string]string{"PPTXGENGO_SCHEMA": "1", "PPTXGENGO_DECK": l.DeckToken, "PPTXGENGO_BUILD": l.BuildToken})
	if e != nil {
		return nil, objects, e
	}
	patch, e := lineageInsert(ps, lineageCustomerData(relID), map[string]bool{"kinsoku": true, "defaultTextStyle": true, "modifyVerifier": true, "extLst": true}, pb)
	if e != nil {
		return nil, objects, e
	}
	changed[presentation], e = lineageApply(pb, []lineagePatch{patch})
	if e != nil {
		return nil, objects, e
	}
	parts := []string{}
	for part := range byPart {
		parts = append(parts, part)
	}
	sort.Strings(parts)
	for _, part := range parts {
		b, err := p.read(part)
		if err != nil {
			return nil, objects, err
		}
		root, err := lineageSpans(b)
		if err != nil {
			return nil, objects, err
		}
		patches := []lineagePatch{}
		stamped := map[string]bool{}
		slideStamped := false
		var walk func(*lineageSpan) error
		walk = func(s *lineageSpan) error {
			if s.node.Name.Space == lineagePML && s.node.Name.Local == "cSld" {
				rid, err := addTags(part, map[string]string{"PPTXGENGO_SLIDE": slideTokens[part], "PPTXGENGO_BUILD": l.BuildToken})
				if err != nil {
					return err
				}
				patch, err := lineageInsert(s, lineageCustomerData(rid), map[string]bool{"controls": true, "extLst": true}, b)
				if err != nil {
					return err
				}
				patches = append(patches, patch)
				slideStamped = true
			}
			var id, nvpr *lineageSpan
			for _, c := range s.children {
				if c.node.Name.Space == lineagePML && c.node.Name.Local == "cNvPr" {
					id = c
				}
				if c.node.Name.Space == lineagePML && c.node.Name.Local == "nvPr" {
					nvpr = c
				}
			}
			if id != nil && nvpr != nil {
				nativeID := lineageAttr(id.node, "", "id")
				if token := byPart[part][nativeID]; token != "" {
					if stamped[nativeID] {
						return fmt.Errorf("duplicate native ID while stamping")
					}
					rid, err := addTags(part, map[string]string{"PPTXGENGO_SHAPE": token, "PPTXGENGO_BUILD": l.BuildToken})
					if err != nil {
						return err
					}
					patch, err := lineageInsert(nvpr, lineageCustomerData(rid), map[string]bool{"extLst": true}, b)
					if err != nil {
						return err
					}
					patches = append(patches, patch)
					stamped[nativeID] = true
				}
			}
			for _, c := range s.children {
				if err := walk(c); err != nil {
					return err
				}
			}
			return nil
		}
		if err = walk(root); err != nil {
			return nil, objects, err
		}
		if !slideStamped || len(stamped) != len(byPart[part]) {
			return nil, objects, fmt.Errorf("native baseline shape/slide tags incomplete")
		}
		changed[part], err = lineageApply(b, patches)
		if err != nil {
			return nil, objects, err
		}
	}
	for rp, additions := range relAdditions {
		var raw []byte
		if p.files[rp] != nil {
			raw, e = p.read(rp)
			if e != nil {
				return nil, objects, e
			}
		} else {
			raw = []byte(`<Relationships xmlns="` + lineageOPC + `"></Relationships>`)
		}
		changed[rp], e = lineageAppendRoot(raw, additions.String())
		if e != nil {
			return nil, objects, e
		}
	}
	ct, e = lineageAppendRoot(ct, contentOverrides.String())
	if e != nil {
		return nil, objects, e
	}
	changed["[Content_Types].xml"] = ct
	stamped, e := lineageRewrite(p, changed)
	if e != nil {
		return nil, objects, e
	}
	inspection, e := inspectNativeLineage(stamped, objects, false)
	if e != nil {
		return nil, objects, e
	}
	if len(inspection.Issues) != 0 {
		return nil, objects, fmt.Errorf("generated native lineage incomplete: %s", inspection.Issues[0].Kind)
	}
	parents := map[string]string{}
	for _, r := range inspection.Objects {
		parents[r.ShapeToken] = r.ParentToken
	}
	for i := range objects.Objects {
		objects.Objects[i].NativeParentToken = parents[objects.Objects[i].ShapeToken]
	}
	return stamped, objects, nil
}
func lineageRewrite(p *lineagePackage, changed map[string][]byte) ([]byte, error) {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	done := map[string]bool{}
	for _, f := range p.order {
		if raw, ok := changed[f.Name]; ok {
			h := f.FileHeader
			h.CRC32 = 0
			h.CompressedSize64 = 0
			h.UncompressedSize64 = 0
			fw, e := w.CreateHeader(&h)
			if e != nil {
				return nil, e
			}
			if _, e = fw.Write(raw); e != nil {
				return nil, e
			}
			done[f.Name] = true
		} else {
			if e := w.Copy(f); e != nil {
				return nil, e
			}
		}
	}
	names := []string{}
	for n := range changed {
		if !done[n] {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, n := range names {
		h := zip.FileHeader{Name: n, Method: zip.Deflate}
		h.SetModTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
		h.SetMode(0644)
		fw, e := w.CreateHeader(&h)
		if e != nil {
			return nil, e
		}
		if _, e = fw.Write(changed[n]); e != nil {
			return nil, e
		}
	}
	if e := w.Close(); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}

// InspectNativeLineage verifies the presentation-level baseline token, then
// follows live presentation relationships. Names, numeric IDs and ZIP part paths
// are diagnostic addresses only. Missing or duplicated tags remain manual issues;
// they never authorize deletion or a guessed source match. The caller must first
// authenticate the baseline sidecar against its immutable build receipt.
func InspectNativeLineage(data []byte, baseline Objects) (NativeLineageInspection, error) {
	return inspectNativeLineage(data, baseline, true)
}
func inspectNativeLineage(data []byte, baseline Objects, compareParents bool) (NativeLineageInspection, error) {
	out := NativeLineageInspection{Schema: NativeLineageSchema, Objects: []NativeLineageObject{}, Issues: []NativeLineageIssue{}}
	l := baseline.Lineage
	if l == nil || l.Schema != NativeLineageSchema || !validLineageToken(l.DeckToken) || !validLineageToken(l.BuildToken) || l.DeckToken != lineageToken([]string{"deck", baseline.DeckID}) || l.SourceSemanticSHA256 != baseline.SourceSHA256 {
		return out, fmt.Errorf("baseline lacks valid native lineage; rebuild before reconciliation")
	}
	if len(l.LockSHA256) != 64 || len(l.UnstampedNativeSHA256) != 64 || l.BuildToken != lineageToken([]string{NativeLineageSchema, l.DeckToken, l.SourceSemanticSHA256, l.LockSHA256, l.UnstampedNativeSHA256}) {
		return out, fmt.Errorf("baseline generation token does not match source/lock/native pins")
	}
	expected := map[string]ObjectRecord{}
	expectedSlides := map[string]string{}
	for token, id := range l.Slides {
		if !validLineageToken(token) || token != lineageToken([]string{"slide", l.BuildToken, id}) {
			return out, fmt.Errorf("invalid baseline slide token")
		}
		expectedSlides[id] = token
	}
	for _, r := range baseline.Objects {
		if !validLineageToken(r.ShapeToken) || r.ShapeToken != lineageToken([]string{"shape", l.BuildToken, r.SlideID, r.LogicalID, r.NativeID}) || expected[r.ShapeToken].ShapeToken != "" || expectedSlides[r.SlideID] == "" {
			return out, fmt.Errorf("invalid or duplicate baseline shape token")
		}
		expected[r.ShapeToken] = r
	}
	p, e := openLineagePackage(data)
	if e != nil {
		return out, e
	}
	rootRels, e := p.relationships("")
	if e != nil {
		return out, e
	}
	presentation := ""
	for _, r := range rootRels {
		if r.Type == lineageRML+"/officeDocument" {
			if presentation != "" || (r.Mode != "" && r.Mode != "Internal") {
				return out, fmt.Errorf("ambiguous/external native presentation")
			}
			presentation, e = lineageTarget("", r.Target)
			if e != nil {
				return out, e
			}
		}
	}
	if presentation == "" {
		return out, fmt.Errorf("missing native presentation relationship")
	}
	tree, e := p.tree(presentation)
	if e != nil {
		return out, e
	}
	if tree.Name != (xml.Name{Space: lineagePML, Local: "presentation"}) {
		return out, fmt.Errorf("invalid native presentation")
	}
	tags, e := p.tags(presentation, tree)
	if e != nil {
		return out, e
	}
	if tags["PPTXGENGO_SCHEMA"] != "1" || tags["PPTXGENGO_DECK"] != l.DeckToken || tags["PPTXGENGO_BUILD"] != l.BuildToken {
		return out, fmt.Errorf("edited PPTX deck/build lineage does not match baseline")
	}
	out.DeckToken = l.DeckToken
	out.BuildToken = l.BuildToken
	rels, e := p.relationships(presentation)
	if e != nil {
		return out, e
	}
	list := lineageChild(tree, lineagePML, "sldIdLst")
	if list == nil {
		return out, fmt.Errorf("missing native slide list")
	}
	countShapes := map[string]int{}
	countSlides := map[string]int{}
	seenParts := map[string]bool{}
	issue := func(kind, part, shape, slide, detail string) {
		out.Issues = append(out.Issues, NativeLineageIssue{kind, part, shape, slide, detail})
	}
	for _, entry := range list.Children {
		if entry.Name != (xml.Name{Space: lineagePML, Local: "sldId"}) {
			return out, fmt.Errorf("invalid native slide list entry")
		}
		r := rels[lineageAttr(entry, lineageRML, "id")]
		if r.Type != lineageRML+"/slide" || (r.Mode != "" && r.Mode != "Internal") {
			return out, fmt.Errorf("invalid/external slide relationship")
		}
		part, err := lineageTarget(presentation, r.Target)
		if err != nil {
			return out, err
		}
		if seenParts[part] {
			return out, fmt.Errorf("duplicate live native slide relationship")
		}
		seenParts[part] = true
		st, err := p.tree(part)
		if err != nil {
			return out, err
		}
		if st.Name != (xml.Name{Space: lineagePML, Local: "sld"}) {
			return out, fmt.Errorf("invalid native slide root")
		}
		common := lineageChild(st, lineagePML, "cSld")
		if common == nil {
			return out, fmt.Errorf("missing native common slide data")
		}
		t, err := p.tags(part, common)
		if err != nil {
			return out, err
		}
		slideToken := t["PPTXGENGO_SLIDE"]
		if slideToken == "" {
			issue("slide_untagged", part, "", "", "slide requires manual review")
		} else if !validLineageToken(slideToken) || l.Slides[slideToken] == "" || t["PPTXGENGO_BUILD"] != l.BuildToken {
			issue("slide_unmatched", part, "", slideToken, "slide tag does not belong to baseline")
		} else {
			countSlides[slideToken]++
		}
		spTree := lineageChild(common, lineagePML, "spTree")
		if spTree == nil {
			return out, fmt.Errorf("missing native shape tree")
		}
		var walk func(*xmlNode, string) error
		walk = func(n *xmlNode, parentToken string) error {
			isShape := n.Name.Space == lineagePML && (n.Name.Local == "sp" || n.Name.Local == "pic" || n.Name.Local == "graphicFrame" || n.Name.Local == "grpSp" || n.Name.Local == "cxnSp")
			if isShape {
				nv := lineageNV(n)
				if nv == nil {
					return fmt.Errorf("missing/ambiguous own native identity container")
				}
				id := lineageChild(nv, lineagePML, "cNvPr")
				props := lineageChild(nv, lineagePML, "nvPr")
				if id == nil || props == nil {
					return fmt.Errorf("missing own native identity")
				}
				tags, err := p.tags(part, props)
				if err != nil {
					return err
				}
				token := tags["PPTXGENGO_SHAPE"]
				obj := NativeLineageObject{ShapeToken: token, SlideToken: slideToken, NativePart: part, NativeID: lineageAttr(id, "", "id"), NativeName: lineageAttr(id, "", "name"), Kind: n.Name.Local, ParentToken: parentToken, shape: n}
				if len(out.Objects) >= 10000 {
					return fmt.Errorf("native lineage supports at most 10000 shapes")
				}
				out.Objects = append(out.Objects, obj)
				if token == "" {
					issue("shape_untagged", part, "", slideToken, "new or untagged shape requires manual review")
				} else if !validLineageToken(token) || expected[token].ShapeToken == "" || tags["PPTXGENGO_BUILD"] != l.BuildToken {
					issue("shape_unmatched", part, token, slideToken, "shape tag does not belong to baseline")
				} else {
					countShapes[token]++
					if compareParents && expected[token].NativeParentToken != parentToken {
						issue("shape_parent_changed", part, token, slideToken, "group ownership changed; requires manual review")
					}
					if expected[token].NativeKind != "" && expected[token].NativeKind != n.Name.Local {
						issue("shape_kind_changed", part, token, slideToken, "native object kind differs from baseline")
					}
					if expectedSlides[expected[token].SlideID] != slideToken {
						issue("shape_moved_between_slides", part, token, slideToken, "tagged shape is on a different authored slide")
					}
				}
				if n.Name.Local == "grpSp" {
					parentToken = token
				}
			}
			for _, c := range n.Children {
				if err := walk(c, parentToken); err != nil {
					return err
				}
			}
			return nil
		}
		if err = walk(spTree, ""); err != nil {
			return out, err
		}
	}
	for token := range expected {
		switch countShapes[token] {
		case 0:
			issue("shape_missing", "", token, "", "missing shape does not imply source deletion")
		case 1:
		default:
			issue("shape_duplicated", "", token, "", "multiple shapes carry one baseline token; no match may be guessed")
		}
	}
	for token := range l.Slides {
		switch countSlides[token] {
		case 0:
			issue("slide_missing", "", "", token, "missing slide does not imply source deletion")
		case 1:
		default:
			issue("slide_duplicated", "", "", token, "multiple slides carry one baseline token")
		}
	}
	sort.Slice(out.Issues, func(i, j int) bool {
		a, b := out.Issues[i], out.Issues[j]
		return a.Kind+"\x00"+a.NativePart+"\x00"+a.SlideToken+"\x00"+a.ShapeToken < b.Kind+"\x00"+b.NativePart+"\x00"+b.SlideToken+"\x00"+b.ShapeToken
	})
	return out, nil
}
func lineageNV(shape *xmlNode) *xmlNode {
	names := map[string]string{"sp": "nvSpPr", "pic": "nvPicPr", "graphicFrame": "nvGraphicFramePr", "grpSp": "nvGrpSpPr", "cxnSp": "nvCxnSpPr"}
	name := names[shape.Name.Local]
	var out *xmlNode
	for _, c := range shape.Children {
		if c.Name == (xml.Name{Space: lineagePML, Local: name}) {
			if out != nil {
				return nil
			}
			out = c
		}
	}
	return out
}

func (p *lineagePackage) contentType(part string) (string, error) {
	if p.contentTypes == nil {
		tree, e := p.tree("[Content_Types].xml")
		if e != nil {
			return "", e
		}
		if tree.Name != (xml.Name{Space: lineageCT, Local: "Types"}) {
			return "", fmt.Errorf("invalid native content types root")
		}
		overrides, defaults := map[string]string{}, map[string]string{}
		for _, c := range tree.Children {
			if c.Name.Space != lineageCT {
				return "", fmt.Errorf("foreign native content type entry")
			}
			mime := lineageAttr(c, "", "ContentType")
			if mime == "" || len(mime) > 256 {
				return "", fmt.Errorf("invalid native content type")
			}
			switch c.Name.Local {
			case "Override":
				name := lineageAttr(c, "", "PartName")
				if !strings.HasPrefix(name, "/") || path.Clean(name) != name || strings.ContainsAny(name, "\\\x00?#:") {
					return "", fmt.Errorf("invalid native content type part name")
				}
				name = strings.TrimPrefix(name, "/")
				if overrides[name] != "" {
					return "", fmt.Errorf("duplicate native content type override")
				}
				overrides[name] = mime
			case "Default":
				ext := strings.ToLower(lineageAttr(c, "", "Extension"))
				if ext == "" || strings.ContainsAny(ext, "./\\\x00?#:") || defaults[ext] != "" {
					return "", fmt.Errorf("invalid/duplicate native content type extension")
				}
				defaults[ext] = mime
			default:
				return "", fmt.Errorf("invalid native content type entry")
			}
		}
		p.contentTypes, p.contentDefaults = overrides, defaults
	}
	if mime := p.contentTypes[part]; mime != "" {
		return mime, nil
	}
	return p.contentDefaults[strings.TrimPrefix(strings.ToLower(path.Ext(part)), ".")], nil
}
