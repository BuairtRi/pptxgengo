# WMDS foundation bundle v1

Frozen adapter input for `wmds-native.v1`, reviewed on 2026-10-01.

- `source/`: 21 byte-identical source files from `wm-design-system`; original
  schemas and all metadata are retained. Source code is provenance, not executed.
- `inventory.json`: original inventory and file hashes. It records source
  inspection, not runtime support or qualification.
- `fonts/`: unmodified IBM Plex Sans and Mono static Regular, Medium, Semibold
  and Bold, copied from the existing WM brand bundle, with OFL licenses and hashes.
  Fonts are packaged for measurement; this prototype does not embed them in PPTX.
  Native rendering still needs matching fonts installed on the recipient machine.
- `assets/`: original WM positive/reverse horizontal SVG logos and PNG fallbacks.
- `bundle.json`: pinned font/asset hashes; any change requires adapter migration.

Implementation, usage, current capabilities and qualification limits are described
in [the command documentation](../../../cmd/pptxdesign/README.md). The normative
[adapter contracts](../../../planning/wm-design-contracts/v1/README.md) also cover
future component/template slices; their existence does not imply implementation.

Generation is offline Go. PowerPoint export and visual review are separate steps.
This bundle and prototype have not been promoted into the installed local release.
