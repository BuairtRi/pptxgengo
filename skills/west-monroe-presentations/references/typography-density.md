# Typography density

Use these controls with a V11 source lock. Existing projects retain their pinned
source; earlier sources without density tokens reject the new options. Inspect
the installed template and preview before adjusting density.

## Edit one slide

Put these fields beside `template` and `content` in the slide YAML:

```yaml
density: compact
header_density: comfortable
auto_density: false
```

`density` applies one tier to the entire slide body, including cards, nested
lists, charts and table text. Corresponding roles remain consistent across the
slide. Choose `comfortable`, `compact` or `dense`; an omitted body tier uses
the template's authored tier. Headers default to Comfortable independently.
Use `header_density` explicitly when a two-line title needs smaller type.

Automatic fitting is enabled by default for V11. It starts at the requested
body tier and can step down to Compact and Dense. It does not rewrite the YAML,
so reducing copy can restore the requested tier on the next build. Set
`auto_density: false` to keep the requested tier and receive a fit error instead.
Review every automatic change reported on stderr and in the layout report,
especially on introductions, overviews and other slides intended to be light.

## Designer role scales

Sizes below are points; leading also follows the pinned source tokens.

| Role | Comfortable | Compact | Dense |
| --- | ---: | ---: | ---: |
| Standard action title | 32 | 28 | 24 |
| Card title (`subhead`) | 18 | 16 | 14 |
| Body | 14 | 12 | 11 |
| Small/card copy | 12 | 11 | 10 |
| Small table cell | 12 | 10 | 8 |

The normal reading floor is 8 pt. Fixed circle and numeral-tile text retain
their defined sizes; source/legal typography retains its own rules. The
designer defines two smaller utility exceptions: Gantt period sublabels at
7.5 pt and Draft Review Note status chips at 6.5 pt. Outer boxes, padding,
table row heights and title/body boundaries retain their authored geometry.
Density is a typography control; choose another layout when geometry needs to change.

Ten agenda/schedule and lead-question templates allow Comfortable only.
Their source-owned `densityLimit` cannot be relaxed in project YAML.
`density.prohibited_tier` rejects an explicit unsupported tier;
`density.limit_exhausted` means automatic fitting reached the template limit.
Shorten the copy, split the slide or select another stock template.

## Build and inspect

```sh
pptxgengo design project build --project /path/to/deck
pptxgengo design measure-style --style small --density dense --scope cell \
  --text 'A longer table value' --width 180
```

`measure-style` reports shaped text and estimated height, not template fit or
native acceptance. Build the edited deck and use native PowerPoint review to
check wrapping, spacing and reading quality. The qualified gallery uses stock
copy; its acceptance does not transfer automatically to replacement copy.
