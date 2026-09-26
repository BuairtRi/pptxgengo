# Numbered and metric cards

`cards.json` contains two illustrative slide fixtures for the dynamic composition
API: three vertically stacked numbered explanation rows and three metric value /
label cards. The metric values on the second slide are sample values, not client
results. The first slide's workstream wording is illustrative as well.

## Source inspiration and sizing

The numbered row follows reference N2, software-modernization source slide 8:
horizontal pale surface, thin blue side rule, separate number, title, and body.
The single metric card follows M1, modernization source slide 57: a pale card with
value above label. M3 is a related value / label arrangement whose source notes
explicitly leave the shared navy strip and separators to its parent composition;
this component likewise does not create dividers.

The shortlist and component slices record source identity and observed visual
structure, but mark text capacity and adaptation as unmeasured. The component
catalog gives the source previews as 1920 × 1080 pixel images, which supports the
16:9 page proportion but does not expose native card bounds. The fixture uses the
compose package's 960 × 540 point widescreen coordinate system. Its bounds are
newly chosen sample geometry, not copied source object bounds, and do not claim
source-measured capacity.

## Spec fields

Each `cards[]` entry has a unique `id`, `kind` (`numbered` or `metric`), fixed
`bounds`, and an optional `profile` (`light` or `dark`). A numbered card requires
`number`, `title`, and `body`; a metric card requires `value` and `label`. The
light profile defaults to `surface.light`; dark defaults to navy. `surface` may
override that default. `foreground` defaults to `auto`, which selects a brand
foreground that passes the compose package's 4.5:1 contrast rule; an explicit
`#RRGGBB` is also accepted only when it passes that rule. The numbered side rule
defaults to blue and may use an explicit `accent` color.

Each text field becomes its own editable native text shape. Probe widths preserve
the exact usable width of each planned slot after insets. Planning rejects
measured text that exceeds the fixed slot; it does not shrink type or estimate
line count. Cards cannot overlap one another, the title, pods, roles, phases,
legend, or canvas primitives unless a canvas primitive explicitly names the card
in its `allow_overlap` list. The component owns its card background and numbered
side rule. A parent composition owns shared bands, separators, and dividers.

## Limits

These are measured layout primitives, not source-object transplant contracts.
They support a single number/title/body row or value/label pair per card and use
fixed Arial styles. No alternative typography, rich text, bullets, or per-field
style tuning is exposed. Source-fitting and final native visual review remain
necessary for production copy.
