package wmdesign

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// sceneTableTypography is v2-scoped. Native tables remain tables: this pass
// replaces cell text formatting and planned row heights, never their geometry,
// merges, fills, borders, workbook parts, or relationships.
func sceneTableTypography(raw []byte, slides []SlideReport) ([]byte, error) {
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, err
	}
	catalog := map[string]map[string]SceneTableRecord{}
	for i, slide := range slides {
		if len(slide.Tables) == 0 {
			continue
		}
		m := map[string]SceneTableRecord{}
		for _, table := range slide.Tables {
			if _, exists := m[table.ID]; exists {
				return nil, fmt.Errorf("scene.table_duplicate_record: %s", table.ID)
			}
			m[table.ID] = table
		}
		catalog[fmt.Sprintf("ppt/slides/slide%d.xml", i+1)] = m
	}
	frameRE := regexp.MustCompile(`(?s)<p:graphicFrame>.*?</p:graphicFrame>`)
	nameRE := regexp.MustCompile(`<p:cNvPr\b[^>]*\bname="([^"]*)"`)
	rowRE := regexp.MustCompile(`(?s)<a:tr\b[^>]*>.*?</a:tr>`)
	cellRE := regexp.MustCompile(`(?s)<a:tc\b[^>]*>.*?</a:tc>`)
	bodyRE := regexp.MustCompile(`(?s)<a:txBody>.*?</a:txBody>`)
	heightRE := regexp.MustCompile(`\bh="[^"]*"`)
	mergeRE := regexp.MustCompile(`\bhMerge="1"`)
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, file := range z.File {
		rd, err := file.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rd)
		rd.Close()
		if err != nil {
			return nil, err
		}
		tables, has := catalog[file.Name]
		if has {
			seen := map[string]bool{}
			var rewriteErr error
			data = frameRE.ReplaceAllFunc(data, func(frame []byte) []byte {
				if rewriteErr != nil {
					return frame
				}
				match := nameRE.FindSubmatch(frame)
				if match == nil {
					return frame
				}
				var named struct {
					Name string `xml:"name,attr"`
				}
				if err := xml.Unmarshal([]byte(`<n name="`+string(match[1])+`"/>`), &named); err != nil {
					rewriteErr = err
					return frame
				}
				table, exists := tables[named.Name]
				if !exists {
					return frame
				}
				if seen[table.ID] {
					rewriteErr = fmt.Errorf("scene.table_duplicate_native_object: %s", table.ID)
					return frame
				}
				seen[table.ID] = true
				if !bytes.Contains(frame, []byte("<a:tbl>")) {
					rewriteErr = fmt.Errorf("scene.table_native_object_type: %s", table.ID)
					return frame
				}
				cells := map[[2]int]TextRecord{}
				for _, cell := range table.Cells {
					key := [2]int{cell.Row, cell.Column}
					if _, exists := cells[key]; exists {
						rewriteErr = fmt.Errorf("scene.table_duplicate_cell_record: %s/%d/%d", table.ID, cell.Row, cell.Column)
						return frame
					}
					cells[key] = cell.Text
				}
				used := map[[2]int]bool{}
				rowIndex := 0
				updated := rowRE.ReplaceAllFunc(frame, func(row []byte) []byte {
					if rewriteErr != nil {
						return row
					}
					originColumn := 0
					var previous TextRecord
					rowHeight := 0.
					updatedRow := cellRE.ReplaceAllFunc(row, func(cell []byte) []byte {
						if rewriteErr != nil {
							return cell
						}
						startEnd := bytes.IndexByte(cell, '>')
						if startEnd < 0 {
							rewriteErr = fmt.Errorf("scene.table_invalid_native_cell")
							return cell
						}
						dummy := mergeRE.Match(cell[:startEnd+1])
						var tr TextRecord
						if dummy {
							if previous.ID == "" {
								rewriteErr = fmt.Errorf("scene.table_orphan_merge_cell: %s/%d", table.ID, rowIndex)
								return cell
							}
							tr = previous
							tr.Rich = nil
							tr.Layout.Original = ""
							tr.Layout.Displayed = ""
							tr.Layout.Lines = nil
						} else {
							key := [2]int{rowIndex, originColumn}
							var ok bool
							tr, ok = cells[key]
							if !ok {
								rewriteErr = fmt.Errorf("scene.table_missing_cell_record: %s/%d/%d", table.ID, rowIndex, originColumn)
								return cell
							}
							used[key] = true
							originColumn++
							previous = tr
							rowHeight = math.Max(rowHeight, tr.Rect.H)
						}
						if dummy && len(bodyRE.FindAll(cell, -1)) == 0 {
							body, err := sceneTableCellTextXML(tr)
							if err != nil {
								rewriteErr = err
								return cell
							}
							return bytes.Replace(cell, []byte("<a:tcPr"), append(body, []byte("<a:tcPr")...), 1)
						}
						if len(bodyRE.FindAll(cell, -1)) != 1 {
							rewriteErr = fmt.Errorf("scene.table_cell_textbody_count: %s", tr.ID)
							return cell
						}
						body, err := sceneTableCellTextXML(tr)
						if err != nil {
							rewriteErr = err
							return cell
						}
						return bodyRE.ReplaceAllLiteral(cell, body)
					})
					if rewriteErr != nil {
						return row
					}
					if rowHeight <= 0 {
						rewriteErr = fmt.Errorf("scene.table_missing_planned_row_height: %s/%d", table.ID, rowIndex)
						return row
					}
					openingEnd := bytes.IndexByte(updatedRow, '>')
					if openingEnd < 0 || !heightRE.Match(updatedRow[:openingEnd]) {
						rewriteErr = fmt.Errorf("scene.table_native_row_height_missing")
						return row
					}
					opening := heightRE.ReplaceAllString(string(updatedRow[:openingEnd]), `h="`+strconv.Itoa(int(math.Round(rowHeight*12700)))+`"`)
					updatedRow = append([]byte(opening), updatedRow[openingEnd:]...)
					rowIndex++
					return updatedRow
				})
				if rewriteErr != nil {
					return frame
				}
				if rowIndex != table.Rows {
					rewriteErr = fmt.Errorf("scene.table_native_row_count: %s got%d want%d", table.ID, rowIndex, table.Rows)
					return frame
				}
				if len(used) != len(cells) {
					rewriteErr = fmt.Errorf("scene.table_unconsumed_cell_records: %s", table.ID)
					return frame
				}
				return updated
			})
			if rewriteErr != nil {
				return nil, rewriteErr
			}
			if len(seen) != len(tables) {
				return nil, fmt.Errorf("scene.table_native_object_missing: %s", file.Name)
			}
		}
		header := file.FileHeader
		dst, err := w.CreateHeader(&header)
		if err != nil {
			return nil, err
		}
		if _, err = dst.Write(data); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func sceneTableXMLEscape(s string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
func sceneTableXMLBool(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
func sceneTableDefaultRun(tr TextRecord) string {
	s, id := tr.Layout.Style, tr.Layout.Font
	return ` lang="en-US" sz="` + strconv.Itoa(int(math.Round(s.Size*100))) + `" spc="` + strconv.Itoa(int(math.Round(s.TrackingPt*100))) + `" kern="0" b="` + sceneTableXMLBool(id.Bold) + `" i="` + sceneTableXMLBool(id.NativeItalic) + `" dirty="0">` + `<a:solidFill><a:srgbClr val="` + sceneTableXMLEscape(tr.Color) + `"/></a:solidFill><a:latin typeface="` + sceneTableXMLEscape(id.Typeface) + `"/><a:ea typeface="` + sceneTableXMLEscape(id.Typeface) + `"/><a:cs typeface="` + sceneTableXMLEscape(id.Typeface) + `"/>`
}
func sceneTableParagraphProperties(tr TextRecord, gap float64, bullet bool) []byte {
	align := map[string]string{"left": "l", "center": "ctr", "right": "r"}[tr.Align]
	if align == "" {
		align = "l"
	}
	margin := ` marL="0" indent="0"`
	marker := `<a:buNone/>`
	if bullet {
		margin = ` marL="152400" indent="-152400"`
		marker = `<a:buSzPts val="300"/><a:buFont typeface="` + sceneTableXMLEscape(tr.Layout.Font.Typeface) + `"/><a:buChar char="■"/>`
	}
	return []byte(`<a:pPr algn="` + align + `"` + margin + `><a:lnSpc><a:spcPts val="` + strconv.Itoa(int(math.Round(tr.Layout.Style.Leading*100))) + `"/></a:lnSpc><a:spcBef><a:spcPts val="0"/></a:spcBef><a:spcAft><a:spcPts val="` + strconv.Itoa(int(math.Round(gap*100))) + `"/></a:spcAft>` + marker + `<a:defRPr` + sceneTableDefaultRun(tr) + `</a:defRPr></a:pPr>`)
}
func sceneTableCellTextXML(tr TextRecord) ([]byte, error) {
	if tr.Layout.Style.Size <= 0 || tr.Layout.Style.Leading <= 0 || tr.Layout.Font.Typeface == "" || tr.Color == "" {
		return nil, fmt.Errorf("scene.table_invalid_cell_typography: %s", tr.ID)
	}
	var text strings.Builder
	if tr.Rich != nil {
		for _, para := range tr.Rich.Paragraphs {
			one := tr
			one.Rich = &RichTextLayout{Contract: tr.Rich.Contract, Paragraphs: []RichParagraphLayout{para}, NativeQualified: false}
			if para.ParagraphGapAfter < 0 {
				return nil, fmt.Errorf("scene.table_negative_paragraph_gap")
			}
			// Each paragraph may have its own token (e.g. body plus small subtitle).
			// Use the first run as the paragraph default and the greatest run leading
			// for mixed spans, matching the shaped paragraph's allocation.
			if len(para.Runs) > 0 {
				one.Layout.Style = para.Runs[0].Style
				one.Layout.Font = para.Runs[0].Font
				one.Color = para.Runs[0].Color
				for _, run := range para.Runs[1:] {
					one.Layout.Style.Leading = math.Max(one.Layout.Style.Leading, run.Style.Leading)
				}
			}
			text.Write(richParagraphXML(one, sceneTableParagraphProperties(one, para.ParagraphGapAfter, para.Bullet)))
		}
	}
	if tr.Rich == nil || len(tr.Rich.Paragraphs) == 0 {
		for _, line := range strings.Split(tr.Layout.Displayed, "\n") {
			pp := sceneTableParagraphProperties(tr, 0, false)
			text.WriteString("<a:p>")
			text.Write(pp)
			properties := sceneTableDefaultRun(tr)
			if line != "" {
				text.WriteString(`<a:r><a:rPr` + properties + `</a:rPr><a:t xml:space="preserve">` + sceneTableXMLEscape(line) + `</a:t></a:r>`)
			}
			text.WriteString(`<a:endParaRPr` + properties + `</a:endParaRPr></a:p>`)
		}
	}
	return []byte(`<a:txBody><a:bodyPr wrap="square" anchor="ctr" lIns="0" rIns="0" tIns="0" bIns="0"><a:noAutofit/></a:bodyPr><a:lstStyle/>` + text.String() + `</a:txBody>`), nil
}
