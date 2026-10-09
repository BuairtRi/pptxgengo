# Working in pptxgengo

## Choose the current workflow

- Deck operators: read [the West Monroe presentation skill](skills/west-monroe-presentations/SKILL.md) and only the references needed for the task. Use `pptxgengo design project`; inspect installed help and project toolchain pins before relying on main-only features.
- Repository development: use the registered `pptxgengo` repository through `slotctl` and its local development skill. Work in a managed slot. Publish through the configured dual `origin` push URLs; GitHub is review/merge authority and private GitLab runs all CI and holds release artifacts.
- Read [the documentation index](docs/README.md) for maintained contracts. Dated plans and qualification ledgers describe their named inputs, not universal capability.

## PowerPoint failures and recovery

Before diagnosing a native render, GUI, AppleScript or permission failure, read
[PowerPoint automation diagnostics](docs/powerpoint-automation-diagnostics.md).
It records failure layers, error evidence, observed recovery patterns, tested
code fixes, and unresolved causes. Deck operators should use the focused
[PowerPoint recovery runbook](skills/west-monroe-presentations/references/powerpoint-recovery.md).

Record the exact failing phase/code, caller, effective execution policy, owned
document path, and last successful action. Treat GUI capture, Apple events,
caller filesystem access, PowerPoint document access and stale handles separately.
Use fresh computer-use state after resuming or rebinding. Preserve failed attempts.
Only an observed denial establishes denied access; successful app health does
not qualify opening or exporting a document. Do not reset privacy settings,
kill PowerPoint or close other people's presentations as speculative recovery.

## Composition and native editing

Templates are customizable visual examples. People, roles, pods, tiers, tasks,
phases and gates vary with the intended message. Read the applicable family
runbook before changing its topology or scale. Preserve stable authored IDs,
frame allocations, source facts, units, relationships and explicit color meanings.

Keep generated builds immutable; edit a separate PowerPoint copy. Review and
explicitly adopt supported changes, retaining unknown formatting and objects.
Native layout changes do not imply changes to dates, membership or reporting.
Imports need explicit mappings and a declared formatting policy; do not imply
arbitrary PowerPoint objects or Office versions are qualified.

Use previews and measured fit before guarded source apply. Preserve predecessors,
native overrides and template pins. Keep operator instructions in skill references;
implementation history, qualification limits and future contracts belong in `docs/`.

## Delivery

Keep normal GitLab PR/main checks light; broad race, desktop and catalog suites
use the agreed nightly/on-demand lanes. Run one final manifest preflight through
`slotctl slot submit` for a merge candidate; do not duplicate it immediately before
submission. If the operator only requests a commit/push checkpoint, use
`slotctl slot publish` without creating a PR. Verify both remote object IDs.
Retire clean, integrated slots normally through slotctl; preserve any unintegrated
work and do not use force/discard as routine cleanup.
