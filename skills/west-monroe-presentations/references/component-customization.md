# Component and template customization

Read this when a request combines source-template edits with specific palette,
component, gauge, or accent changes. Start with the exact installed contract or
recipe; the same visual label can describe a different editable structure.

## Keep source designs intact

Use `template` or `component` when the selected slide's existing layout supports
the message. Replace only declared text slots, style roles, and empty zones.
These operations keep the source slide's positions, order, objects, chart data,
images, and artwork unless the selected contract explicitly identifies a safe
edit. They do not reflow a source slide or turn it into a new dynamic layout.

For `template`, inspect the exact ID and copy its complete example values before
editing. Preserve each slot's exact run/binding order and count. A `profile` is
available only when that contract lists it; profiles recolor only their declared
bindings. For `component`, use only the roles and profiles in that exact source
hash-bound contract. A profile called `source` means retain the selected source
tokens. Neither route offers a universal palette switch or unrestricted color
replacement. Review any retained labels, ratings, logos, facts, and data marks
for relevance after changing surrounding text.

Choose `compose` when the request requires a changed hierarchy, new content
counts, or independently controlled components that the selected source does
not expose. This creates a new editable composition. It does not claim to
reproduce the source slide's layout or appearance.

## Color roles

For new component compositions, the common WM palette is:

| Role | Typical color | Use |
| --- | --- | --- |
| Neutral surface | Gray `#E8EEF8` or `#F4F6FA` | Component background and alternating rows |
| Primary ink | Navy `#070154` | Titles, labels, and high-contrast marks |
| Emphasis | Blue `#0047FF` | Active state, key border, or selected emphasis |
| Accent | Pink `#F900D3` | A small, intentional callout where the meaning is clear |

Use the exact color-role names and ranges accepted by the chosen recipe. New
adaptive styles accept `neutral`, `subtle`, and `inverse` profiles and declared
semantic color overrides. Other components may have a narrower role list or only
contract-owned source profiles. Check contrast on the actual surface for each
text and mark; a palette entry alone is not a contrast guarantee.

Keep source colors where they encode a chart series, status, role, target range,
or other evidence. Do not globally replace every occurrence of gray, navy, blue,
or pink. Pink in the team component is specifically the client part-time token;
use it for that meaning only when the legend is present. If pink is used as a
decorative accent elsewhere, keep it distinct from data/status encoding.

## Gauges: separate data channels

A gauge can show a highlighted range or cell plus a pointer. Their meanings come
from the source legend or explicit composition specification:

- The **highlight** marks the selected cell/range defined by that source or
  specification.
- The **pointer** marks a selected cell/value; it may coincide with the highlight
  or be placed independently when the contract allows it.

Preserve the source legend and supplied data meaning. Never move a pointer to
create a desired impression or assign target/actual semantics that the source
does not specify.

There are two materially different paths today:

- The T045 source-bound comparison template retains its five native semicircular
  gauges and offers a separate `pptxgengo template apply-gauge` operation. Per
  gauge, it accepts a list of
  highlighted cells and a pointer cell. By default, one highlighted cell also
  selects the pointer cell; an explicit `pointer_cell` can place the pointer
  elsewhere. Highlight color can be gray, navy, blue, or pink; the five cells and
  pointer remain source freeforms. The control is discrete, using
  the five source pointer positions; it does not interpolate a numeric value or
  change the two scale bands above the table. The ordinary text contract can edit
  scale labels, but changing those labels does not calculate the cell selection
  or pointer position. Keep those values explicitly consistent with the facts.
- The semantic comparison composer computes an editable linear or dial gauge
  from caller-supplied `value`, `scale_min`, `scale_max`, `target_min`, and
  `target_max`. It positions the target highlight and pointer separately. The
  target highlight follows `state.active`; the track follows `state.inactive`.
  The pointer uses contrast-resolved ink based on its track/target context, so it
  has no independent `gauge.pointer` color role yet. This is a concrete component
  model gap for continuously scaled gauges when a user asks to choose pointer
  color independently.

For a T045 request, copy the packaged gauge values example and set each row's
`highlight_cells` and `highlight_color`; set `pointer_cell` only when its intended
position differs from the single highlighted cell. Preserve the source legend's
meaning and do not imply continuous numeric interpolation. For a
new comparison composition, keep actual and target geometry tied to the supplied
numbers. A future continuous-gauge style contract should expose distinct
`gauge.track`, `gauge.target`, and `gauge.pointer` roles, with contrast checks
and legend semantics, while keeping value and target geometry independent.

## Accents belong to components

An accent is a local emphasis attached to a specific phrase or component, such
as a measured highlight behind selected title words, an underline under a short
phrase, a numbered-card side rule, or a bounded component border. It does not
replace the slide's full layout, background, or source-template design. Preserve
the surrounding source structure and use the exact measured target. The phrase
accent workflow supports only its declared artwork, targeting, and geometry;
multiline, rotated, ambiguous, or unsupported cases need a deliberate manual
layout or another component pattern. Review the accent in context with the whole
slide.

## Review

Keep source and target renders side by side. Check that the source hierarchy,
data meaning, and retained material remain intact; verify that highlights and
pointers remain distinct and numerically truthful; check text and mark contrast;
and inspect the final PowerPoint render at presentation size. A successful
binding, style change, or fit report does not establish visual acceptance.
