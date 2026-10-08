# Deck projects and resuming work

`deck.yaml` and the slide, template and notes files it references hold the deck's visible content. Supporting Markdown files hold evidence, audience context, decisions and reviews. Generated PowerPoint, JSON and layout reports are build outputs; never edit them as the source.

## Start a project

```sh
pptxgengo design project create --out ./client-deck --id client-deck \
  --title 'Client deck' --bundle v11 --template cards/3
```

Choose the initial template from [template selection](template-selection.md); `cards/3` above is an example. Replace scaffold copy and update `context/project.md` and the composition log. Add slides with `project scaffold --stock` ([editing slides](editing-slides.md#add-a-slide)), then run `project check`. `project create` already writes and pins the portable project; use `project init` and `project split` for an existing unsplit source project.

New projects use the installed V11 library. Existing projects retain their lock.
For an upgrade, choose an explicit bundle target and follow
[project upgrades](upgrading-projects.md); schema conversion and template
revision changes are separate.
For a historical project, split with the bundle matching its lock.

## Project files

Use the [canonical folder tree and repair workflow](project-structure.md) when
creating files or resuming a project with scattered paths. Supporting documents
belong under `context/`; sources, briefs and notes have their own folders.

Create each file when its stage needs it. Formats are in [project templates](project-templates.md).

| File | Holds | `context` key |
| --- | --- | --- |
| `context/project.md` | Stage, next move, frame with confirmed/working/open status, decision log, style decisions | `project` |
| `context/win-strategy.md` | Internal win themes, competitor concerns, pricing logic | `win_strategy` (never packaged) |
| `sources/` (with `sources/index.md`) | Source files and the index of IDs, locations, authority and limits | `sources` (link the folder to track every file) |
| `context/audience-context.md` | The readers' starting position; no conclusions or win strategy | `audience` |
| `context/outline.md` | Narrative and ordered pages with titles, support and transitions | `outline` |
| `context/claims.md` | Claims under `## <claim-id>` headings; slides' `evidence_refs` are checked against it | `claims` (link it, or `evidence_refs` go unchecked) |
| `briefs/<slide-id>.md` | Each page's job, required content, evidence and open choices | Per slide, via the slide's `brief` field |
| `composition-log.yaml` | Required. Each slide's purpose, candidates, chosen template and rationale; checked by `project check` | `composition_log` (optional; the root file is found automatically) |
| `reviews/` | Reviewer findings and operator decisions | Not tracked |
| `deck.yaml`, `slides/`, `slides/templates/`, `notes/` | Visible copy, order, templates, notes | Tracked automatically |

`context` accepts only `project`, `audience`, `outline`, `sources`, `claims`, `composition_log`, `decisions`, `win_strategy` and `state`. Linked files invalidate dependent approvals when they change. Put internal win logic in a file linked as `win_strategy`; it is never copied into any review packet. Link `decisions` to a decision-log file, not the `decisions/` folder where the CLI writes its receipts. Once the deck source exists, generate review drafts from it and mark older drafts as superseded.

## Build and approve

Command details are in the [CLI reference](cli-reference.md#deck-projects).

- Each `project build` writes a new directory in `builds/` with `deck.pptx`, `layout-report.json`, an object map and a receipt. The receipt tracks Go fit, native rendering and visual review separately; native and visual review stay pending until you render and look.
- V11 builds report automatic density changes on stderr and in the layout report.
  Check those warnings before rendering, especially on overview and introductory
  slides. [Typography density](typography-density.md) describes slide controls.
- Record an approval (`project approve --stage S --actor NAME [--slides IDS]`) only for a real operator decision, with the operator as actor. Stages: `intake`, `framing`, `outline`, `content`, `selection`, `build`, `review`, `delivered`.
- Changing a slide's copy, layout or assets invalidates that slide's affected approvals; other slides' approvals stay valid. `project status` lists what was invalidated.

## Resume

1. Read the project context path declared in `deck.yaml` (normally `context/project.md`) and `composition-log.yaml`, then run `project status`.
2. Check invalidated approvals, and whether the generated deck was edited by hand (a changed baseline blocks build and export). If `check` reports `project.toolchain_drift`, the CLI was upgraded; follow the re-pin steps in the [CLI reference](cli-reference.md#move-a-project-to-a-new-cli-version) with the operator's agreement.
3. Continue from the first unresolved step. Don't redo approved stages whose inputs haven't changed.
4. If someone edited a PowerPoint working copy, preserve it and the immutable build baseline. Use [reviewed text reconciliation](editing-slides.md#rebuild-or-update-an-existing-deck) for supported fields; manually translate geometry, formatting and unsupported content after review.
5. If `state.json` is missing but `builds/` exists, ask the operator to restore it; don't guess.

## Export

```sh
pptxgengo design project export --project ./client-deck --mode client --out ./client-delivery.zip
pptxgengo design project export --project ./client-deck --mode maintainer --out ./maintenance.zip
pptxgengo design project export --project ./client-deck --mode offline --out ./offline.zip
pptxgengo design project review --project ./client-deck --out ./reviewer.zip
```

| Mode | Contains | Use for |
| --- | --- | --- |
| `client` | Current PowerPoint and export manifest; no project files or sources | Sending to the client |
| `maintainer` | Source, context, evidence, assets, decisions, approval history and the current build (private) | Handing the project to another WM author |
| `offline` | Maintainer contents plus the compiler, pinned library, fonts and used assets | Rebuilding on the same OS and architecture without the release |
| `reviewer` / `project review` | PowerPoint, layout report, object map, receipt and visible slide values | Technical review only; build comprehension packets with [review packets](review-packets.md) |

Exports need a current build of the current source, an unedited baseline and a new output path. They only write a local archive. PDF is not included; use `design render` for PDF.

## Keep history

Build experiments in temporary directories. Never prune `builds/`, locks, state, original assets, `decisions/` or review findings; they are the baseline for reconciling later edits.

## Portable numbered projects and colleague handoff

Use the project created above or migrate an existing project's folder layout.

Replace scaffold example copy before approval or delivery. New/split sources use
`slides/<stable-id>.yaml`, local definitions use `slides/templates/`, and owned
asset revisions share `assets/objects/sha256/`. Deck.yaml remains the authoritative
slide order. Preview existing paths with `project layout --project PATH --dry-run`
and apply with `--apply`; historical receipts/source preimages are retained,
and changed dependency approvals must be reviewed. Rebuild after migration.

```sh
pptxgengo design project version save --project ./client-deck --actor 'Named author' --message 'Reviewed source and generated deck'
pptxgengo design project version list --project ./client-deck
pptxgengo design project version verify --project ./client-deck --number 000001
pptxgengo design project share --project ./client-deck --out ./private-colleague.zip
pptxgengo design project share-extract --archive ./private-colleague.zip --out ./colleague-deck
pptxgengo design project share-verify --project ./colleague-deck
pptxgengo design project version materialize --project ./colleague-deck --number 000001 --out ./restored-version
```

Numbered versions retain complete source, generated PowerPoint, exact asset
revisions, pins and receipt/history references. The visible current pointer is
`versions/current.json`. Store stable images once for the entire deck; use
`project asset revise --expect-sha256 OLD_HASH` for an intentional changed asset.
A snapshot author is not an approval actor unless a separate actual approval is
recorded. Source/deck edits remain subject to normal approval invalidation and
native reconciliation.

A complete share is **private** source/evidence handoff, distinct from client
export. ZIP transport deduplicates asset bytes and safely expands legacy path
aliases during `share-extract`; it uses ordinary files and no links. Toolchain
and private branding/font requirements remain pinned. Offline runtime export
is separate. Read the included SHARE.md before resuming in another directory.

OneDrive is synchronization, not a transaction/lock service. Wait for sync and
run share verification after extraction. Preserve both conflicting source/native
copies; never pick an arbitrary winner or remove a build/source/version lock to
force progress. A complete interrupted snapshot can be recovered only with
`version recover --number NUMBER --expect-current-sha256 EXACT_POINTER_HASH`
(or `absent` for the initial unpublished pointer). Incomplete snapshots remain
retained for explicit resolution. Preserve the reported conflict and resolve its exact predecessor before resuming.

### Native editing profile

New v2 projects default to `editing_profile: native-v1`, persisted in `deck.yaml`
and build receipts. Existing projects with absent or `stock` retain their prior
structure; explicitly opt in and create a new baseline. CLI template generation
and the packaged browsing deck also use native-v1. `--editing-profile stock`
selects original structure when needed.

The renderer adapts eligible measured lists, cards and tables while retaining
source styles and geometry. Read conversion/retention reasons in layout reports
and the catalog-wide `native-editing-coverage.json`. Shared template definitions remain pinned. Gallery previews describe original specimens and do not qualify all
native adaptations. Preserve old builds and review the new PowerPoint. Complex
artwork and incompatible geometry retain their structure. Combined card/list text currently requires manual source adoption; do not infer original fields or list keys from edited paragraphs. Install the bundled IBM Plex fonts and review the rendered result; a profile setting does not record visual acceptance.
