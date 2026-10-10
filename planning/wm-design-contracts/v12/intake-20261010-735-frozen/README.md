# V12 expanded frozen intake — 2026-10-10

Upstream `36132d5637abdbdeb70945ad650b795cabea05ef`: **735 templates,
734 active**. Compared with the preserved October 8 V12 candidate: 58 additions,
two revisions and 675 retained identities/compositions. Compared with V11:
86 additions, two revisions and 647 retained specimens.

## Implemented source contracts

Measured `toclist` rows preserve numbered gutters, hierarchy, page numbers,
dot leaders and group rules. Source `preset` markers remain native editable
rectangles/diamonds. Page-only schedules and table bullet `{lead,text}` objects
have explicit bindings. The original backlog, roadmap and resource-plan V12
contracts remain isolated from historical revisions.

Three named allocation repairs preserve business facts: the split Gantt reserves
7pt for its final milestone marker and moves its label 1pt into the allocated
body; the split contents list keeps its final .75pt rule inside the footer;
revised `plan/gantt` gives the one-row Change caption a 6pt inset. Dates, durations,
copy, identities, marker semantics and topology remain unchanged.

## Qualification

- All 735 native-v1 source specimens build; all 734 active stock-content bindings build.
- The 804-slide template browsing deck builds with explicit media placeholders.
- All 88 new/revised specimens exported through macOS Microsoft PowerPoint in both stock and native-v1 profiles to PDF and 1920×1080 PNGs.
- The primary agent inspected all 88 native-v1 pages on labeled six-page sheets. 71 stock pages are pixel-identical; all 17 differing stock pages were inspected separately. Differences concern native bullet treatment. No visible clipping or overlap was found.
- The gallery publisher accepted all 88 fresh pages, verified their exact slide bytes and visible dependencies against a reconstruction, and inherited 647 V11 previews after exact composition and paired rendered dependency comparison.
- Search index regenerated: 735 templates, 1,204 assets, 38 components, 29 composites, 36 frames, 13 primitives.

[Accepted stock review](native-review/accepted-review.json), native PDF/PNGs and
source deck are retained beside the [frozen bundle](bundle). The signed local
render receipt is unmodified and refers to its original temporary source path;
the acceptance receipt hashes the retained sibling deck. Additional native-v1
render artifacts remain in `/tmp/v431-native-review`. Full local build hashes
are in [build evidence](build-evidence.json). The original October 8 intake and
its evidence are preserved in the sibling directory.

These checks qualify source specimens. Arbitrary supplied copy, other densities,
Windows PowerPoint and native edit/Save As round trips need separate acceptance.
The placeholder browsing deck's desktop copy/paste qualification is pending.

## Reproduce

```sh
go build -o /tmp/pptxdesign-v431 ./cmd/pptxdesign
/tmp/pptxdesign-v431 library-source-reference --bundle v12 --editing-profile native-v1 --year 2026 --out /tmp/v431-source-new
/tmp/pptxdesign-v431 library-reference --bundle v12 --editing-profile native-v1 --year 2026 --out /tmp/v431-bound-new
/tmp/pptxdesign-v431 browsing-library --kind templates --frames catalog --media-policy placeholders --bundle v12 --as-of 2026-10-10 --out /tmp/v431-browsing-new
```

Release publication uses protected GitLab gates. Local evidence does not substitute
for signed package scans, platform signing or independent download verification.
Slotctl was bypassed under the user's explicit instruction; the original dirty
checkout was preserved at `/tmp/pptxgengo-v431-original-worktree.tar.gz` before work.
