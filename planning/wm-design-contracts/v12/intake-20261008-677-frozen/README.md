# V12 product planning template intake

Exact source: `efec671fe40d2145d14780dc39c3128bc9c65308`.
Baseline: V11 `3c56d842ba3abb5f24eb33cf0082be7a7a67f116`.
The upstream worktree was clean. Canonical `id/variant` comparison records 28
additions, 649 unchanged definitions, no removals and no revisions. JSON object
ordering/formatting changes in family files do not change retained compositions.
The candidate bundle lives at `library/wm-design-system/v12`.

- [Complete source delta](update-inventory.json)
- [Committed blob acquisition and byte/hash evidence](intake-audit.json)
- [Ordered additions](new-template-keys.json)
- [Editable content examples](new-template-content.json)
- [Compiled source specimens](new-template-compositions.json)
- [Native preview gallery](native-review/index.html)
- [Primary agent visual review](visual-review.json)
- [Signed native export receipt](native-review/render-manifest.json)

Source acquisition used `git show <exact-commit>:<path>` for every inventory
entry. The historical inventory spelling `templates/changelog.md` maps to the
upstream Git path `templates/CHANGELOG.md`. Thirty-three equal source files are
linked to V11; seven changed files are frozen physically. The source renderer,
frames, components, typography, fonts and asset manifests are unchanged.
Upstream section/vocabulary documentation edits are not renderer dependencies.

## Implemented behavior

The adapter recognizes V12 as a hash-pinned expanded library, preserves the V11
density/contrast rules and all historical named refinements, exposes all 677
catalog entries and supports `--bundle v12` across catalog/search/index/project
routes. Closed bindings preserve all content, array cardinalities and identities.

V12 heat cells accept finite numeric strings used for staffing totals. Their
original display strings (including decimal zeros) are preserved. Invalid or
nonfinite numeric strings are rejected. Earlier pins keep their existing parser.
The browser accepts `deltaUnit: " FTE"` but formats a suffix only for `$M`;
V12 follows this behavior, with FTE identified in column labels and surrounding
copy. The browser also leaves `labelPos: "end"` at the longest connector segment
midpoint; V12 accepts that authored metadata and retains that position.

Four explicit native allocation amendments leave frozen JSON/copy intact:

| Template | Native allocation |
| --- | --- |
| `backlog/epic-hierarchy` | User-story column 354 → 378 pt, consuming the 24 pt slack after the row-group rail. |
| `backlog/epic-overview` | Epic column 186 → 210 pt, consuming the 24 pt slack after the row-group rail. |
| `backlog/wsjf-scoring` | Tier column 60 → 72 pt and backlog-item column 222 → 210 pt, reserving native chip padding at the same total width. |
| `roadmap-strategy/cascade-split` | Outcome column 150 → 156 pt; six rows 63 → 61 pt; source note allocated two lines, preserving reserved footer clearance. |

Automatic density fitting selects Dense for `backlog/story-detail` and
`roadmap-strategy/nav-sections`; all headers remain Comfortable. The full source
and bound sweeps also retain three earlier automatic-density adjustments. Source
stock acceptance does not require all authored values to fit every density.

## Validation

- `go test ./internal/wmdesign -run '^TestLibraryV12' -count=1`: passed, 10.596s.
  Exact catalog delta, all 649 retained compiled compositions, 28 source/bound
  content roundtrips, and 56 individual addition builds.
- `go test ./internal/wmdesign -short -count=1`: passed, 82.170s.
- Focused CLI shorthand/default/search tests: passed, 9.240s, including V12 routes.
- Full stock generation: 677 source and 676 active bound specimens passed.
- SQLite discovery: 677 templates, 1,204 assets, 38 components, 29 composites,
  36 frames and 13 primitives. Rebuildable index; no candidate previews certified
  by the index itself.
- Native PowerPoint PDF/PNG export: 28/28; every issued full-size PNG was viewed
  by the primary agent, with no remaining clipping or overlap findings. Footer
  and release-card crops were additionally inspected. This is primary-agent
  review of the effective source/automatic densities, not independent review.

The published V11 release/defaults are retained. V12 publication requires its
independent review and retained-render inheritance evidence; this implementation
does not assert those gates or alter installed user tools.

## Rebuild

Run from the repository root; each output directory must be new.

```sh
go build -o /tmp/pptxdesign-v12 ./cmd/pptxdesign
intake=planning/wm-design-contracts/v12/intake-20261008-677-frozen
/tmp/pptxdesign-v12 library-source-reference --bundle v12 --template-keys-file "$intake/new-template-keys.json" --year 2026 --out /tmp/v12-new-source
/tmp/pptxdesign-v12 library-reference --bundle v12 --template-keys-file "$intake/new-template-keys.json" --year 2026 --out /tmp/v12-new-bound
/tmp/pptxdesign-v12 library-source-reference --bundle v12 --year 2026 --out /tmp/v12-all-source
/tmp/pptxdesign-v12 library-reference --bundle v12 --year 2026 --out /tmp/v12-all-bound
/tmp/pptxdesign-v12 render --pptx /tmp/v12-new-source/library-reference.pptx --out /tmp/v12-native --pdf --png --contact-sheet
/tmp/pptxdesign-v12 library-index --bundle v12 --out /tmp/v12-library.sqlite
```
