# Upgrade schemas, templates and project pins

Upgrade only the part the operator requested. A skill documentation update does
not change a deck. Read `pptxgengo --version`, `pptxgengo paths`, the project's
`deck.yaml` and `toolchain.lock.json` before choosing a target.

## Know which version is changing

| Version | Meaning | Upgrade route |
| --- | --- | --- |
| CLI release, such as v4.2.1 | Executable identity and available commands | Install the selected authenticated release, then explicitly migrate existing project pins. |
| Engine, such as `wmds-go-foundation.v2` | Compiler behavior | `project migrate --engine ENGINE`; omission retains the existing engine. |
| Bundle, such as `v11` | Shared templates, components, fonts and source contracts | `project migrate --bundle REVISION_OR_PATH`; choose the target explicitly. |
| Shared template ID and optional `revision` | A slide's chosen layout and contract | Inspect the target template, review bindings and deliberately swap or update the slide reference. |
| Deck schema, currently `pptxgengo.deck-document.v1` | YAML document format | Follow an actual schema converter or translate a preserved copy against the supported source format. `project migrate` does not convert schemas. |
| `editing_profile`, such as `native-v1` | Native object structure in generated PowerPoint | Deliberately edit `deck.yaml`, rebuild and review. Migration does not enable it in existing decks. |
| `versions/000001` | A complete authored deck snapshot | Save or materialize snapshots; these numbers are independent of template and schema versions. |

The v2 engine still uses the v1 deck schema. Never change a schema suffix to
match the engine or CLI version. Folder migration with `project layout` is also
separate from schema and toolchain migration.

## Preserve the predecessor

With the existing compatible CLI, check and build the current source, then save
its reviewed state with `project version save`. Preserve native edited working
copies separately. Retain the old runtime and full project, including shared
asset objects, local templates, locks and history. When the source cannot build,
copy the complete project before repair; do not delete its lock to bypass drift.
Use a working copy for the upgrade and avoid concurrent OneDrive edits.

## Migrate the toolchain deliberately

After installing the target CLI, choose the bundle and retain or explicitly
select the engine. For a deliberate upgrade to V11:

```sh
pptxgengo design project migrate --project ./client-deck --bundle v11 --dry-run
pptxgengo design project migrate --project ./client-deck --bundle v11
pptxgengo design project check --project ./client-deck
pptxgengo design project build --project ./client-deck
```

For a CLI-only upgrade, use the project's existing bundle in both commands.
Omitting `--bundle` selects the current default, V11, even for an older project.
The chosen library must actually be installed or supplied by a verified path.

Migration validates source compatibility, template references and compiler fit
before atomically replacing the lock. Dry runs and validation failures preserve
the lock. Success retains exact old bytes in
`<configured-lock>.pre-migrate-<hash>`; an already current project is unchanged.
It does not rewrite copy, schema, local definitions, template revision pins,
editing profile, asset revisions or existing PowerPoint files.

For a shared-template change, inspect the target with `library-inspect` and
`library-authoring`. `project swap --slide ID --template KEY` produces a proposal;
review mapped copy and required slots before `--apply`. Keep a revision pin until
a reviewed replacement is chosen; do not remove it simply to silence an error.
Derived local templates can retain ancestry tied to the old contract: compare
that ancestry and source definitions, then deliberately rebase or retain the
matching bundle. Migration does not automatically merge upstream changes into a
local derivative.

## Schema changes and unsupported input

There is no general deck-schema upgrade command in the current release.
`project migrate` must be able to load the existing document first. If loading
fails on its schema, retain the compatible CLI and original files. Use a
specific documented converter when one is available; otherwise translate a
copy against [source format](source-format.md), preserving IDs, order, copy,
bindings, notes, assets and local definitions. Validate the translated copy with
`project check` and build it before replacing the working source. A standalone
PowerPoint is not converted into YAML by migration.

## Review and rollback

Measure the new build, render it through PowerPoint and inspect all affected
slides. Old renders and native acceptance do not qualify a changed build;
record real decisions through `project attach-render` and the normal review
workflow. Keep old and new baselines separate, then save a numbered version
with a message naming the CLI, bundle and template changes.

To return to a prior project, use `project version materialize --number NUMBER
--out NEW_DIR` with a compatible CLI and select the matching retained runtime.
`installation rollback` changes the active runtime; it does not restore deck
source. An old lock backup is likewise useful only with its matching source,
bundle and executable. Keep the failed upgrade copy for comparison.

A colleague share retains executable, OS and architecture pins. On another
platform, migrate the extracted working copy explicitly and review its rebuilt
deck; sharing does not waive font or media requirements. See
[project setup and snapshots](project-and-resume.md).
