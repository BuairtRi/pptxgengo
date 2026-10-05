# 616-template workshop intake candidate

Pinned upstream `c788cefeb5bb409118ac217adb53216d8156eec3` adds **14 workshop templates**. Go source and editable bound generation pass for all 14; **14/14 pass native PowerPoint visual review**. All 602 earlier definitions and binding APIs are retained. Full builds pass for 616 source templates and 615 active bound templates; one template is deprecated.

This qualified source is now published as the repository default **v7 / 616**, with one production gallery and SQLite index in `library/wm-design-system/v7`. The global CLI is now **local.16 / v7**, installed with `--cli-only` to preserve the presentation skill link. The release commit excludes the other agent's presentation skill and DentalXChange slide edits. Publication verifies all 30 new/changed native specimens and all 586 inherited renderer outputs.

## Final artifacts

- [Native review deck](native-review.pptx): 14 editable workshop slides in alphabetical template order.
- [Review notes](native-review.md) and [qualification](qualification.json): repairs, test results and acceptance scope.
- [Native PDF](native-review/deck.pdf), `native-review/native-pages/` and [unsigned visual review manifest](native-review/review-manifest.json): final native output and hashes.
- `bundle/` and `delta.json`: immutable source pins and upstream changes. No duplicate working decks, full catalog projections, gallery, SQLite index or font set are retained.

`bundle/` pins 39 source files: six changed files stored here and 33 unchanged sources linked to the v6 candidate. Fonts, assets and calibration are inherited. Retain the pinned v6 and v5 dependency bundles. Six named geometry amendments preserve source bytes, copy, font sizes and semantic values; qualification records their exact scope.

## Checks

The final `make test` suite passed in **120.61 seconds**. The exhaustive Go source/bound build passed in **89.106 seconds** and is skipped by the everyday `-short` lane. Workshop round trips, inheritance, amendment atomicity/idempotence and SQLite discovery pass. The temporary SQLite index surfaced workshop agenda/facilitation templates across all 616 entries and was removed after qualification.

Visual acceptance used PowerPoint's local **Export PDF → Best for printing**, then the existing Swift PDF rasterizer. This is not a signed automated CLI render receipt. A fresh automated final export remains unqualified because this restricted caller reports Apple-event dispatch `-10827`; no doctor pass is claimed.

## Reproduce

The existing repository `./pptxdesign` binary was rebuilt from the final source. Installed local.16 includes the v7 implementation. Write new output folders:

```sh
./pptxdesign library-source-reference --bundle planning/wm-design-contracts/v7/intake-20261005-616-frozen/bundle --family workshops --out /tmp/wmds-616-workshops-source
./pptxdesign library-reference --bundle planning/wm-design-contracts/v7/intake-20261005-616-frozen/bundle --family workshops --out /tmp/wmds-616-workshops-bound
./pptxdesign render --pptx planning/wm-design-contracts/v7/intake-20261005-616-frozen/native-review.pptx --out /tmp/wmds-616-workshops-native --pdf --png --contact-sheet
```

The final command requires a caller with working PowerPoint automation. Generation reports deliberately keep native qualification flags false; native visual acceptance is recorded separately.
