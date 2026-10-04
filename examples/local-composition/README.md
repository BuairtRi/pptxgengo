# Local composition example

Three synthetic slides demonstrate maintained local templates using the v5
component adapters:

| Slide | Component |
| --- | --- |
| `align-three-lenses` | `wmds/component/venn` |
| `assess-capability` | `wmds/component/maturity` |
| `sequence-three-gates` | `wmds/component/road` |

Copy this example into a new project. Keep generated locks, state and builds out
of the example source directory. Run from the repository root:

```sh
go build -o /tmp/pptxdesign ./cmd/pptxdesign
cp -R examples/local-composition /tmp/local-composition
/tmp/pptxdesign project init --project /tmp/local-composition --bundle library/wm-design-system/v5
/tmp/pptxdesign project check --project /tmp/local-composition --bundle library/wm-design-system/v5
/tmp/pptxdesign project build --project /tmp/local-composition --bundle library/wm-design-system/v5
```

All copy, labels, numbers, synthetic dates and diagram states live in
`slides[].values`. Definitions retain allocation, style, topology and binding
rules. Venn region indices and maturity active positions are zero-based; road
`at` and `side` fields place the authored milestones. No real client evidence is
implied. The example has no native acceptance receipt: inspect every generated
page before using it, and review changed text and geometry independently.
