# Native measurement performance finding

The production adapter, `scripts/measure-compose-text.applescript`, reads each PowerPoint text-range character's bounds and font properties separately. This is costly, but it preserves character-level validation of font name, size, weight, italic, underline, color, and bounds.

On 2026-09-27, a separate bulk-read prototype compiled successfully but failed at runtime on the first footer of the current process-v3 deck. PowerPoint reported `text length = 31`; `properties of every character of tr` returned a character-class aggregate record with a count of 11, and `properties of font of every character of tr` returned a font-class aggregate record with a count of 36. Neither was an ordered list of 31 character property records. The prototype stopped with error 75, `Bulk character/font property count mismatch`, before producing measurements.

The prototype was removed. The character-measurement path is unchanged, and no character validation was relaxed. Bulk property syntax compiling does not establish that PowerPoint supports an indexed bulk result. Any future speed change needs a separate native experiment and exact output comparison before adoption; no further optimization is planned for this milestone.

## Remaining bottleneck

The final comparison bundle has 297 text frames and 9,709 characters; architecture/team has 357 text elements and 6,244 characters. Final native verification currently re-reads the actual built deck rather than treating probe-cache reuse as final evidence. The character-level adapter is consequently the dominant serial step. A future optimization should prove equivalent font, rich-text and character-bound observations on fixed evidence before replacing this path. This milestone retains that character-level measurement path.

## Separate SVG reporting correction

Native arrow verification exposed a distinct correctness issue, unrelated to measurement speed: PowerPoint reports some SVG graphics with line style `single line` and the exact unavailable/mixed width sentinel `-2147483648`. The adapter now preserves raw `shape_type`, line style, visibility and width instead of stopping before the CLI can interpret them. The CLI accepts that anomaly only for an expected SVG image with no outline and only after package validation proves the picture has no `<a:ln>`. Explicit picture outlines, ordinary shapes and connectors retain their existing checks. Because the adapter identity changed, the arrow probe and final deck must be regenerated and remeasured; old evidence is not relabeled as current.
