# Semantic color survey

## Scope and evidence

This is a proposal for discussion, not a unified multibrand policy API. Local authority is the external `branding/DESIGN.md` source (palette: Grounded Blue `#070154`, white, yellow `#F6EB21`, magenta `#F900D3`, Highlight Blue `#0047FF`, Dark Gray `#50658E`, Medium Gray `#CED7E6`, Light Gray `#E8EEF8`). It explicitly limits yellow to the marker behind 1–4 main-headline words, blue to small text/rules rather than general fills, and magenta to sparing emphasis with navy as the preferred ink. Existing `library/component-taxonomy.json` describes neutral/subtle/inverse/blue-accent/magenta-accent/headline-highlight as proposed presentation profiles and says semantic data colors are separate. `library/component-styling.md` likewise says bindings must target verified objects or text runs and retain transforms. Model two explicit modes: `preserve_reference` retains pinned source colors; `wm_current` overlays only declared presentation roles. Brand overlays never remap status, actor, phase or data colors automatically.

### Current behavior observed

| Area | Implemented evidence | Implication |
|---|---|---|
| General adapt style | `internal/adapt/style.go` accepts `text.primary`, `text.secondary`, `surface`, `surface.muted`, `accent`, `state.active`, `state.inactive`, `border`; `ResolveStyle` defaults to navy/white/light-gray/blue and only implements neutral, subtle (`#F4F6FA`), inverse (`#070154`) profiles. | Runtime roles exist, but profile palette does not match the proposed six profiles; subtle is not the brand Light Gray. |
| Components | `internal/compose/cards.go` defaults numbered-card accent to `#0047FF`, rendered as a 5pt side rule, not a background; adaptors such as process/architecture use semantic accent/active/muted roles and derive ink with `Style.Ink`. | The side rule is consistent with blue-as-emphasis. Audit actual active fills family by family; do not characterize the card accent as a surface conflict. |
| Status/semantic data | `internal/adapt/roadmap.go` maps statuses through `roadmapStatusColor`; comparison rows map strong/mixed/limited/unknown; team composition requires a generated legend for semantic staffing colors. | Status, comparison meaning, and actor ownership are already separate concerns in some adaptors. Preserve these mappings and legends when changing presentation profiles. |
| Gauge control | `cmd/pptxtemplate/gauge.go`, `library/templates/rollout/evidence_people/t045-graphics-and-layouts-045/gauge-authoring.md` and `planning/requested-templates/gauge-review.json` establish exact values: pink `#F900D3`, navy `#070154`, blue `#0047FF`, selected gray `#7F7F7F`, inactive gray `#CED7E6`. Pointer cell conveys discrete intensity position. | Pink is confirmed brand magenta and the two grays deliberately distinguish selected from inactive. Preserve the verified mapping and legend/state. |
| Contract census | Root survey reports 101 template contracts, 11 with roles/profiles; 124 source subcontracts, 20 with roles/profiles. Recorded profile vocabulary includes `source`, `series_match`, `cool_emphasis`, `blue_emphasis`, and gray/navy/blue/pink source accent profiles (these profile names do not by themselves establish gauge semantics). Sampled portable metric-card/pod-two contracts have empty roles/profiles; `t027-software-modernization-035/contract.json` is one example with them. | Roles/profiles are already implemented in a subset; coverage and schema semantics vary. This is a governance/interoperability gap, not total absence. |
| Headline accents | `cmd/pptxanchor`, `cmd/pptxtemplate/accents.go`, and rollout accent evidence implement anchored underline/highlight operations; library checkpoint material records review evidence. | Anchoring is an existing bounded capability. Remaining gap is consistent role/profile policy and coverage, not absence of headline anchoring. |

## Proposed architecture: brand tokens, semantic roles, bindings

Keep three layers distinct:

1. **Brand palette tokens** are immutable source values, e.g. `brand.groundedBlue`, `brand.lightGray`, `brand.magenta`.
2. **Presentation roles** describe visual function (`surface`, `text.primary`, `text.secondary`, `accent`, `border`, `surface.muted`, `headline.marker`). A named brand profile resolves these roles to palette tokens.
3. **Semantic roles** encode information (`status.atRisk`, `status.complete`, `actor.client`, `actor.consultant`, `series.1`, `phase.current`, `intensity.4`). They own their own mapping and legend; a presentation profile must not overwrite them.

Each reusable component declares role-capable slots and source bindings. A binding identifies an exact source object/property or text run, carries provenance (source hash/object id/run path), and states whether it is a presentation or semantic role. It may preserve OOXML alpha/tint transforms. Resolution is `profile role -> palette token -> concrete color`, followed by contrast and role-specific policy checks. Unbound source colors remain source colors. Ambiguous/inherited theme colors remain unresolved until effective color is known. Images, logos, and baked-in artwork are outside simple fill bindings.

Contracts should explicitly choose `preserve_reference` or a `wm_current` overlay. The latter binds only presentation roles and retains explicit status/actor/data roles and legends. Never infer that a source profile named `blue` or a color value matching brand blue is presentation-only.

### Proposed WM presentation profiles

| Profile | surface / ink / secondary / accent / border | Intended bindings |
|---|---|---|
| `wm-neutral` | white / navy / slate / navy / medium gray | default text, metrics, neutral diagrams |
| `wm-subtle` | Light Gray / navy / slate / navy / Medium Gray | quiet panel or row surface |
| `wm-inverse` | navy / white / white / white / white | bounded inverse component; verify secondary copy contrast |
| `wm-blue-emphasis` | white / navy / slate / Highlight Blue / Medium Gray | text, small heading or rule only; never generic surface fill |
| `wm-magenta-callout` | white / navy / slate / magenta / Medium Gray | large callout/rule/emphasis, navy foreground on magenta fill; test actual text size |
| `wm-headline-marker` | white / navy / slate / yellow / Medium Gray | headline marker layer only, anchored to 1–4 words on one line; never generic `accent` |

The word “profile” selects role values; the component binding decides where the role applies. Avoid a universal `accent` fill behavior: accent may mean text/rule, and headline yellow requires a dedicated marker role and geometry. Slide-164 source preview notes recorded in taxonomy also require ≥20pt, one line, 1–4 words, at most one marker per headline.

## Distinguish presentation from information color

| Semantic domain | Examples | Rule |
|---|---|---|
| Status | planned, active, complete, at risk, blocked; comparison strong/mixed/limited/unknown | Keep stable status-to-token mapping across profiles. Ensure label/icon/pattern also conveys status; no status role may alias decorative accent by default. |
| Actor / responsibility | client, consultant, joint owner; staffing roles | Preserve actor identity colors and legend, including the team adaptor's legend requirement. A role tile surface may be restyled only if actor encoding stays explicit. |
| Data series / rating | chart series, 1–5 gauge intensity, source gray scale | Treat as data palette with legend/scale order and source provenance. Never map to brand presentation tokens just because their RGB happens to match. |
| Phase / chronology | current, future, completed phase | Preserve phase order and state distinction; keep chronology connectors and labels independently bound. |

Color resemblance does not establish semantic equivalence: Highlight Blue is not automatically “active”; gray is not automatically “inactive”; magenta is not automatically “risk”; yellow is not a generic “attention” state.

## Conflicts and decisions to resolve

1. **Blue fills:** brand guidance says Highlight Blue is not a general background. The numbered-card component uses it as a 5pt side rule (appropriate small emphasis), while some adaptors use `Style.Active` as a fill. Decide whether each active fill is presentation or information state. Do not global replace.
2. **Yellow highlight:** existing taxonomy exposes it as a profile accent; this risks general-fill use. Restrict it to `headline.marker` and enforce one-line text-run bounds/marker geometry. Until implemented, classify the profile as proposed only.
3. **Source gray versus brand gray:** gauge's selected gray `#7F7F7F` and inactive gray `#CED7E6` deliberately differ. Preserve this verified intensity encoding; any remap needs an explicit data-role contract.
4. **Magenta contrast:** the design brief reports navy on magenta passes AA; white on magenta (~3.5:1) is only for large text (18pt regular or 14pt bold). Generic runtime `Style.Ink` picks the higher contrast of navy/white but does not encode large-text thresholds. Magenta text on white also needs review; use as callout fill or large accent only after checking the actual pair.
5. **Runtime profile drift:** proposed Light Gray is `#E8EEF8`, but `ResolveStyle(subtle)` uses `#F4F6FA`; taxonomy inverse secondary is white while runtime sets secondary `#CED7E6`; runtime exposes state active/inactive beyond the simple proposed profiles. Align canonical token names/values without assuming every extant contract uses these runtime profiles.
6. **Contract coverage:** some current source contracts preserve source theme, fonts and encodings; others explicitly declare roles and profiles. Keep per-contract evidence/provenance and require effective-color certainty only for the bindings being changed.

## Priority decisions

1. Adopt separate `presentation`, `status`, `actor`, `data`, `phase` namespaces; require explicit object/run bindings and profile IDs.
2. Distinguish `preserve_reference` from `wm_current` overlays; apply brand roles only to explicitly bound presentation objects and leave data roles unchanged.
3. Decide whether current blue active fills are presentation or information encoding, family by family. The numbered-card blue side rule is a small accent and aligns with guidance.
4. Preserve the verified gauge mapping; only change it through a separately declared data-role/profile contract, retaining inactive and selected gray distinction.
5. Promote only neutral, subtle, inverse first. Add blue, magenta and headline-marker options after role-specific contrast and geometry behavior is explicit. Existing bounded underline/highlight anchoring is implemented and should be reused.
6. Record proposed versus implemented states in catalog/contracts; role/profile coverage is uneven across current contracts.

## Proposed vs implemented

| Capability | Status |
|---|---|
| WM palette and usage guidance | Local authoritative design source present |
| Taxonomy profile names and values | Proposed metadata present; not universal runtime behavior |
| Semantic style overrides in adaptors | Partially implemented (`internal/adapt/style.go`) |
| Per-object/run contract role bindings | Present in some rollout contracts; absent in sampled portable metric/pod fixtures; coverage is uneven |
| Unified multibrand token registry | Not implemented; this document proposes the WM profile layer as one brand namespace |
| Status/data/actor separation | Partial in roadmap/comparison/team/gauge code and selected source contracts; mappings are component-specific |
| Contrast/policy enforcement by role | Partial: primary text vs surface, and local adaptor checks; no role-wide policy validator |
| Headline marker / underline anchoring | Implemented for bounded source-backed operations with evidence; not a universal dynamic headline profile/role API |
