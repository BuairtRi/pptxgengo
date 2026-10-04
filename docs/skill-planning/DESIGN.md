---
version: "alpha"
name: "West Monroe"
description: "A practical visual system for creating clear, human, energetic West Monroe assets."
colors:
  primary: "#070154"
  white: "#FFFFFF"
  highlight-yellow: "#F6EB21"
  highlight-magenta: "#F900D3"
  highlight-blue: "#0047FF"
  dark-gray: "#50658E"
  medium-gray: "#CED7E6"
  light-gray: "#E8EEF8"
typography:
  display:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 4rem
    fontWeight: 600
    lineHeight: 1
    letterSpacing: -0.03em
  h1:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 3rem
    fontWeight: 600
    lineHeight: 1.05
    letterSpacing: -0.025em
  h2:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 2rem
    fontWeight: 600
    lineHeight: 1.15
    letterSpacing: -0.015em
  h3:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 1.5rem
    fontWeight: 600
    lineHeight: 1.2
  body-lg:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 1.25rem
    fontWeight: 400
    lineHeight: 1.5
  body-md:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 1rem
    fontWeight: 400
    lineHeight: 1.5
  body-sm:
    fontFamily: "IBM Plex Sans, Arial, sans-serif"
    fontSize: 0.875rem
    fontWeight: 400
    lineHeight: 1.5
  eyebrow:
    fontFamily: "IBM Plex Mono, Arial, monospace"
    fontSize: 0.75rem
    fontWeight: 600
    lineHeight: 1.35
    letterSpacing: 0.08em
  label:
    fontFamily: "IBM Plex Mono, Arial, monospace"
    fontSize: 0.875rem
    fontWeight: 500
    lineHeight: 1.35
  infographic-number:
    fontFamily: "IBM Plex Mono, Arial, monospace"
    fontSize: 2.5rem
    fontWeight: 600
    lineHeight: 1
    letterSpacing: -0.02em
  office-body:
    fontFamily: "Arial, sans-serif"
    fontSize: 1rem
    fontWeight: 400
    lineHeight: 1.5
rounded:
  none: 0px
  sm: 0px
  md: 0px
  full: 9999px
spacing:
  none: 0px
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 24px
  xl: 32px
  2xl: 48px
  3xl: 64px
  4xl: 96px
components:
  body-copy:
    backgroundColor: "{colors.white}"
    textColor: "{colors.primary}"
    typography: "{typography.body-md}"
  secondary-copy:
    backgroundColor: "{colors.white}"
    textColor: "{colors.dark-gray}"
    typography: "{typography.body-sm}"
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.white}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: 16px
  button-primary-hover:
    backgroundColor: "{colors.highlight-magenta}"
    textColor: "{colors.primary}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: 16px
  button-secondary:
    backgroundColor: "{colors.white}"
    textColor: "{colors.primary}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: 16px
  text-link:
    backgroundColor: "{colors.white}"
    textColor: "{colors.highlight-blue}"
    typography: "{typography.body-md}"
  headline-highlight:
    backgroundColor: "{colors.highlight-yellow}"
    textColor: "{colors.primary}"
    typography: "{typography.h1}"
    rounded: "{rounded.none}"
  surface-light:
    backgroundColor: "{colors.light-gray}"
    textColor: "{colors.primary}"
    rounded: "{rounded.none}"
    padding: 24px
  surface-medium:
    backgroundColor: "{colors.medium-gray}"
    textColor: "{colors.primary}"
    rounded: "{rounded.none}"
    padding: 24px
  icon:
    textColor: "{colors.primary}"
    size: 42px
    height: 42px
    width: 42px
  logo-digital:
    textColor: "{colors.primary}"
    height: 24px
---

## Overview

West Monroe is a business and technology consultancy and a team of value accelerators. The visual identity should feel like **intelligence in action**: clear, pragmatic and confident; collaborative and human; energetic enough to create momentum. It should connect business and technology without becoming sterile, futuristic or abstract for its own sake.

Build layouts from four recognizable devices: the whiteboard grid, square content frames, a restrained yellow headline highlight and supplemental hand-drawn marks. The system is structured but not rigid. Use the grid and squares to organize evidence; use imperfect marks and people-centered imagery to show co-creation.

The YAML tokens are implementation-ready defaults for digital assets. Brand colors and font assignments come directly from the supplied guidance. Type sizes, spacing and zero-radius UI defaults create a reusable hierarchy where the source material does not prescribe exact digital values; scale them proportionally for the medium rather than treating them as print specifications.

## Colors

White is the primary canvas. Grounded Blue (`#070154`) carries headlines, body copy, outlines, the logo and dark backgrounds. Prefer generous white space with decisive areas of Grounded Blue rather than distributing every brand color evenly.

- **Highlight Yellow (`#F6EB21`)** is reserved for the marker-style highlight behind one to four words in a main headline. Do not use it for buttons, fills, decoration or data.
- **Highlight Magenta (`#F900D3`)** adds high-energy emphasis to large callouts, hover states and illustrations. Grounded Blue on magenta passes AA for normal text. White on magenta is limited to large text—at least 18pt regular or 14pt bold—because its contrast is approximately 3.5:1.
- **Highlight Blue (`#0047FF`)** emphasizes small text such as body-copy headers and links, and may appear in illustrations. It is not a general background color.
- **Dark Gray (`#50658E`)**, **Medium Gray (`#CED7E6`)** and **Light Gray (`#E8EEF8`)** organize dense content. Use Light Gray or Medium Gray for quiet surfaces and the whiteboard grid.

Use only approved text/background combinations. Default to Grounded Blue on White or White on Grounded Blue. Do not rely on color alone to communicate state or meaning.

For print, use the official CMYK/Pantone formulas in the brand guidance rather than converting these hex values ad hoc.

## Typography

IBM Plex Sans is the voice of designed headlines, subheads and body copy. IBM Plex Mono is the technical counterpoint for eyebrows, infographic numbers, captions, labels and footers. This pairing expresses the brand's fluency in business and technology.

- Use sentence case for most copy. Reserve all caps for short Mono eyebrows or functional labels.
- Lead with a strong, compact headline; follow with plain-language support. Keep line lengths readable—about 45–75 characters for body copy.
- Use weight and scale before adding color. Avoid a page full of competing type treatments.
- Keep Mono concise. It is a signal for data and metadata, not a substitute for long-form text.
- In Word, PowerPoint, Excel and email, use Arial for compatibility. Preserve the hierarchy and color logic even when the brand fonts are unavailable.
- The supplied sizes form a web-oriented scale. For slides, documents, social assets or large-format work, preserve the relationships while adapting to viewing distance and format.

## Layout

Use a square-based composition with visible alignment. Start on a white canvas, establish a strong content edge, then add one dominant point of emphasis. Asymmetry is welcome when the alignment remains clear.

The 8px-derived spacing scale supports implementation; use larger jumps between sections than within components. Favor whitespace over boxes. When a container is necessary, use a square-cornered Light Gray or Medium Gray field.

### Whiteboard grid

The whiteboard grid is the backbone of the system and represents an open space for ideas and solutions.

- Use the standard square 15×15 grid in Light Gray or Medium Gray on White.
- Scale it from 100% to 200% only, in 25% increments. Tile it if more area is needed, always retaining a square ratio.
- Anchor it to the left or right margin and protect the logo's clear space.
- Align content to grid-cell edges. At least one edge of an overlaid element must touch or bleed beyond an outer grid edge.
- Keep at least three rows and three columns visible.
- Copy may sit over the grid when contrast and readability remain strong.
- Do not crop, distort, recolor or rebuild the grid as a non-square pattern.

### Squares

Squares represent individual data points and frame photography, data imagery, color or important type. Use one to three purposeful squares on the grid. Make them part of the alignment system, not floating decoration. Prefer true squares; do not turn the motif into a field of generic rounded cards.

## Elevation & Depth

Keep the system predominantly flat. Hierarchy comes from scale, color blocks, overlap, grid alignment and whitespace—not gloss or simulated depth.

- Do not add gradients, bevels or drop shadows to logos, icons or illustrations.
- Avoid generic card shadows. If an interactive interface requires elevation, keep it subtle and functional, and never let it compete with the square/grid structure.
- Use overlap deliberately: a photograph or content square may cross the grid's outer edge while remaining aligned to a cell.

## Shapes

Squares and cubes are the core geometry. Corners are square by default in layouts and controls. Pill shapes are not a brand motif; reserve fully rounded geometry for inherently circular controls, avatars or required status indicators.

Hand-drawn shapes are the human counterpoint. They should show slight irregularity, round line caps and joins, and natural variation. Avoid perfect circles and duplicated identical marks when drawing an illustration. Abstract imagery should use dimensional cubes or cube-like particles—not soft organic blobs or flat, generic tech patterns.

## Components

The component tokens define safe starting states, not an exhaustive product UI library.

- **Primary action:** Grounded Blue with White text; on hover, Highlight Magenta with Grounded Blue text. Keep the silhouette rectangular.
- **Secondary action:** White with Grounded Blue text and a visible Grounded Blue boundary in implementation.
- **Links and small emphasis:** Highlight Blue on White, with a non-color cue such as an underline where context requires it.
- **Callouts:** Use Magenta sparingly. Grounded Blue is the accessible default text color; White is only for large text.
- **Muted surfaces:** Light Gray is the quieter choice; Medium Gray creates stronger separation. Keep copy Grounded Blue.
- **Icons:** Draw on a 42×42 grid with 2pt line art. Keep forms simple and free-standing, without modeling, gradients or shadows. Approved icon colors are Grounded Blue, Highlight Magenta and White, chosen for an approved contrasting background.
- **Logo:** Place approved artwork rather than recreating the mark in code or type. The minimum symbol height is 24px on screen and 12pt in print.

When implementing focus, disabled, error or success states not defined here, prioritize WCAG accessibility and familiar interaction behavior. Do not repurpose Highlight Yellow or introduce an unapproved brand color merely to complete a state set.

## Do's and Don'ts

### Do

- Start with the point and build one unmistakable hierarchy.
- Use White and Grounded Blue as the dominant pair.
- Combine crisp structure with one human, hand-drawn or photographic touch.
- Highlight only one to four words in the main headline, using the approved yellow marker graphic on White.
- Use real, bright, naturally lit photography that shows people in meaningful action.
- Keep graphics aligned to the square grid and preserve generous logo space.
- Use active, specific language that connects technology to business and human outcomes.
- Check contrast, legibility, crop quality and asset licensing before export.

### Don't

- Do not turn every brand color into an equal-weight rainbow palette.
- Do not use Highlight Yellow anywhere except the approved headline highlight treatment.
- Do not use another color for that highlight or place the highlight on a colored background.
- Do not use data-visualization colors as general design colors.
- Do not distort, recolor, rotate, separate or add effects to the logo.
- Do not crowd the logo or place it over busy, low-contrast photography.
- Do not use the shorthand logo as a decorative graphic or a replacement for the primary logo outside approved app/social contexts.
- Do not use detailed, shaded, gradient-filled or three-dimensional icons.
- Do not create non-square whiteboard grids, show fewer than three rows/columns or float elements entirely inside the grid.
- Do not default to jargon, inflated claims, passive language or generic “future of technology” imagery.

## Logo

Use the supplied West Monroe logo artwork. The preferred configuration is one-color Grounded Blue on White. Grounded Blue on Light Gray or Medium Gray is also approved. Limited-use photographic applications require a simple, quiet image: Grounded Blue on a light image or White on a dark image.

Place the primary logo in any corner, inside margins that honor the clear-space artwork. Never change the relationship between the compass symbol and wordmark. Use the compass alone only in approved social avatars, app icons and favicons. Contact the brand team before supplying the logo to a third party.

For “Clear Action. Lasting Impact.” use the official logo/tagline artwork so spacing, color, typography and graphic details remain intact. Use White or dark backgrounds—not gray—and do not reduce the tagline below roughly 8pt. If the tagline artwork is the primary content, do not add another competing highlight or hand-drawn emphasis. If the tagline appears alone, place the West Monroe logo elsewhere in the layout.

## Graphic Elements

### Highlight graphic

Use the official marker-style asset, centered horizontally and vertically behind the chosen phrase. Determine its width and height from the cap height of the type and scale the graphic rather than stretching it. Use it sparingly on covers, hero pages or pivotal content—not on every page, subhead or body paragraph. On colored backgrounds, use a hand-drawn accent instead.

### Hand-drawn graphics and illustration

Hand-drawn marks signify human interaction, add fluidity and direct attention. They supplement the message; they should never become visual noise.

For illustration artwork built around a 6×6-inch source size:

- Use Grounded Blue strokes at 0.7pt with round caps and joins.
- Use the custom Highlight Yellow brush at roughly 2–5pt for highlighted areas, adjusted to scale.
- Draw traditional arrowheads manually with slight imperfection; avoid stock arrowhead presets.
- Enable scale-strokes-and-effects when resizing.
- Vary small diamonds and other shapes; do not copy/paste identical forms.
- Keep irregularity intentional and controlled. The result should feel drawn by a capable collaborator, not messy or childish.

## Photography and Imagery

Choose imagery from three complementary modes:

1. **Industry-focused:** real people taking meaningful action in environments relevant to client industries.
2. **Provocative:** unexpected angles and compositions for brand-forward moments.
3. **Abstract:** dimensional cubes in active settings, acting as particles that suggest teamwork, collaboration, insight and outcomes.

For real photography, prefer natural light, neutral tones, bright/clear tonality and authentic human behavior. Avoid staged handshakes, isolated devices, dark corporate clichés and images that portray technology without a human or business consequence.

For abstract imagery, make cubes central, dimensional and active. Avoid organic rounded shapes, colorless scenes and flat two-dimensional compositions. Integrate abstract work with real photography or the broader graphic system rather than letting it become a separate visual language.

## Data Visualization

The extended data palette in the official guidance is restricted to charts, graphs and color-coding. Do not sample colors from the reference image; use the approved source assets or specifications when exact values are available.

- Build charts on White whenever possible.
- Outline every colored data mark in Grounded Blue with at least a 1pt stroke.
- Use Grounded Blue text except over Grounded Blue, where text must be White.
- Use IBM Plex Sans and IBM Plex Mono according to the type roles above.
- Encode critical distinctions with shape, size, labels or patterns as well as color.
- Prefer direct descriptive labels and leader lines over a detached legend.
- Make leader lines Grounded Blue. Where they cross content, center a 1px black stroke over a 4px White stroke for separation.
- Do not separate adjacent bars or pie segments merely for decoration.

## Voice and Messaging

Write with intelligence in action. The voice is passionately pragmatic, kindred and creative, and fiercely energized. It is confident without posturing, technical without becoming mechanical, and optimistic without making vague promises.

1. **Start with the point.** Lead with what matters to the audience. Use plain, precise words and cut throat-clearing.
2. **Be bilingual.** Connect industry understanding and technology to a concrete outcome. Technology is the means, not the story by itself.
3. **Use “we” language.** Speak with people from the same table, showing co-creation and shared ownership.
4. **Generate momentum.** Favor active verbs, specifics, measurable value and next actions.

Useful message anchors include “At West Monroe, we work with you,” “Own the now. Win the future,” and “Clear Action. Lasting Impact.” Treat them as approved strategic language, not filler to repeat on every asset.

## Asset Creation Checklist

Before delivering an asset, verify:

- The purpose, audience and single most important message are obvious.
- White and Grounded Blue dominate; accent colors have assigned roles.
- Fonts, hierarchy and fallback behavior match the medium.
- Grid, squares and hand-drawn elements reinforce the message rather than decorate it.
- Logo artwork, minimum size, clear space and background are approved.
- Photography feels natural, human, bright and relevant.
- Interactive states and text/background pairs meet WCAG needs.
- The voice starts with the point, connects business and technology and ends with momentum.
- Exported dimensions, bleed, resolution, color mode and file format fit the destination.

## Sources and Maintenance

This file consolidates the local `guidance/` materials supplied with this workspace and follows the alpha [DESIGN.md format](https://github.com/google-labs-code/design.md). The local brand guidance is authoritative when it conflicts with implementation defaults here. Preserve the YAML schema and canonical section order when editing. Revalidate token references and contrast after any change.
