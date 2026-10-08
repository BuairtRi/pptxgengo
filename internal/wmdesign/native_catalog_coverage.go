package wmdesign

import "strings"

// NativeCatalogCoverage describes emitted objects, not desktop qualification.
// Every retained template has an entry, including ones with no eligible scenes.
type NativeCatalogCoverage struct {
	Schema         string               `json:"schema"`
	EditingProfile string               `json:"editing_profile"`
	Templates      int                  `json:"templates"`
	NativeLists    int                  `json:"native_list_boxes"`
	NativeCards    int                  `json:"native_card_shapes"`
	NativeTables   int                  `json:"native_tables"`
	Entries        []NativeCatalogEntry `json:"entries"`
	Qualification  string               `json:"qualification"`
}

type NativeCatalogEntry struct {
	Template     string   `json:"template"`
	SlideID      string   `json:"slide_id"`
	NativeLists  int      `json:"native_list_boxes"`
	NativeCards  int      `json:"native_card_shapes"`
	NativeTables int      `json:"native_tables"`
	Decisions    []string `json:"decisions"`
}

func TemplateNativeCoverage(coverage BrowsingCoverage, report Report) NativeCatalogCoverage {
	out := NativeCatalogCoverage{Schema: "pptxgengo.native-template-coverage.v1", EditingProfile: report.EditingProfile, Entries: []NativeCatalogEntry{}, Qualification: "compiler_object_structure_and_measured_geometry_only; catalog-wide PowerPoint visual and editing review pending"}
	byID := map[string]SlideReport{}
	for _, slide := range report.Slides {
		byID[slide.ID] = slide
	}
	for _, entry := range coverage.Entries {
		if entry.Kind != "template" {
			continue
		}
		slide := byID[entry.SlideID]
		row := NativeCatalogEntry{Template: entry.Key, SlideID: entry.SlideID, NativeTables: len(slide.Tables), Decisions: []string{}}
		for _, text := range slide.Texts {
			if text.NativeParagraphContract == EditableListContract {
				row.NativeLists++
			}
			if text.NativeShape != nil && text.NativeShape.ParagraphContract == EditableCardContract {
				row.NativeCards++
			}
		}
		for _, scene := range slide.Scenes {
			for _, warning := range scene.Warnings {
				if strings.HasPrefix(warning, "native-v1 ") {
					row.Decisions = append(row.Decisions, warning)
				}
			}
		}
		out.Entries = append(out.Entries, row)
		out.NativeLists += row.NativeLists
		out.NativeCards += row.NativeCards
		out.NativeTables += row.NativeTables
	}
	out.Templates = len(out.Entries)
	return out
}
