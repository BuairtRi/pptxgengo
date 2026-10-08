# Authoring workflow

Use `pptxgengo --version`, `pptxgengo paths` and current command help to resolve
the installed library and available resources. The current archives contain a
V11 source bundle, SQLite index, fonts, specimen gallery and template browsing
deck. Private media originals and the agent skill are separate.

```sh
pptxgengo design library-find --query 'weekly status' --kinds template --summary
pptxgengo design project create --out ./deck --id deck --title 'Working deck' --bundle v11 --template cards/3
pptxgengo design project check --project ./deck
pptxgengo design project build --project ./deck
```

Replace scaffold example copy, register actual assets and keep the composition
log current before building. Review the generated layout report and density
warnings. A source build is an immutable baseline; render it through PowerPoint,
inspect every page and attach real review evidence.

New projects persist native-v1; existing locks and profiles stay unchanged.
Eligible lists use native paragraphs, simple cards use one editable shape and
tables retain native cells. Complex variants can retain multiple objects. Keep
native edits in a separate PowerPoint copy and review supported reconciliation
proposals before adopting text. Combined lists/cards and geometry need manual
source updates.

Use stable `slides/`, `slides/templates/`, shared `assets/objects/sha256/` and
numbered `versions/`. Save versions through `project version save`; transport a
complete private project with `project share`, then `share-extract` and
`share-verify`. Client exports contain only delivery material. Toolchain changes
require deliberate `project migrate`; sharing does not erase OS/compiler pins.

For standalone source-bound scenes or experimental measured compositions,
read the selected route in `docs/legacy-authoring/workflow.md` in the repository.
Those tools are separate from the installed design/project route. Use an existing
PowerPoint-accessible staging folder and preserve unrelated open/unsaved decks.

For schema, engine, bundle and template revision changes, use the West Monroe
skill's project-upgrade reference. Toolchain migration changes pins, not schemas
or local template definitions; select the bundle explicitly and review a fresh
baseline before adopting it.

For project folder repair, read the West Monroe skill's project-structure
reference. `project layout` normalizes active slide/local-template references;
it does not reorganize assets, context or immutable build/version history.
Preserve predecessors, update other authored references deliberately and rebuild.

For architecture customization and the geometry development commands, read
[architecture and geometry](../../west-monroe-presentations/references/architecture-geometry.md).
The promoted v4.2.1 still has text-only adoption; the development build adds
reviewed tagged transforms and existing-object paint order. Native topology and
route edits remain explicit review.
