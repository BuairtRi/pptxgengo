# Wave 1 smoke tests and independent review

Date: October 3, 2026. The operator explicitly requested smoke tests and independent reviews for bugs, performance, and memory. Checks used the repository implementation and temporary outputs. Retained samples and the installed release were not replaced. No commits or pushes were made.

## Review findings and fixes

Two dedicated agents independently reviewed native media optimization and catalog discovery. The integration agent reproduced findings, added regression tests, and applied or delegated fixes. Reviewers checked the resulting changes again.

| Finding | Fix and evidence |
| --- | --- |
| A namespaced `x:Target` could be rewritten instead of the actual relationship target | Interpret core attributes by namespace; select the actual unqualified attribute |
| A regex could match `Target` text inside another quoted attribute value | Replace the regex with a quote-aware attribute scanner; regressions cover both quote forms and extension attribute order |
| Namespaced content-type attributes could override core values or remove a retained Override | Restrict content-type elements to the OPC namespace and attributes to unqualified names; preserve unrelated extension elements |
| Removed payloads and repeated hashing increased memory and work | Reuse read-only STORE payload views with CRC validation, release removed payload references, and cache hashes per group |
| Large cubic scaling allocated substantial intermediate buffers | Keep Catmull–Rom antialiasing; use its Transform path when the separable Scale intermediate would exceed 16 MiB. A large JPEG regression compares every channel to Scale within one value |
| Comparison row tuples were counted as semantic groups | Count complete repeated rows rather than their cells |
| RAID table data discriminators appeared as executable components | Derive scene zones and component types from actual composition nodes |
| Nested biography bullets ranked equally with primary card groups | Distinguish primary, supporting, and nested groups; count-match weights are 35, 15, and 10 respectively |
| Paired comparisons lacked structural comparison metadata | Derive the relationship from paired layout topology and explicit linking, independently of labels |
| Search accepted an engine hint without explaining its effect | Return the requested engine and `not_evaluated_search_only`; help documents wrapper passthrough |
| A weekly-status query tied incidental purpose mentions with direct scenario names | Prioritize direct key/name/family tokens while keeping text ranking capped and structural matching independent |

## Full-library size and resource measurement

Input: the retained 167-slide `samples/wmds-feedback-20261003/latest/WMDS-template-library.pptx`.

Input SHA-256: `181fb24b1d3fc7bb2510a4ceb50693072d98f8c9f8b0e1ad529c02039635c711`.

Measurements are decimal MB/GB. The harness reads the entire input, invokes the public optimizer with delivery defaults, writes the output and receipt, and reports Go allocation statistics. `/usr/bin/time -l` measures the whole process on this Mac, including input/output. These are development measurements, not portable resource guarantees.

| Measure | Initial implementation | After review fixes |
| --- | ---: | ---: |
| Input bytes | 196,841,070 | 196,841,070 |
| Output bytes | 31,550,932 | 31,550,932 |
| Media parts | 560 → 111 | 560 → 111 |
| Unique JPEG derivatives | 8 | 8 |
| Optimizer elapsed | 4.60 seconds | 12.58–13.11 seconds |
| Peak resident memory | 1.38 GB | 650–746 MB across two runs |
| Total Go allocations | 2.91 GB | 1.03 GB |

The output is approximately **84% smaller**. Lower temporary memory costs more CPU time while preserving the cubic resampling kernel. The optimizer still rebuilds packages in memory; this is a material remaining scaling limit. It now refuses individual uncompressed parts above 256 MiB and packages above 2 GiB of uncompressed parts before allocating them, and keeps the 64-million-pixel JPEG decode guard. Streaming packaging and tighter whole-build resource budgets remain future work.

## Package integrity and integration smoke checks

The optimized full library passed:

- ZIP CRC checks and parsing of all 1,032 XML/relationship parts.
- Resolution of all 1,756 internal relationships to existing parts.
- Byte comparison of all 525 retained non-media, non-relationship, non-content-type parts against the input. This includes slide/layout/master content, text, geometry, and object identity.
- Per-part source and output SHA-256 verification against the receipt.
- Byte preservation for all 546 source entries that were not resized, including vector and lossless assets.
- Final output hash verification and confirmation that the retained input hash was unchanged.

Three actual CLI builds exercised the integration:

| Build | Output size | Checks |
| --- | ---: | --- |
| Three-slide photo fixture, default policy | 444,796 bytes | Automatic delivery policy; 12 → 8 media parts; DEFLATE; final receipt hash |
| Same photo fixture, explicit preservation policy | 21,402,947 bytes | 12 media parts; STORE; every media payload unchanged; final receipt hash |
| Bound `cards/3` fixture, explicit preservation policy | 124,624 bytes | Policy survives binding into the compiled document and renderer; STORE; final receipt hash |

These checks establish packaging and policy propagation. They do not establish native PowerPoint appearance or PDF fidelity of the new JPEG derivatives.

## Catalog and authored-source checks

- Discovery regression tests use the pinned 167-template catalog. The normal search excludes its one deprecated entry and therefore examines 166 active candidates.
- Targeted actual CLI queries covered point counts, paired comparisons, weekly status, comparison row counts, and RAID metadata.
- Three-card primary layouts outrank the named biography/supporting examples. Four-point queries return useful choices across multiple scenario families rather than filtering by title.
- Repeated identical queries produce byte-identical JSON; reversing input catalog order does not change results.
- Negative counts, unsupported engines, excessive limits, and an item role without an item count fail. Unknown soft role hints remain visible as unmatched hints instead of becoming silent hard exclusions.
- Independent Draft 2020-12 schema checking passed. The illustrative YAML has no schema violations. This validates the proposed document shape; it does not implement a YAML loader, resolve assets/definition pins, or make that example runnable.

## Go checks and existing baseline failures

All command binaries compile. Focused media tests, large-resampling/content-type/group/rotation regressions, and discovery tests pass. The focused media tests were also run with the race detector.

Repeat the focused tests with a suitable repository Go toolchain:

```sh
go test ./pptx -run '^(TestOptimizeMedia|TestMedia)' -count=1
go test -race ./pptx -run '^(TestOptimizeMedia|TestMedia)' -count=1
go test ./internal/wmdesign ./cmd/pptxdesign
go build -o /tmp/pptxgengo-wave1-bin/ ./cmd/...
```

The broader `go test ./pptx` does **not** pass. The same 12 failing test functions reproduce in a separate `git archive HEAD` checkout with the same cached toolchain/dependencies. The working tree introduces no additional failing functions in that comparison:

- `TestMakeXmlChartsBarGolden`, `TestMakeXmlChartsLineGolden`
- `TestCreateExcelWorksheetBarGolden`, `TestCreateExcelWorksheetPieGolden`, `TestCreateExcelWorksheetLineGolden`
- `TestGoldenIntegration`, `TestMakeXmlPresentation`
- `TestWriteFileRoundTrip`, `TestWriteCompression`, `TestWriteToOptsCompression`, `TestWriteFileOptsCompression`, `TestWriteFileAtomicReplacesExisting`

These existing failures need a separate baseline repair; this review did not update unrelated goldens to make the run green.

## Remaining acceptance and next step

1. Review representative optimized photographs, SVG/raster fallbacks, transparency, and layout/master images in native PowerPoint and exported PDF.
2. Begin Wave 2: implement strict YAML loading and compilation, portable project assets, coherent toolchain pins, build receipts, object mappings, and protected generated baselines.
3. Implement the unified modern SQLite projection and measured-content selection. Current search reports affordances and advisory capacities, not proven fit.
4. Continue versioned intake of the newer design-system templates and frames. Wave 1 documented that intake without silently changing the pinned catalog.

Temporary logs, receipts, optimized decks, fixture sources, and build outputs were written under `/tmp/pptxgengo-wave1-review/`; additional discovery fixture outputs are under `/tmp/wmds-discovery-smoke-20261003/`. Those directories are disposable; this report preserves the conclusions and reproducible commands.
