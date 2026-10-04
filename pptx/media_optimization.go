package pptx

// Package-level optimization is deliberately separate from the compatible
// PptxGenJS writer: legacy Write callers retain their original package bytes.
import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"image"
	"image/jpeg"
	"io"
	"math"
	"net/url"
	"path"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// MediaOptimizationOptions is an explicit package build policy. Use
// DeliveryMediaOptions for production defaults. A zero policy preserves media
// payloads and uses STORE; it is not implicitly a delivery policy.
type MediaOptimizationOptions struct {
	Deduplicate       bool `json:"deduplicate"`
	ResizeJPEG        bool `json:"resize_jpeg"`
	Compression       bool `json:"compression"`
	PixelsPerInch     int  `json:"pixels_per_inch"`
	JPEGQuality       int  `json:"jpeg_quality"`
	MinSavingsPercent int  `json:"min_savings_percent"`
}

// DeliveryMediaOptions keeps vectors and lossless graphics byte-identical,
// sizes photographic JPEGs for their largest safe placement, and never upsizes.
func DeliveryMediaOptions() MediaOptimizationOptions {
	return MediaOptimizationOptions{Deduplicate: true, ResizeJPEG: true, Compression: true, PixelsPerInch: 220, JPEGQuality: 90, MinSavingsPercent: 10}
}

type MediaPartOptimization struct {
	SourcePart     string `json:"source_part"`
	OutputPart     string `json:"output_part"`
	SourceSHA256   string `json:"source_sha256"`
	OutputSHA256   string `json:"output_sha256"`
	SourceBytes    int    `json:"source_bytes"`
	OutputBytes    int    `json:"output_bytes"`
	SourceWidth    int    `json:"source_width,omitempty"`
	SourceHeight   int    `json:"source_height,omitempty"`
	OutputWidth    int    `json:"output_width,omitempty"`
	OutputHeight   int    `json:"output_height,omitempty"`
	RequiredWidth  int    `json:"required_width,omitempty"`
	RequiredHeight int    `json:"required_height,omitempty"`
	Resized        bool   `json:"resized"`
	Reused         bool   `json:"reused"`
	Reason         string `json:"reason"`
}

type MediaOptimizationReport struct {
	Schema           string                   `json:"schema"`
	Policy           MediaOptimizationOptions `json:"policy"`
	InputSHA256      string                   `json:"input_sha256"`
	OutputSHA256     string                   `json:"output_sha256"`
	InputBytes       int                      `json:"input_bytes"`
	OutputBytes      int                      `json:"output_bytes"`
	InputMediaParts  int                      `json:"input_media_parts"`
	OutputMediaParts int                      `json:"output_media_parts"`
	InputMediaBytes  int                      `json:"input_media_bytes"`
	OutputMediaBytes int                      `json:"output_media_bytes"`
	Parts            []MediaPartOptimization  `json:"parts"`
}

type mediaPackagePart struct {
	header zip.FileHeader
	data   []byte
}
type mediaUsage struct {
	width, height float64
	known, unsafe bool
}
type mediaRelation struct {
	id, target, kind string
	external         bool
}
type mediaXMLNode struct {
	name     xml.Name
	attrs    []xml.Attr
	children []*mediaXMLNode
}

const drawingNS = "http://schemas.openxmlformats.org/drawingml/2006/main"
const presentationNS = "http://schemas.openxmlformats.org/presentationml/2006/main"
const relationshipNS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const packageRelationshipNS = "http://schemas.openxmlformats.org/package/2006/relationships"
const packageContentTypesNS = "http://schemas.openxmlformats.org/package/2006/content-types"

// OptimizeMedia optimizes an already constructed OPC package in memory. It
// does not modify source files. References in every relationship part are
// redirected to retained media, including master/layout/SVG fallback relations.
// It leaves slide XML and object identity intact. Its receipt distinguishes
// original assets from derivatives; it does not promise deterministic upstream
// presentation metadata or IDs.
func OptimizeMedia(raw []byte, opts MediaOptimizationOptions) ([]byte, MediaOptimizationReport, error) {
	report := MediaOptimizationReport{Schema: "pptxgengo.media-optimization.v1", Policy: opts, InputBytes: len(raw), InputSHA256: mediaHash(raw)}
	if opts.PixelsPerInch == 0 {
		opts.PixelsPerInch = 220
	}
	if opts.JPEGQuality == 0 {
		opts.JPEGQuality = 90
	}
	if opts.PixelsPerInch < 72 || opts.PixelsPerInch > 1200 || opts.JPEGQuality < 1 || opts.JPEGQuality > 100 || opts.MinSavingsPercent < 0 || opts.MinSavingsPercent > 99 {
		return nil, report, fmt.Errorf("pptx: invalid media optimization policy")
	}
	report.Policy = opts
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, report, err
	}
	parts := make(map[string]*mediaPackagePart, len(z.File))
	order := make([]string, 0, len(z.File))
	var uncompressedBytes uint64
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "_xmlsignatures/") {
			return nil, report, fmt.Errorf("pptx: cannot optimize a signed package")
		}
		if _, exists := parts[f.Name]; exists {
			return nil, report, fmt.Errorf("pptx: duplicate package part %s", f.Name)
		}
		// Bound eager decompression before allocating. STORE parts can use
		// immutable views of the caller's package instead of copying it.
		if f.UncompressedSize64 > 256<<20 || uncompressedBytes > (2<<30)-f.UncompressedSize64 {
			return nil, report, fmt.Errorf("pptx: media optimization package exceeds resource limits")
		}
		uncompressedBytes += f.UncompressedSize64
		b, e := readMediaPackagePart(raw, f)
		if e != nil {
			return nil, report, fmt.Errorf("pptx: %s: %w", f.Name, e)
		}
		parts[f.Name] = &mediaPackagePart{header: f.FileHeader, data: b}
		order = append(order, f.Name)
	}
	rels := map[string][]mediaRelation{}
	for _, name := range order {
		if !strings.HasSuffix(name, ".rels") {
			continue
		}
		entries, e := readMediaRelations(parts[name].data)
		if e != nil {
			return nil, report, fmt.Errorf("pptx: %s: %w", name, e)
		}
		rels[name] = entries
	}
	// Infer requirements from picture transforms and crops, propagating group
	// scaling. Any unexplained use blocks resizing, but does not block dedup.
	usages := map[string]mediaUsage{}
	for _, name := range order {
		entries, ok := rels[name]
		if !ok {
			continue
		}
		owner := mediaRelationOwner(name)
		seen := map[string]mediaUsage{}
		if part := parts[owner]; part != nil && isMediaDrawingOwner(owner) {
			node, e := readMediaXML(part.data)
			if e != nil {
				return nil, report, fmt.Errorf("pptx: %s: %w", owner, e)
			}
			collectMediaUsage(node, 1, 1, seen)
			collectUnsafeMediaUsage(node, false, seen)
		}
		for _, rel := range entries {
			if rel.external {
				continue
			}
			target := resolveMediaTarget(owner, rel.target)
			if !strings.HasPrefix(target, "ppt/media/") {
				continue
			}
			u := seen[rel.id]
			if !strings.HasSuffix(rel.kind, "/image") || !u.known {
				u.unsafe = true
			}
			usages[target] = mergeMediaUsage(usages[target], u)
		}
	}
	// Equal originals share the largest requirement before derivatives are made.
	groups := map[string][]string{}
	groupOrder := []string{}
	contentTypes := map[string]string{}
	if ct := parts["[Content_Types].xml"]; ct != nil {
		contentTypes, err = readMediaContentTypes(ct.data)
		if err != nil {
			return nil, report, err
		}
	}
	for _, name := range order {
		if !strings.HasPrefix(name, "ppt/media/") || strings.HasSuffix(name, "/") {
			continue
		}
		key := mediaContentKey(name, contentTypes) + ":" + mediaHash(parts[name].data)
		if _, exists := groups[key]; !exists {
			groupOrder = append(groupOrder, key)
		}
		groups[key] = append(groups[key], name)
		report.InputMediaParts++
		report.InputMediaBytes += len(parts[name].data)
	}
	redirect := map[string]string{}
	retainedPayload := map[string]string{}
	removed := map[string]bool{}
	for _, key := range groupOrder {
		names := groups[key]
		usage := mediaUsage{}
		for _, name := range names {
			usage = mergeMediaUsage(usage, usages[name])
		}
		original := parts[names[0]].data
		sourceHash := mediaHash(original)
		resizePolicy := opts
		contentKey := mediaContentKey(names[0], contentTypes)
		compatibleJPEG := contentKey == ".jpg:image/jpeg"
		if !compatibleJPEG {
			resizePolicy.ResizeJPEG = false
		}
		output, receipt := resizeMediaJPEG(original, usage, resizePolicy)
		if !compatibleJPEG {
			receipt.Reason = "preserved_content_type"
		}
		outputHash := sourceHash
		if receipt.Resized {
			outputHash = mediaHash(output)
		}
		outputKey := mediaContentKey(names[0], contentTypes) + ":" + outputHash
		for _, name := range names {
			canonical := name
			if opts.Deduplicate {
				if existing, ok := retainedPayload[outputKey]; ok {
					canonical = existing
				} else {
					retainedPayload[outputKey] = name
				}
			}
			r := receipt
			r.SourcePart = name
			r.OutputPart = canonical
			r.SourceSHA256 = sourceHash
			r.OutputSHA256 = outputHash
			r.SourceBytes = len(original)
			r.OutputBytes = len(output)
			r.Reused = canonical != name
			report.Parts = append(report.Parts, r)
			if canonical != name {
				redirect[name] = canonical
				removed[name] = true
				parts[name].data = nil
			} else {
				parts[name].data = output
				report.OutputMediaParts++
				report.OutputMediaBytes += len(output)
			}
		}
	}
	for _, name := range order {
		if _, ok := rels[name]; !ok {
			continue
		}
		b, e := rewriteMediaRelations(parts[name].data, mediaRelationOwner(name), redirect)
		if e != nil {
			return nil, report, e
		}
		parts[name].data = b
	}
	if contentTypes := parts["[Content_Types].xml"]; contentTypes != nil {
		b, e := removeMediaOverrides(contentTypes.data, removed)
		if e != nil {
			return nil, report, e
		}
		contentTypes.data = b
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, name := range order {
		if removed[name] {
			continue
		}
		part := parts[name]
		h := part.header
		h.Method = zip.Store
		if opts.Compression {
			h.Method = zip.Deflate
		}
		dst, e := w.CreateHeader(&h)
		if e != nil {
			return nil, report, e
		}
		if _, e = dst.Write(part.data); e != nil {
			return nil, report, e
		}
	}
	if err := w.Close(); err != nil {
		return nil, report, err
	}
	report.OutputBytes = out.Len()
	report.OutputSHA256 = mediaHash(out.Bytes())
	return out.Bytes(), report, nil
}

func readMediaPackagePart(raw []byte, f *zip.File) ([]byte, error) {
	if f.Method == zip.Store {
		if f.CompressedSize64 != f.UncompressedSize64 {
			return nil, fmt.Errorf("invalid stored part sizes")
		}
		offset, err := f.DataOffset()
		if err != nil {
			return nil, err
		}
		if offset < 0 || offset > int64(len(raw)) || f.UncompressedSize64 > uint64(int64(len(raw))-offset) {
			return nil, io.ErrUnexpectedEOF
		}
		data := raw[offset : offset+int64(f.UncompressedSize64)]
		if crc32.ChecksumIEEE(data) != f.CRC32 {
			return nil, zip.ErrChecksum
		}
		return data, nil
	}
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	// The extra byte ensures the checksum reader reaches EOF and rejects
	// a stream longer or shorter than its declared uncompressed size.
	data := make([]byte, int(f.UncompressedSize64)+1)
	n, readErr := io.ReadFull(r, data)
	closeErr := r.Close()
	if n != int(f.UncompressedSize64) || (readErr != io.ErrUnexpectedEOF && readErr != io.EOF) {
		if readErr == nil || readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("invalid decompressed part size")
		}
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return data[:n], nil
}

func mediaHash(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func mergeMediaUsage(a, b mediaUsage) mediaUsage {
	return mediaUsage{width: math.Max(a.width, b.width), height: math.Max(a.height, b.height), known: a.known || b.known, unsafe: a.unsafe || b.unsafe}
}
func mediaAttr(n *mediaXMLNode, name string) string {
	if n != nil {
		for _, a := range n.attrs {
			if a.Name.Space == "" && a.Name.Local == name {
				return a.Value
			}
		}
	}
	return ""
}
func mediaChild(n *mediaXMLNode, space, local string) *mediaXMLNode {
	if n != nil {
		for _, c := range n.children {
			if c.name.Space == space && c.name.Local == local {
				return c
			}
		}
	}
	return nil
}
func mediaNumber(n *mediaXMLNode, name string) float64 {
	value, _ := strconv.ParseFloat(mediaAttr(n, name), 64)
	return value
}
func readMediaXML(b []byte) (*mediaXMLNode, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	root := &mediaXMLNode{}
	stack := []*mediaXMLNode{root}
	for {
		t, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch v := t.(type) {
		case xml.StartElement:
			n := &mediaXMLNode{name: v.Name, attrs: v.Attr}
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, n)
			stack = append(stack, n)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		}
	}
	return root, nil
}
func collectMediaUsage(n *mediaXMLNode, sx, sy float64, seen map[string]mediaUsage) {
	if n.name.Space == presentationNS && n.name.Local == "grpSp" {
		x := mediaChild(mediaChild(n, presentationNS, "grpSpPr"), drawingNS, "xfrm")
		ext, child := mediaChild(x, drawingNS, "ext"), mediaChild(x, drawingNS, "chExt")
		cx, cy := mediaNumber(child, "cx"), mediaNumber(child, "cy")
		if cx <= 0 || cy <= 0 {
			sx, sy = 0, 0
		} else {
			localX, localY := mediaNumber(ext, "cx")/cx, mediaNumber(ext, "cy")/cy
			if mediaNumber(x, "rot") != 0 {
				// Matrix norm bound, applied before multiplying axes: rotation
				// can exchange the large axes of parent and child transforms.
				bound := math.Max(sx, sy) * math.Max(localX, localY)
				sx, sy = bound, bound
			} else {
				sx *= localX
				sy *= localY
			}
		}
	}
	if n.name.Space == presentationNS && n.name.Local == "pic" {
		fill := mediaChild(n, presentationNS, "blipFill")
		blip := mediaChild(fill, drawingNS, "blip")
		id := ""
		if blip != nil {
			for _, a := range blip.attrs {
				if a.Name.Space == relationshipNS && a.Name.Local == "embed" {
					id = a.Value
				}
			}
		}
		xfrm := mediaChild(mediaChild(n, presentationNS, "spPr"), drawingNS, "xfrm")
		ext := mediaChild(xfrm, drawingNS, "ext")
		crop := mediaChild(fill, drawingNS, "srcRect")
		cw := 1 - (mediaNumber(crop, "l")+mediaNumber(crop, "r"))/100000
		ch := 1 - (mediaNumber(crop, "t")+mediaNumber(crop, "b"))/100000
		picSX, picSY := sx, sy
		if mediaNumber(xfrm, "rot") != 0 {
			picSX, picSY = math.Max(sx, sy), math.Max(sx, sy)
		}
		width, height := mediaNumber(ext, "cx")*picSX/914400, mediaNumber(ext, "cy")*picSY/914400
		u := mediaUsage{known: true}
		// Vector fallback JPEGs remain exact, including SVG extension relations.
		u.unsafe = width <= 0 || height <= 0 || cw <= 0 || ch <= 0 || cw > 1 || ch > 1 || mediaChild(blip, drawingNS, "extLst") != nil || mediaChild(fill, drawingNS, "tile") != nil
		for _, edge := range []string{"l", "r", "t", "b"} {
			if value := mediaNumber(crop, edge); value < 0 || value > 100000 {
				u.unsafe = true
			}
		}
		if fillRect := mediaChild(mediaChild(fill, drawingNS, "stretch"), drawingNS, "fillRect"); fillRect != nil && len(fillRect.attrs) > 0 {
			u.unsafe = true
		}
		if !u.unsafe {
			u.width, u.height = width/cw, height/ch
		}
		if id != "" {
			seen[id] = mergeMediaUsage(seen[id], u)
		}
	}
	for _, child := range n.children {
		collectMediaUsage(child, sx, sy, seen)
	}
}

func collectUnsafeMediaUsage(n *mediaXMLNode, inPicture bool, seen map[string]mediaUsage) {
	if n.name.Space == presentationNS && n.name.Local == "pic" {
		inPicture = true
	}
	if !inPicture {
		for _, a := range n.attrs {
			if a.Name.Space == relationshipNS && (a.Name.Local == "embed" || a.Name.Local == "id") {
				u := seen[a.Value]
				u.unsafe = true
				seen[a.Value] = u
			}
		}
	}
	for _, child := range n.children {
		collectUnsafeMediaUsage(child, inPicture, seen)
	}
}

func readMediaContentTypes(b []byte) (map[string]string, error) {
	result := map[string]string{}
	d := xml.NewDecoder(bytes.NewReader(b))
	for {
		t, e := d.Token()
		if e == io.EOF {
			return result, nil
		}
		if e != nil {
			return nil, e
		}
		start, ok := t.(xml.StartElement)
		if !ok || (start.Name.Space != packageContentTypesNS && start.Name.Space != "") || (start.Name.Local != "Default" && start.Name.Local != "Override") {
			continue
		}
		key, value := "", ""
		for _, a := range start.Attr {
			if a.Name.Space != "" {
				continue
			}
			switch a.Name.Local {
			case "Extension":
				key = "ext:" + strings.ToLower(a.Value)
			case "PartName":
				key = strings.TrimPrefix(a.Value, "/")
			case "ContentType":
				value = a.Value
			}
		}
		if key != "" {
			result[key] = value
		}
	}
}
func mediaContentKey(name string, contentTypes map[string]string) string {
	ext := strings.ToLower(path.Ext(name))
	contentType := contentTypes[name]
	if contentType == "" {
		contentType = contentTypes["ext:"+strings.TrimPrefix(ext, ".")]
	}
	// The compatible PptxGenJS writer declares .jpg as image/jpg; treat it
	// as the JPEG alias without changing the source package declaration.
	if (contentType == "image/jpeg" || contentType == "image/jpg") && (ext == ".jpeg" || ext == ".jpg") {
		ext = ".jpg"
		contentType = "image/jpeg"
	}
	return ext + ":" + contentType
}

func isMediaDrawingOwner(name string) bool {
	return strings.HasSuffix(name, ".xml") && (strings.HasPrefix(name, "ppt/slides/") || strings.HasPrefix(name, "ppt/slideLayouts/") || strings.HasPrefix(name, "ppt/slideMasters/"))
}
func mediaRelationOwner(name string) string {
	if name == "_rels/.rels" {
		return ""
	}
	return path.Join(path.Dir(path.Dir(name)), strings.TrimSuffix(path.Base(name), ".rels"))
}
func resolveMediaTarget(owner, target string) string {
	if i := strings.IndexAny(target, "?#"); i >= 0 {
		target = target[:i]
	}
	if decoded, e := url.PathUnescape(target); e == nil {
		target = decoded
	}
	if strings.HasPrefix(target, "/") {
		return path.Clean(strings.TrimPrefix(target, "/"))
	}
	return path.Clean(path.Join(path.Dir(owner), target))
}
func readMediaRelations(b []byte) ([]mediaRelation, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	var result []mediaRelation
	for {
		t, e := d.Token()
		if e == io.EOF {
			return result, nil
		}
		if e != nil {
			return nil, e
		}
		if start, ok := t.(xml.StartElement); ok && isMediaRelationshipElement(start.Name) {
			r := mediaRelation{}
			for _, a := range start.Attr {
				if a.Name.Space != "" {
					continue
				}
				switch a.Name.Local {
				case "Id":
					r.id = a.Value
				case "Target":
					r.target = a.Value
				case "Type":
					r.kind = a.Value
				case "TargetMode":
					r.external = a.Value == "External"
				}
			}
			result = append(result, r)
		}
	}
}

// Replace attribute values in their original token spans instead of rewriting
// XML namespaces or relationship IDs. Only targets of removed media change.
func mediaTargetAttributeSpan(token []byte) (int, int, bool) {
	// Token has already passed XML decoding. Walk complete attributes so an
	// apparent Target inside an extension value cannot be selected.
	i := 1
	space := func(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' }
	for i < len(token) && !space(token[i]) && token[i] != '/' && token[i] != '>' {
		i++
	}
	for i < len(token) {
		for i < len(token) && space(token[i]) {
			i++
		}
		if i == len(token) || token[i] == '/' || token[i] == '>' {
			break
		}
		nameStart := i
		for i < len(token) && !space(token[i]) && token[i] != '=' {
			i++
		}
		name := string(token[nameStart:i])
		for i < len(token) && space(token[i]) {
			i++
		}
		if i == len(token) || token[i] != '=' {
			break
		}
		i++
		for i < len(token) && space(token[i]) {
			i++
		}
		if i == len(token) || (token[i] != '\'' && token[i] != '"') {
			break
		}
		valueStart, quote := i, token[i]
		i++
		for i < len(token) && token[i] != quote {
			i++
		}
		if i == len(token) {
			break
		}
		i++
		if name == "Target" {
			return valueStart, i, true
		}
	}
	return 0, 0, false
}

func isMediaRelationshipElement(name xml.Name) bool {
	return name.Local == "Relationship" && (name.Space == packageRelationshipNS || name.Space == "")
}

func rewriteMediaRelations(b []byte, owner string, redirect map[string]string) ([]byte, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	var out bytes.Buffer
	last := int64(0)
	for {
		before := d.InputOffset()
		t, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		start, ok := t.(xml.StartElement)
		if !ok || !isMediaRelationshipElement(start.Name) {
			continue
		}
		target, external := "", false
		for _, a := range start.Attr {
			if a.Name.Space != "" {
				continue
			}
			if a.Name.Local == "Target" {
				target = a.Value
			}
			if a.Name.Local == "TargetMode" && a.Value == "External" {
				external = true
			}
		}
		if external {
			continue
		}
		canonical, ok := redirect[resolveMediaTarget(owner, target)]
		if !ok {
			continue
		}
		relative := relativeMediaTarget(path.Dir(owner), canonical)
		if strings.HasPrefix(target, "/") {
			relative = "/" + canonical
		}
		relative = (&url.URL{Path: relative}).EscapedPath()
		if i := strings.IndexAny(target, "?#"); i >= 0 {
			relative += target[i:]
		}
		var escaped bytes.Buffer
		_ = xml.EscapeText(&escaped, []byte(relative))
		after := d.InputOffset()
		token := b[before:after]
		valueStart, valueEnd, found := mediaTargetAttributeSpan(token)
		if !found {
			return nil, fmt.Errorf("pptx: missing media relationship target")
		}
		out.Write(b[last:before])
		out.Write(token[:valueStart])
		out.WriteString(`"`)
		out.Write(escaped.Bytes())
		out.WriteString(`"`)
		out.Write(token[valueEnd:])
		last = after
	}
	out.Write(b[last:])
	return out.Bytes(), nil
}
func relativeMediaTarget(dir, target string) string {
	a, b := strings.Split(path.Clean(dir), "/"), strings.Split(target, "/")
	if dir == "." {
		a = nil
	}
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	result := make([]string, 0, len(a)+len(b))
	for j := i; j < len(a); j++ {
		result = append(result, "..")
	}
	result = append(result, b[i:]...)
	return strings.Join(result, "/")
}
func removeMediaOverrides(b []byte, removed map[string]bool) ([]byte, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	var out bytes.Buffer
	last := int64(0)
	for {
		before := d.InputOffset()
		t, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		start, ok := t.(xml.StartElement)
		if !ok || start.Name.Local != "Override" || (start.Name.Space != packageContentTypesNS && start.Name.Space != "") {
			continue
		}
		remove := false
		for _, a := range start.Attr {
			if a.Name.Space == "" && a.Name.Local == "PartName" && removed[strings.TrimPrefix(a.Value, "/")] {
				remove = true
			}
		}
		if !remove {
			continue
		}
		if e = d.Skip(); e != nil {
			return nil, e
		}
		out.Write(b[last:before])
		last = d.InputOffset()
	}
	out.Write(b[last:])
	return out.Bytes(), nil
}

func resizeMediaJPEG(original []byte, usage mediaUsage, opts MediaOptimizationOptions) ([]byte, MediaPartOptimization) {
	r := MediaPartOptimization{Reason: "preserved_format"}
	if len(original) < 3 || original[0] != 0xff || original[1] != 0xd8 {
		return original, r
	}
	config, e := jpeg.DecodeConfig(bytes.NewReader(original))
	if e != nil {
		r.Reason = "unreadable_jpeg"
		return original, r
	}
	r.SourceWidth, r.SourceHeight = config.Width, config.Height
	r.OutputWidth, r.OutputHeight = config.Width, config.Height
	if !opts.ResizeJPEG {
		r.Reason = "resize_disabled"
		return original, r
	}
	if !usage.known || usage.unsafe {
		r.Reason = "unknown_or_unsafe_placement"
		return original, r
	}
	r.RequiredWidth = int(math.Ceil(usage.width * float64(opts.PixelsPerInch)))
	r.RequiredHeight = int(math.Ceil(usage.height * float64(opts.PixelsPerInch)))
	if r.RequiredWidth <= 0 || r.RequiredHeight <= 0 {
		r.Reason = "invalid_placement"
		return original, r
	}
	scale := math.Max(float64(r.RequiredWidth)/float64(config.Width), float64(r.RequiredHeight)/float64(config.Height))
	if scale >= 0.95 {
		r.Reason = "resolution_sufficient"
		return original, r
	}
	if int64(config.Width)*int64(config.Height) > 64_000_000 {
		r.Reason = "decode_pixel_limit"
		return original, r
	}
	if jpegHasSensitiveMetadata(original) {
		r.Reason = "color_profile_or_orientation"
		return original, r
	}
	im, e := jpeg.Decode(bytes.NewReader(original))
	if e != nil {
		r.Reason = "unreadable_jpeg"
		return original, r
	}
	w, h := int(math.Ceil(float64(config.Width)*scale)), int(math.Ceil(float64(config.Height)*scale))
	if w < 1 || h < 1 {
		r.Reason = "invalid_placement"
		return original, r
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	scaleMediaJPEG(dst, im)
	var out bytes.Buffer
	if e = jpeg.Encode(&out, dst, &jpeg.Options{Quality: opts.JPEGQuality}); e != nil {
		r.Reason = "encode_failed"
		return original, r
	}
	if out.Len() >= len(original) || float64(out.Len()) > float64(len(original))*float64(100-opts.MinSavingsPercent)/100 {
		r.Reason = "insufficient_byte_savings"
		return original, r
	}
	r.Resized = true
	r.OutputWidth, r.OutputHeight = w, h
	r.Reason = "jpeg_derivative"
	return out.Bytes(), r
}

func scaleMediaJPEG(dst *image.RGBA, im image.Image) {
	// Scale's separable implementation allocates 32 bytes per destination
	// column per source row. For large photographs, use the same antialiased
	// cubic kernel through Transform, which avoids that intermediate buffer.
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	if int64(w)*int64(im.Bounds().Dy())*32 > 16<<20 {
		draw.CatmullRom.Transform(dst, f64.Aff3{float64(w) / float64(im.Bounds().Dx()), 0, 0, 0, float64(h) / float64(im.Bounds().Dy()), 0}, im, im.Bounds(), draw.Src, nil)
	} else {
		draw.CatmullRom.Scale(dst, dst.Bounds(), im, im.Bounds(), draw.Src, nil)
	}
}

// The standard JPEG encoder drops metadata. Preserve images whose color profile
// or EXIF orientation could change appearance rather than silently stripping it.
func jpegHasSensitiveMetadata(b []byte) bool {
	for i := 2; i+4 <= len(b); {
		if b[i] != 0xff {
			return true
		}
		for i < len(b) && b[i] == 0xff {
			i++
		}
		if i >= len(b) {
			return true
		}
		marker := b[i]
		i++
		if marker == 0xda || marker == 0xd9 {
			return false
		}
		if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			continue
		}
		if i+2 > len(b) {
			return true
		}
		length := int(binary.BigEndian.Uint16(b[i : i+2]))
		if length < 2 || i+length > len(b) {
			return true
		}
		payload := b[i+2 : i+length]
		i += length
		if (marker == 0xe2 && bytes.HasPrefix(payload, []byte("ICC_PROFILE\x00"))) || (marker == 0xee && bytes.HasPrefix(payload, []byte("Adobe"))) {
			return true
		}
		// Four-component JPEGs are CMYK/YCCK and would change color model
		// through the standard encoder, even without a profile marker.
		if ((marker >= 0xc0 && marker <= 0xc3) || (marker >= 0xc5 && marker <= 0xc7) || (marker >= 0xc9 && marker <= 0xcb) || (marker >= 0xcd && marker <= 0xcf)) && len(payload) >= 6 && payload[5] == 4 {
			return true
		}
		if marker == 0xe1 && bytes.HasPrefix(payload, []byte("Exif\x00\x00")) {
			tiff := payload[6:]
			if len(tiff) < 8 {
				return true
			}
			var order binary.ByteOrder
			switch string(tiff[:2]) {
			case "II":
				order = binary.LittleEndian
			case "MM":
				order = binary.BigEndian
			default:
				return true
			}
			offset := uint64(order.Uint32(tiff[4:8]))
			if offset+2 > uint64(len(tiff)) {
				return true
			}
			count := uint64(order.Uint16(tiff[offset : offset+2]))
			for j := uint64(0); j < count; j++ {
				pos := offset + 2 + j*12
				if pos+12 > uint64(len(tiff)) {
					return true
				}
				entry := tiff[pos : pos+12]
				if order.Uint16(entry[:2]) == 0x0112 {
					if order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) != 1 || order.Uint16(entry[8:10]) != 1 {
						return true
					}
				}
			}
		}
	}
	return false
}
