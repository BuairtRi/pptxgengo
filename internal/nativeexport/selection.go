package nativeexport

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PageMapping preserves source identity when selected or hidden slides change PDF order.
type PageMapping struct {
	Page         int    `json:"page"`
	SourceSlide  int    `json:"source_slide"`
	SourcePart   string `json:"source_part"`
	SourceHidden bool   `json:"source_hidden"`
	PNG          string `json:"png,omitempty"`
}

// ParseSlides accepts one-based numbers and inclusive ranges, in presentation order.
func ParseSlides(selector string, total int) ([]int, error) {
	if selector == "" {
		numbers := make([]int, total)
		for i := range numbers {
			numbers[i] = i + 1
		}
		return numbers, nil
	}
	seen := map[int]bool{}
	for _, item := range strings.Split(selector, ",") {
		ends := strings.Split(strings.TrimSpace(item), "-")
		if len(ends) > 2 {
			return nil, fmt.Errorf("invalid --slides item %q", item)
		}
		start, err := strconv.Atoi(strings.TrimSpace(ends[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid --slides item %q", item)
		}
		end := start
		if len(ends) == 2 {
			end, err = strconv.Atoi(strings.TrimSpace(ends[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid --slides item %q", item)
			}
		}
		if start < 1 || end < start || end > total {
			return nil, fmt.Errorf("--slides item %q is outside 1-%d or reversed", item, total)
		}
		for n := start; n <= end; n++ {
			seen[n] = true
		}
	}
	result := make([]int, 0, len(seen))
	for n := range seen {
		result = append(result, n)
	}
	sort.Ints(result)
	return result, nil
}

var slideIDList = regexp.MustCompile(`(?s)<(?:[A-Za-z_][\w.-]*:)?sldIdLst\b[^>]*>.*?</(?:[A-Za-z_][\w.-]*:)?sldIdLst\s*>`)
var slideIDElement = regexp.MustCompile(`<(?:[A-Za-z_][\w.-]*:)?sldId\b[^>]*(?:/>|>\s*</(?:[A-Za-z_][\w.-]*:)?sldId\s*>)`)

func selectedReview(original []byte, selector string, include bool) ([]byte, int, []string, []PageMapping, error) {
	_, total, hidden, err := reviewCopy(original, false)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	z, _ := zip.NewReader(bytes.NewReader(original), int64(len(original)))
	ordered, err := presentationSlides(z)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	numbers, err := ParseSlides(selector, total)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	hiddenSet := map[string]bool{}
	for _, p := range hidden {
		hiddenSet[p] = true
	}
	selected := map[int]bool{}
	mappings := []PageMapping{}
	for _, n := range numbers {
		selected[n] = true
		if include || !hiddenSet[ordered[n-1]] {
			mappings = append(mappings, PageMapping{Page: len(mappings) + 1, SourceSlide: n, SourcePart: ordered[n-1], SourceHidden: hiddenSet[ordered[n-1]]})
		}
	}
	if len(mappings) == 0 {
		return nil, 0, nil, nil, fmt.Errorf("selection has no visible slides; use --include-hidden to render hidden slides")
	}
	data := original
	if len(numbers) != total {
		var buffer bytes.Buffer
		w := zip.NewWriter(&buffer)
		for _, f := range z.File {
			if f.Name != "ppt/presentation.xml" {
				if err = w.Copy(f); err != nil {
					return nil, 0, nil, nil, err
				}
				continue
			}
			r, e := f.Open()
			if e != nil {
				return nil, 0, nil, nil, e
			}
			body, e := io.ReadAll(r)
			r.Close()
			if e != nil {
				return nil, 0, nil, nil, e
			}
			list := slideIDList.Find(body)
			ids := slideIDElement.FindAll(list, -1)
			if len(ids) != total {
				return nil, 0, nil, nil, fmt.Errorf("cannot safely select slides: presentation list has %d entries, expected %d", len(ids), total)
			}
			index := 0
			changed := slideIDElement.ReplaceAllFunc(list, func(id []byte) []byte {
				index++
				if selected[index] {
					return id
				}
				return nil
			})
			// Replace only the presentation list. Section extension lists also use
			// sldIdLst and must never be replaced with presentation relationship IDs.
			location := slideIDList.FindIndex(body)
			body = append(append(append([]byte{}, body[:location[0]]...), changed...), body[location[1]:]...)
			dest, e := w.Create(f.Name)
			if e != nil {
				return nil, 0, nil, nil, e
			}
			if _, e = dest.Write(body); e != nil {
				return nil, 0, nil, nil, e
			}
		}
		if err = w.Close(); err != nil {
			return nil, 0, nil, nil, err
		}
		data = buffer.Bytes()
	}
	review, _, madeVisible, err := reviewCopy(data, include)
	return review, total, madeVisible, mappings, err
}
