# WMDS v10 committed intake

Frozen upstream commit: `c14fb286fb38e15800a6fd476a1ed67956f1165f`.
649 templates: 648 active and one deprecated. Workshops remain 29.

Every source file was acquired using `git show` at that exact commit. Upstream
working tree edits, including density work, are excluded. The bundle contains
40 source paths: 12 new or changed physical files and 28 verified links into the
immutable v9 intake. Fonts, assets and typography evidence retain v9 identities.

## Delta from qualified v9

18 additions: five argument pillar compositions, seven branching roadmaps and
six roadmap narratives. All 631 retained template objects, including their
slide compositions, are unchanged. Twenty-three retained templates move into
the Roadmaps family; 25 catalog rows change family or section metadata.

The exact addition order follows the upstream catalog: five pillars, seven
forks, six narratives. See [ordered keys](new-template-keys.json) and the
[comma-separated CLI keys](new-template-keys.txt).

[Intake audit](intake-audit.json) records the full delta and each committed Git
blob identity, SHA-256, byte count and storage relationship.

- Bundle SHA-256: `45d25e4d920165425a661aa8ceea36a24b979070b673f547f2294d0a65d109c0`
- Inventory SHA-256: `aea092e8e5e1aca02900ab87b90b294d014ac2c19937049f735e1ad30a7c9124`

## Engineering recipe

Run from the pptxgengo repository. Use a new temporary output directory for each
reference generation. The explicit bundle path avoids changing a CLI default.

```sh
go test ./internal/wmdesign -run 'TestLibraryV10(FrozenSnapshotClosure|PinnedIntakeAndV9Inheritance)$' -count=1
go test ./internal/wmdesign -run TestLibraryV10NewTemplatesRoundTripAndBuild -count=1
go test ./internal/wmdesign -run TestLibraryV10AllTemplateBuilds -count=1
go test ./internal/wmdesign -run TestLibraryV10RetainedPublicationRenderInheritance -count=1
go test ./internal/wmdesign -run TestLibraryV10RetainedIndividualSourceBuilds -count=1
go test ./internal/wmdesign -run TestLibraryV10AllocationsPreserveContentAndAreAtomic -count=1

go build -o /private/tmp/pptxdesign-v10 ./cmd/pptxdesign
intake=planning/wm-design-contracts/v10/intake-20261006-649-frozen
keys=$(cat "$intake/new-template-keys.txt")
/private/tmp/pptxdesign-v10 library-source-reference --bundle "$intake/bundle" --template-keys "$keys" --year 2026 --out /private/tmp/wmds-v10-new18-source
/private/tmp/pptxdesign-v10 library-reference --bundle "$intake/bundle" --template-keys "$keys" --year 2026 --out /private/tmp/wmds-v10-new18-bound
```

The full catalog has 649 source specimens and 648 active bound specimens.
Generation tests establish source and binding execution, not native acceptance.

## Bounded native allocation corrections

The frozen source remains byte-identical to the upstream commit. The adapter
applies six named, guarded corrections while preserving copy, fonts and values:

- `road-fork/parallel`, `road-fork/decision`,
  `road-fork/foundation-then-waves`: road height 324 → 306 pt for footer clearance.
- `roadmap-narrative/objectives-tree`: delivery column 248 → 292 pt to fill the
  816 pt table allowance after its row-group rail.
- `roadmap-narrative/waves-criteria`: two title lines, table top 126 → 162 pt,
  four row heights 72 → 65 pt, preserving the reserved footer boundary.
- `roadmap-narrative/horizon-table`: horizon/investment columns 108/72 → 90/90 pt,
  preserving the 846 pt table width and the complete investment header.

The roadfork primitive uses native editable cubic paths and measured text.
Its branch and current-position tags reserve 2 pt of native text clearance;
numeric zero milestone labels use the source renderer's ordinal fallback.
Preservation tests check idempotence, unchanged original objects and atomic
rejection of unexpected geometry. No density or font policy is introduced.

The final native export's 18 exact stock specimens passed independent full-size
visual review. Branch tags and the current-position label are complete, table
headers fit, and the two-line waves title is present with clear footer spacing.
[Independent review evidence](independent-native-review.json) records every
ordered key and PNG SHA-256. This acceptance covers these stock specimens.

## Publication inheritance

Existing compositions can inherit their accepted v9 native specimens only after
the publication dependency comparison succeeds. Family file moves change source
paths and file hashes, including hashes for unchanged templates sharing revised
family files. Refresh current contracts and discovery metadata while retaining
the exact composition, text, chart, workbook, theme and media dependency checks.
The 18 additions require their own native exports and visual review.

The closure, retained source/binding/composition checks and paired visible-package
comparison passed together in 72.801 seconds. All 631 retained specimens matched
their v9 rendering dependencies. This comparison covers 23 family-file moves:
nine from diagrams, twelve from software and two from approach.

After restoring nav chrome error propagation, all 631 retained specimens also
passed individual source builds in 66.716 seconds. That sweep continues through
the full retained catalog to expose every failure rather than stopping at the
first specimen. The manifest comparison separately locks all 15 font, license
and asset identities to v9.

The published v10 gallery's inheritance gate subsequently passed in 73.631
seconds: 649 stock specimens, 631 inherited with equal render dependencies and
18 newly reviewed native specimens. Arbitrary authored content remains outside
this specimen qualification.

This intake does not itself promote a production bundle or change a default.
