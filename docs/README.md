# Documentation

Use the current operator guides for deck work. Dated plans, qualification ledgers
and inherited upstream docs describe the exact checkpoint named in each file;
they do not override the installed CLI or a project's lock.

## Agent operator guidance

- [West Monroe presentation skill](../skills/west-monroe-presentations/SKILL.md): intake, voice, narrative, composition, native editing, review and delivery.
- [Repository PowerPoint skill](../skills/pptxgengo/SKILL.md): select the current CLI or a specific source-bound repository route.
- [Project setup and resume](../skills/west-monroe-presentations/references/project-and-resume.md), [folder structure and repair](../skills/west-monroe-presentations/references/project-structure.md), [commands](../skills/west-monroe-presentations/references/cli-reference.md), [editing](../skills/west-monroe-presentations/references/editing-slides.md), [template selection](../skills/west-monroe-presentations/references/template-selection.md), and [available media](../skills/west-monroe-presentations/references/assets.md).

Skill references retain what changes an agent's next authoring action: command
forms, source/working-copy boundaries, bindings, available resources, migration,
fit, review, snapshots and colleague shares. Release history, internal schemas,
CI infrastructure and qualification ledgers belong in the repository docs below.

## Current features and contracts

| Topic | Guide |
| --- | --- |
| Release contents and actual qualification scope | [Release status](release-status.md) and [project changelog](../CHANGELOG.md) |
| Install, upgrade, rollback and recovery | [Installation](installation.md), [Codex/Claude skill setup](skill-installation.md), [schema and template upgrades](../skills/west-monroe-presentations/references/upgrading-projects.md) |
| Folder layout, shared asset objects, numbered snapshots and complete shares | [Portable projects](portable-projects.md) |
| Packaged template deck and deferred reusable content deck | [Browsing libraries](browsing-libraries.md) |
| Native profile conversions and retained variants | [Native editing profile](native-editing-profile.md) |
| Explicit card/list/table components | [Card](native-editable-card.md), [list](native-editable-list.md), [table](native-editable-table.md) |
| Family composition, stable identities and guarded source edits | [Composition operators](composition-operators.md), [teams/pods/governance runbook](../skills/west-monroe-presentations/references/team-composition.md), [Gantt runbook](../skills/west-monroe-presentations/references/gantt-composition.md) |
| Geometry changes and bounded native imports | [Geometry editing](geometry-editing.md), [native imports](../skills/west-monroe-presentations/references/native-imports.md) |
| Native ownership and text adoption | [Lineage](native-lineage.md), [field mapping](native-field-mapping.md), [text reconciliation](text-reconciliation.md) |
| Reviewed native Gantt interval/gate adoption and semantic limits | [Native semantic reconciliation](native-semantic-reconciliation.md), [Gantt runbook](../skills/west-monroe-presentations/references/gantt-composition.md) |
| Variable cycle steps, feedback relationships and active marker | [Cycle contract](cycle-composition-scope.md), [cycle runbook](../skills/west-monroe-presentations/references/cycle-composition.md) |
| Ordinal assessment axes, explicit score domains and missing values | [Assessment contract](assessment-composition-scope.md), [assessment runbook](../skills/west-monroe-presentations/references/assessment-composition.md) |
| Operator-provided maintained slide revisions | [Finished slides](finished-slides.md) |
| Discovery/index and optional offline search | [Template discovery](semantic-template-discovery.md), [performance](search-performance.md) |
| CLI APIs and template engineering | [CLI engineering](engineering-cli.md), [new templates](engineering-new-templates.md) |
| Brand copy rules | [Brand voice](brand-voice.md) |

## Maintainer operations and evidence

- [PowerPoint failure types and evidence](powerpoint-automation-diagnostics.md), [operator recovery](../skills/west-monroe-presentations/references/powerpoint-recovery.md), and [GUI Save As evidence](native-save-as-equivalence.md).
- [Release CI](../release/CI.md), [release packaging](../release/README.md), and [tagging](../RELEASING.md).
- [CI cadence](ci-cadence.md), [GitHub review to private GitLab relay](github-pr-gitlab-poller.md), and [test lanes](testing.md).
- [Native editing pilot](native-editing-pilot.md), [roundtrip harness](native-roundtrip.md), and [component demo](native-component-demo.md) retain exact fixture evidence and outstanding desktop tasks.
- [Legacy repository workflows](legacy-authoring/workflow.md), [legacy contracts](legacy-authoring/library-authoring.md), and [source-template adaptation](legacy-authoring/template-rollout.md) cover older scene/compose tools outside the installed design route.
- [Core-port conventions](../PORTING.md) and [historical port review](../REVIEW.md) document the original OOXML port.

## Historical material

[Skill planning](skill-planning/README.md), [product plan](../PRODUCT_PLAN.md),
[implementation plan](../IMPLEMENTATION_PLAN.md), [engineering waves](engineering-waves.md),
[engineering follow-ups](engineering-followups.md), and the source `planning/`
folder retain decisions and checkpoints. Older GitHub CI run links in evidence
sections are historical records; current jobs execute in private GitLab.

Pinned `library/` documentation, source receipts and benchmark/sample decks stay
with their original bundles. Their counts/qualification describe those inputs,
not every current profile conversion. The upstream JavaScript [changelog](upstream-pptxgenjs-changelog.md)
and inherited Node demos are also separate from the Go release process.

## Current development and remaining scope

[Geometry editing](geometry-editing.md) describes source editing, tagged native
geometry reconciliation and bounded desktop qualification on main.
[Composition operators](composition-operators.md) covers the team, pod, governance,
Gantt and bounded native-import additions. Inspect installed help and toolchain
pins: these additions do not retroactively change the promoted release.

[Remaining geometry scope](geometry-editing-scope.md), [routing backlog](geometry-routing-backlog.md),
and the [semantic family audit](template-customization-audit.md) retain gaps.
Arbitrary native imports, automatic obstacle routing and additional family
operators remain separate work; successful specimens do not qualify every variant.
