# Conversion usability observations

Task: make new maintained-format versions of three existing decks after587 visual qualification and slide section support. Preserve originals; native review all outputs. Presentation skill owned by another agent.

| Observation | Evidence / implication | Action |
| --- | --- | --- |
| Sample inventory omitted ignored incoming decks | Ordinary rg--files saw only pickdeck; --no-ignore revealed Patterson/DentalXChange originals | Use explicit input paths and include ignored originals in import discovery; no alteration of ignore policy |
| Default wrapper overrode old project pins | Existing wrapper injectedv3 bundle even forproject commands | Fixed wrapper delegates project selection to lock; new stages read default-bundle config; exact runtime pins remain enforced |
| Optional chart decimals render a bare separator | Native gallery labels1./0./$1.M | v5generatedformat keeps precision using required fractional digits; quoted source units preserved |
| Native labelposition schema alone is insufficient | Candidate doughnutdLblPos parsed/tested butPowerPoint demandedrepair | Rejectedcandidate; keepnativeopen as gate; supportedhole50geometry accepted |
| Global legend reserve can create footer wrap | Three fullsource legendsrejected despite isolatedpositivecases | Consumeexistinggap forreserve, retainpackingfootprint; full7033node audit nowpasses |

Record build timings, repeated render/review loops, YAML edit pain and missing commands during actual conversions. Prefer fixes supported by an observed task; do not expand skill writing.

## Conversion run: source inspection and section pilot

- All179 source pages were individually inspected in native PowerPoint renders. Original source bytes are unchanged.
- Catalog search required a separately built SQLite index; `library-find --bundle v5 --query …` was insufficient. Bundle help/default messaging must match the selected runtime. Useful improvement: resolve the packaged index automatically and explain how to build an absent index.
- `--help` exits1 (`flag: help requested`). This surprises ordinary CLI discovery and stops shell workflows that condition subsequent discovery on success.
- Existing opaque `nodeXX` slots require reading large contracts. Named local content zones and a concise template structure/required-values view would reduce YAML edits.
- A matching candidate name is insufficient: seven-row guides, three-lane Gantt, four/four/four lifecycle and two/three/two seven-R grouping differ from original4-entry guides,10-row schedule,4/4/3 lifecycle and3/2/2 disposition. Record topology decisions before binding; derive local composition when necessary.
- Section YAML edits now preserve comments and scalar styles, retain exact predecessors and report visible-divider rename/removal behavior. Native PowerPoint verified Unicode groups, hidden state, notes paragraphs and local PDF output.
- Native PDF export omits hidden slides by design. Reviews require an explicitly marked diagnostic all-visible copy; final deck retains original hidden flags.

## Final conversion run: measured friction and recovery

- `project check` passed source/pin validation while `project build` caught unsupported connector `stroke` and absolute points outside the component. The commands should say which validation layers they cover. A dry compilation verb, or stronger check mode, would expose these errors sooner.
- Asset derivation JSON required `pptxgengo.asset-derivation.v1`, `source_asset`, `source_sha256`, `result_sha256` and `operation`. The first attempted receipt failed with `unknown field "source"` without useful asset/path context. Provide a receipt scaffold and path-qualified errors.
- Local connectors default to an end arrow and use absolute slide coordinates; a dashed legend swatch needs `head:none`. The team curve accepts stroke width but connector does not. Supported argument inspection would avoid trial builds.
- Exact source20 logo cropping required recovering original DrawingML fractions, deriving a separate image and pinning both original and result. A crop/import verb could generate the strict receipt automatically.
- Source25 native icons lost white constituent shapes in the first conversion. Full source copy and media hash retention did not detect the visual loss. Group-aware import should preserve fills, outlines, relative geometry and complete constituents.
- Reformatting JSON-as-YAML with yaml.v3 exposed mixed literal-indent problems and49 leading-newline notes being dropped. The rejected candidates were never applied. A fresh node serializer quoted leading-newline notes; parsed source equality and entire emitted PPTX equality proved the readable replacement. A formatting command should require those guarantees.
- Final readable Software YAML still contains more than80,000 lines of composition geometry. Stable slide IDs and separately delivered source-to-page maps make navigation easier; a content-only editor or `project slide show/set` would avoid searching that file for ordinary copy edits.
- PowerPoint local PDF export of the focused two-page grid took about50 seconds; the183-page deck took about60 seconds to reach completion. These are observed wall times, not controlled performance benchmarks. Small changed-page review packets reduce repetitive full-deck exports; exact native pixel/package carry receipts make unchanged acceptance traceable.
- Native PDF export deliberately omits hidden pages. The final all-visible inspection copy changes only five `show` flags, preserving the preferred deck. A built-in inspection export should make this distinction explicit.
- The final183-page source retains every original source paragraph, authored note, hidden state and12 section memberships. Diagram topology, phase associations, source icons and curve legends required visual/semantic repair beyond copy validation.

Implemented during this run: native sections and optional library dividers; exact chart-category casing and workbook zero opt-ins; local curve stroke reserve; stable delivery manifests and readable Software YAML with semantic/byte proof. Further CLI proposals are in `samples/wmds-migration-20261003/CLI-YAML-IMPROVEMENTS.md`.
