# West Monroe presentation library

The maintained design library is [wm-design-system/v5](wm-design-system/v5/README.md):
587 templates, 586 active, with accepted native PowerPoint source previews.
Previous source revisions and temporary typography candidates are retired.

- Browse `wm-design-system/v5/catalog/design-system.html`.
- Query `wm-design-system/v5/library.sqlite`.
- Author against `wm-design-system/v5` with the Go CLI.

```sh
pptxdesign library-find --index library/wm-design-system/v5/library.sqlite \
  --query 'phased delivery roadmap' --kinds template --summary
pptxdesign library-inspect --index library/wm-design-system/v5/library.sqlite \
  --id lifecycle/three-phases --summary
```

Run from the repository root or use absolute paths. Discovery exposes template
purpose, topology, named slots, content zones and verified screenshot paths.
Select layouts that express the supplied content's meaning; template counts or
successful compilation do not prove that new content fits.

The pinned bundle owns source, fonts and registered assets. The catalog and
SQLite are discovery resources alongside that bundle. Catalog artifact paths are
relative to its root; the index records its resolved root and checks every hash.
See [semantic discovery](../docs/semantic-template-discovery.md) and
[Go authoring](../cmd/pptxdesign/README.md).
