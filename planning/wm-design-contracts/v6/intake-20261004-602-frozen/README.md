# 602-template intake candidate

Pinned upstream `c355c881d543dceccb82dbe12d7129bca9b1fcac`: 15 added templates and one revised heat map. All 16 changed source specimens passed individual native PowerPoint visual review on 2026-10-05. The installed production library remains v5 (587 definitions).

`bundle/` pins all 38 source files. It stores seven changed files and uses relative
links to the 31 unchanged v5 files and existing fonts, logos and calibration.
Every dependency is still hash verified. Keep the candidate inside this repository;
it requires the pinned v5 bundle. There is no duplicate production gallery, SQLite
index or font set. `delta.json` lists the intake.

`native-review.pptx` contains the 16 changed source specimens in alphabetical
template order. It is editable and passes Go layout checks. Its final SHA256 is
`9b5ab2e0a675ba734f8e77279d521a77145c8eae73b6927d62201de7cf652376`.
`native-review/` retains the local PowerPoint PDF, all 16 original-size PNGs and
an unsigned review manifest. [native-review.md](native-review.md) records the
page-by-page acceptance and repairs.

The accepted PDF was exported through PowerPoint **File → Export → PDF → Best
for printing**, using the local option. The existing PDFKit renderer rasterized
it to 1920×1080 PNGs. Fresh automated CLI export of the repaired deck remained
blocked by Apple-event dispatch `-10827` under the current restricted caller.
An earlier automated export succeeded for the preceding deck, before the native
badge, priority and header width repairs; its signed manifest is retained as
historical evidence. The final GUI review does not qualify automated export or
the PowerPoint doctor. `qualification.json` distinguishes these results.

After desktop dispatch works, use the current CLI with a new output directory:

```sh
pptxgengo design render --pptx planning/wm-design-contracts/v6/intake-20261004-602-frozen/native-review.pptx --out /tmp/wmds-602-native-review --pdf --png --contact-sheet
```

The 16-page visual review checked group-rail centering, note-row margins, heat
color endpoints, priority-chip labels, reference badge centering, detail locator
outlines, split-column clearance and continuation labels. Gallery/index and
default-library promotion remain a separate release decision; this evidence
update does not change the production v5 bundle or source pins.
