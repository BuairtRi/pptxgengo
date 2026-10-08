# Documentation

Use the current operator guides for deck work. Dated plans, qualification ledgers
and inherited upstream docs describe the exact checkpoint named in each file;
they do not override the installed CLI or a project's lock.

## Agent operator guidance

- [West Monroe presentation skill](../skills/west-monroe-presentations/SKILL.md): intake, voice, narrative, composition, native editing, review and delivery.
- [Repository PowerPoint skill](../skills/pptxgengo/SKILL.md): select the current CLI or a specific source-bound repository route.
- [Project setup and resume](../skills/west-monroe-presentations/references/project-and-resume.md), [commands](../skills/west-monroe-presentations/references/cli-reference.md), [editing](../skills/west-monroe-presentations/references/editing-slides.md), [template selection](../skills/west-monroe-presentations/references/template-selection.md), and [available media](../skills/west-monroe-presentations/references/assets.md).

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
| Native ownership and text adoption | [Lineage](native-lineage.md), [field mapping](native-field-mapping.md), [text reconciliation](text-reconciliation.md) |
| Operator-provided maintained slide revisions | [Finished slides](finished-slides.md) |
| Discovery/index and optional offline search | [Template discovery](semantic-template-discovery.md), [performance](search-performance.md) |
| CLI APIs and template engineering | [CLI engineering](engineering-cli.md), [new templates](engineering-new-templates.md) |
| Brand copy rules | [Brand voice](brand-voice.md) |

## Maintainer operations and evidence

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
