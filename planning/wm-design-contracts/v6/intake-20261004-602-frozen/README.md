# 602-template intake candidate

Pinned upstream `c355c881d543dceccb82dbe12d7129bca9b1fcac`: 15 added templates and one revised heat map. This candidate has not passed native PowerPoint review. The installed production library remains v5 (587 definitions).

`bundle/` pins all 38 source files. It stores seven changed files and uses relative
links to the 31 unchanged v5 files and existing fonts, logos and calibration.
Every dependency is still hash verified. Keep the candidate inside this repository;
it requires the pinned v5 bundle. There is no duplicate production gallery, SQLite
index or font set. `delta.json` lists the intake.

`native-review.pptx` contains the 16 changed source specimens in alphabetical
template order. It is editable and passes Go layout checks. Native CLI export
failed with Apple-event dispatch `-10827`; this file is **not visually accepted**.
`qualification.json` records the completed checks and remaining native gate.

After desktop dispatch works, use the current CLI with a new output directory:

```sh
pptxgengo design render --pptx planning/wm-design-contracts/v6/intake-20261004-602-frozen/native-review.pptx --out /tmp/wmds-602-native-review --pdf --png --contact-sheet
```

Review all 16 pages for group-rail centering, note-row margins, heat color
endpoints, priority-chip labels/keys, reference badge centering, detail locator
outlines, split-column clearance and continuation labels. Review original-size
slide images as well as the contact sheet. Only then promote the candidate and
update the production gallery/index and global release together.
