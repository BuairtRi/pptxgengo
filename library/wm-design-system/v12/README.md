# V12 template library — v4.3.1

Pinned upstream: `36132d5637abdbdeb70945ad650b795cabea05ef` (2026-10-10).
735 templates, 734 active, one deprecated. All V11 identities remain available;
86 additions and two revisions expand the frozen library. The earlier 677-template
V12 intake is preserved in planning as a historical source pin.

V12 is the default for new projects and release packaging. Existing projects
retain their declared source pin. All source specimens and active bound stock
content build with native-v1. The published specimen gallery qualifies 88 freshly
reviewed PowerPoint pages and 647 retained previews with exact composition and
visible rendered dependency checks. Arbitrary supplied copy, other densities,
Windows PowerPoint and native editing round trips need separate qualification.

- [Native specimen gallery](catalog/design-system.html)
- [Frozen intake and qualification](../../../planning/wm-design-contracts/v12/intake-20261010-735-frozen/README.md)
- [Discovery index](library.sqlite)

```sh
pptxgengo design library-search --bundle v12 --query 'milestone timeline'
pptxgengo design library-reference --bundle v12 --template-keys toc/levels-pages --year 2026 --out /tmp/v12-toc
```

Source, fonts, assets and typography are physical files, allowing release
packaging without symlink dependencies. Both manifests and all consumed source
files retain verified SHA-256 identities. Gallery acceptance is recorded beside
the frozen bundle; status labels inside immutable intake manifests describe the
state when those source bytes were pinned.
