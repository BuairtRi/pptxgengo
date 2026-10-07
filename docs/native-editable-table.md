# Directly selectable native tables

`wmds/component/editable-table` uses the existing table planner and removes its
outer group when the result is exactly one native PowerPoint table. Cell content,
widths, native styling, density and generation identity are unchanged.

Use the same arguments and authored row keys as `wmds/component/table`.
A scene can use `type: editable-table`. It rejects any table requiring external
objects, including heat underlays, continued labels, reference/priority badges,
highlight decoration or legends. Keep those tables on the existing renderer
until their coordinated editing design is qualified.

The comparison fixture checks all three densities and inspects the generated
PPTX: each candidate is one top-level `graphicFrame`, without a group. The real
Mac task selected the Compact table directly and changed Review to Reviewed,
preserving its five other cells and generation identities. This does not qualify
column resizing, Windows behavior or automatic migration of stock templates.
See [component demo](native-component-demo.md).
