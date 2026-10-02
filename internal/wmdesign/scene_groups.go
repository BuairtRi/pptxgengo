package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
)

// nativeSceneObject is a complete direct spTree child. Balanced XML scanning
// keeps nested groups opaque, so their descendants never become siblings.
type nativeSceneObject struct {
	start, end int
	name, kind string
}
type nativeSceneInventory struct {
	objects []nativeSceneObject
	names   map[string]bool
	nextID  int
}

func sceneNativeInventory(data []byte) (nativeSceneInventory, error) {
	out := nativeSceneInventory{names: map[string]bool{}, nextID: 1}
	ids := map[int]bool{}
	d := xml.NewDecoder(bytes.NewReader(data))
	depth, treeDepth, objectDepth := 0, -1, -1
	var current nativeSceneObject
	treeCount := 0
	for {
		before := int(d.InputOffset())
		tok, e := d.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			return out, fmt.Errorf("scene.group_invalid_native_xml: %w", e)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "spTree" {
				treeCount++
				if treeCount != 1 {
					return out, fmt.Errorf("scene.group_multiple_native_shape_trees")
				}
				treeDepth = depth
			}
			if treeDepth >= 0 && depth == treeDepth+1 {
				switch t.Name.Local {
				case "sp", "pic", "graphicFrame", "grpSp", "cxnSp":
					at := bytes.LastIndexByte(data[before:int(d.InputOffset())], '<')
					if at < 0 {
						return out, fmt.Errorf("scene.group_invalid_native_object_start")
					}
					current = nativeSceneObject{start: before + at, kind: t.Name.Local}
					objectDepth = depth
				}
			}
			if t.Name.Local == "cNvPr" && treeDepth >= 0 {
				name, idText := "", ""
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "name":
						name = a.Value
					case "id":
						idText = a.Value
					}
				}
				id, e := strconv.Atoi(idText)
				if e != nil || id < 1 {
					return out, fmt.Errorf("scene.group_invalid_native_object_id: %s", idText)
				}
				if ids[id] {
					return out, fmt.Errorf("scene.group_duplicate_native_object_id: %d", id)
				}
				ids[id] = true
				if id >= out.nextID {
					out.nextID = id + 1
				}
				if name != "" {
					if out.names[name] {
						return out, fmt.Errorf("scene.group_duplicate_native_name: %s", name)
					}
					out.names[name] = true
				}
				if objectDepth >= 0 && depth == objectDepth+2 {
					if current.name != "" {
						return out, fmt.Errorf("scene.group_multiple_native_object_names")
					}
					current.name = name
				}
			}
		case xml.EndElement:
			if objectDepth >= 0 && depth == objectDepth {
				if current.name == "" {
					return out, fmt.Errorf("scene.group_missing_native_name: %s", current.kind)
				}
				current.end = int(d.InputOffset())
				out.objects = append(out.objects, current)
				objectDepth = -1
			}
			if t.Name.Local == "spTree" {
				treeDepth = -1
			}
			depth--
		}
	}
	if treeCount != 1 {
		return out, fmt.Errorf("scene.group_missing_native_shape_tree")
	}
	return out, nil
}

func sceneGroupHeader(g ComponentRecord, id int) ([]byte, error) {
	b := g.Rect
	if b.W < 0 || b.H < 0 || math.IsNaN(b.X+b.Y+b.W+b.H) || math.IsInf(b.X+b.Y+b.W+b.H, 0) {
		return nil, fmt.Errorf("scene.group_invalid_geometry: %s", g.ID)
	}
	emu := func(v float64) int64 { return int64(math.Round(v * 12700)) }
	x, y, w, h := emu(b.X), emu(b.Y), emu(b.W), emu(b.H)
	// A line-only group's zero axis is one EMU in both ext and chExt. This
	// maintains an identity mapping and preserves all actual child geometry.
	if w == 0 {
		w = 1
	}
	if h == 0 {
		h = 1
	}
	var name bytes.Buffer
	if e := xml.EscapeText(&name, []byte(g.ID)); e != nil {
		return nil, e
	}
	return []byte(fmt.Sprintf(`<p:grpSp><p:nvGrpSpPr><p:cNvPr id="%d" name="%s"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="%d" y="%d"/><a:ext cx="%d" cy="%d"/><a:chOff x="%d" y="%d"/><a:chExt cx="%d" cy="%d"/></a:xfrm></p:grpSpPr>`, id, name.String(), x, y, w, h, x, y, w, h)), nil
}

// sceneNativeGroups applies bottom-up scene ownership after the existing
// typography/component/card-row passes. Parts are an ownership set; the native
// object's actual paint order is authoritative and is never reordered.
func sceneNativeGroups(raw []byte, slides []SlideReport) ([]byte, error) {
	byPart := map[string][]ComponentRecord{}
	sceneParts := map[string][]string{}
	for i, slide := range slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		for _, scene := range slide.Scenes {
			byPart[part] = append(byPart[part], scene.Groups...)
			sceneParts[part] = append(sceneParts[part], scene.Parts...)
		}
	}
	hasGroups := false
	for _, gs := range byPart {
		hasGroups = hasGroups || len(gs) > 0
	}
	if !hasGroups {
		return raw, nil
	}
	z, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		return nil, e
	}
	seenParts := map[string]bool{}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, f := range z.File {
		rc, e := f.Open()
		if e != nil {
			return nil, e
		}
		data, e := io.ReadAll(rc)
		closeErr := rc.Close()
		if e != nil {
			return nil, e
		}
		if closeErr != nil {
			return nil, closeErr
		}
		groups := byPart[f.Name]
		if len(groups) > 0 {
			if seenParts[f.Name] {
				return nil, fmt.Errorf("scene.group_duplicate_slide_part: %s", f.Name)
			}
			seenParts[f.Name] = true
			inventory, e := sceneNativeInventory(data)
			if e != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, e)
			}
			declared := map[string]int{}
			owned := map[string]string{}
			nativeParts := map[string]bool{}
			for _, part := range sceneParts[f.Name] {
				if part == "" || nativeParts[part] {
					return nil, fmt.Errorf("scene.duplicate_scene_part: %s/%s", f.Name, part)
				}
				nativeParts[part] = true
				if !inventory.names[part] {
					return nil, fmt.Errorf("scene.missing_scene_part: %s/%s", f.Name, part)
				}
			}
			for i, g := range groups {
				if g.ID == "" {
					return nil, fmt.Errorf("scene.group_empty_id: %s", f.Name)
				}
				if _, exists := declared[g.ID]; exists || inventory.names[g.ID] {
					return nil, fmt.Errorf("scene.group_id_collision: %s", g.ID)
				}
				declared[g.ID] = i
			}
			for i, g := range groups {
				if len(g.Parts) == 0 {
					return nil, fmt.Errorf("scene.group_empty_ownership: %s", g.ID)
				}
				local := map[string]bool{}
				for _, part := range g.Parts {
					if part == "" || part == g.ID || local[part] {
						return nil, fmt.Errorf("scene.group_invalid_or_duplicate_child: %s/%s", g.ID, part)
					}
					local[part] = true
					if owner, exists := owned[part]; exists {
						return nil, fmt.Errorf("scene.group_multiple_owners: %s in %s and %s", part, owner, g.ID)
					}
					owned[part] = g.ID
					if childIndex, exists := declared[part]; exists {
						if childIndex >= i {
							return nil, fmt.Errorf("scene.group_not_bottom_up: %s references %s", g.ID, part)
						}
					} else if !inventory.names[part] {
						return nil, fmt.Errorf("scene.group_missing_native_child: %s/%s", g.ID, part)
					}
				}
			}
			nextID := inventory.nextID
			for _, g := range groups {
				inventory, e = sceneNativeInventory(data)
				if e != nil {
					return nil, e
				}
				wanted := map[string]bool{}
				for _, part := range g.Parts {
					wanted[part] = true
				}
				first, last, count := -1, -1, 0
				for _, object := range inventory.objects {
					if !wanted[object.name] {
						continue
					}
					if first < 0 {
						first = object.start
					} else if len(bytes.TrimSpace(data[last:object.start])) != 0 {
						return nil, fmt.Errorf("scene.group_noncontiguous_native_children: %s", g.ID)
					}
					last = object.end
					count++
				}
				if count != len(g.Parts) {
					return nil, fmt.Errorf("scene.group_missing_direct_native_children: %s expected %d found %d", g.ID, len(g.Parts), count)
				}
				header, e := sceneGroupHeader(g, nextID)
				if e != nil {
					return nil, e
				}
				nextID++
				data = bytes.Join([][]byte{data[:first], header, data[first:last], []byte("</p:grpSp>"), data[last:]}, nil)
			}
			if _, e = sceneNativeInventory(data); e != nil {
				return nil, e
			}
		}
		dst, e := w.CreateHeader(&f.FileHeader)
		if e != nil {
			return nil, e
		}
		if _, e = dst.Write(data); e != nil {
			return nil, e
		}
	}
	for part, groups := range byPart {
		if len(groups) > 0 && !seenParts[part] {
			return nil, fmt.Errorf("scene.group_missing_slide_part: %s", part)
		}
	}
	if e = w.Close(); e != nil {
		return nil, e
	}
	return out.Bytes(), nil
}
