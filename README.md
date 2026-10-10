# pptxgengo

Go tools for creating, inspecting, adapting and reviewing editable PowerPoint
decks. The Go presentation writer retains PptxGenJS 4.0.1 compatibility goals;
the repository also contains the `pptxgengo` authoring CLI and West Monroe
template system. Current release source is **v4.3.1**. Requires **Go 1.27.1 or newer**;
release builds use Go 1.27.2.

The default template bundle is **v12** (`release/default-bundle.txt`):
735 source templates (734 active), with bundled fonts, a native specimen gallery,
compiler sources and a regenerated SQLite discovery index. The frozen upstream
commit is `36132d5637abdbdeb70945ad650b795cabea05ef`. V12 adds 86 templates
and revises two against V11, including openers, contents/indexes, milestones,
product backlog, strategy roadmaps and resource plans.

All 735 source specimens and 734 active bindings build. Gallery qualification
combines 88 fresh PowerPoint reviews with 647 retained previews verified by
composition and rendered dependency comparison. New projects default to native-v1;
existing project pins remain explicit. Supplied copy needs its own fit/build and
native review. See [release status](docs/release-status.md) and the
[V12 bundle](library/wm-design-system/v12/README.md).

The three final converted decks, editable YAML and offline packages are together
in [`samples/final`](samples/final/README.md).

- [Local CLI release and Codex installation](release/README.md)
- [Documentation index and current versus historical records](docs/README.md)
- [Current release and qualification status](docs/release-status.md)
- [West Monroe authoring skill with progressive references](skills/west-monroe-presentations/SKILL.md)
- [Current capabilities and verified baseline](CAPABILITIES.md)
- [WMDS Go authoring: typography, grids, components and slide designs](cmd/pptxdesign/README.md)
- [Product architecture, experiments, and implementation sequence](PRODUCT_PLAN.md)
- [Current remaining-work plan and library inventory/curation sequence](IMPLEMENTATION_PLAN.md)
- [SQLite catalog, narrative model, composition, and design passes](DESIGN_WORKFLOW.md)
- [Searchable library checkpoint, deduplication decisions and commands](library/README.md)
- [Measured roles, pods, teams, cards and canvas composition](library/dynamic-components/README.md)
- [Capability review deck and reproducible spec](library/showcase/README.md)
- [101-template implementation and native review workflow](cmd/pptxtemplate/README.md)
- [Agent skill pack](skills/pptxgengo/SKILL.md)
- [Source corpus and inventory policy](planning/README.md)
- [UHG reconstruction results, native placement experiments, and QA evidence](planning/RECONSTRUCTION_CHECKPOINT.md)
- [Go port conventions](PORTING.md)
- [Reviewed native text reconciliation](docs/text-reconciliation.md)
- [Testing lanes and integration dependencies](docs/testing.md)
- [Engineering follow-up status and qualification](docs/engineering-followups.md)
- [Historical code review and resolution log](REVIEW.md)

```sh
go build ./...
make test
make test-race
```

`make test` runs the bounded short suite. The full integration and exhaustive
race lanes are documented in [docs/testing.md](docs/testing.md).

## Install or update the agent skill

From an authenticated private GitLab checkout, run
`python3 scripts/install-skill.py --agent both` to install or update the complete
West Monroe skill for Codex and Claude Code. On Windows use `py -3` in place of
`python3`. Previous skill contents are backed up; CLI and deck pins are preserved.
See [skill installation](docs/skill-installation.md) for project scope, legacy
locations, updates and rollback, and [project upgrades](skills/west-monroe-presentations/references/upgrading-projects.md)
for schema and template changes.

## Native reconstruction experiment

```sh
go run ./cmd/pptxscene extract \
  --source "samples/software modernization campaign pick deck v1 - Repaired.pptx" \
  --slides 1 --out /tmp/modernization-scene

go run ./cmd/pptxscene build \
  --project /tmp/modernization-scene \
  --freeze-slide-numbers --out /tmp/modernization-rebuilt.pptx

go run ./cmd/pptxdiff --reference reference.png \
  --candidate candidate.png --out /tmp/pixel-comparison
```

The scene compiler regenerates native objects from extracted value bindings
while retaining source topology, layouts, masters, themes and assets. It is
an experimental path for reuse and revision, not a new-content design engine.
[`pptxanchor`](cmd/pptxanchor/README.md) calculates visible-art placement from
measured phrase bounds. Native rendering and verification use
the macOS PowerPoint adapter in `scripts/`. An opt-in
[pure Go font layout prototype](library/dynamic-components/go-layout.md)
supports `compose measure|fit-report|build --engine go` without PowerPoint or
separate scripts. Its [height and spacing calibration](library/dynamic-components/font-height-calibration.md)
records native character-height controls and PDF baseline comparisons.
Measured composition supports
explicit installed font families such as IBM Plex Sans; see
[font selection, measurement and caching](library/dynamic-components/fonts.md).

The TypeScript source, distributions, demos, and documentation below are retained
from upstream for reference and golden-output comparisons. They describe
PptxGenJS rather than the Go CLI.

## Upstream PptxGenJS documentation

![PptxGenJS Sample Slides](https://raw.githubusercontent.com/gitbrent/PptxGenJS/gh-pages/img/readme_banner.png)

![jsdelivr downloads](https://data.jsdelivr.com/v1/package/gh/gitbrent/pptxgenjs/badge)
![NPM Downloads](https://img.shields.io/npm/dm/pptxgenjs?style=flat-square)
![GitHub Repo stars](https://img.shields.io/github/stars/gitbrent/pptxgenjs?style=flat-square)
![GitHub License](https://img.shields.io/github/license/gitbrent/pptxgenjs?style=flat-square)
![TypeScript defs](https://img.shields.io/npm/types/pptxgenjs?style=flat-square)

## 🚀 Features

**PptxGenJS lets you generate professional PowerPoint presentations in JavaScript - directly from Node, React, Vite, Electron, or even the browser.**
The library outputs standards-compliant Open Office XML (OOXML) files compatible with:

- ✅ Microsoft PowerPoint
- ✅ Apple Keynote
- ✅ LibreOffice Impress
- ✅ Google Slides (via import)

Design custom slides, charts, images, tables, and templates programmatically - no PowerPoint install or license required.

### Works Everywhere

- Supports every major modern browser - desktop and mobile
- Seamlessly integrates with **Node.js**, **React**, **Angular**, **Vite**, and **Electron**
- Compatible with **PowerPoint**, **Keynote**, **LibreOffice**, and other OOXML apps

### Full-Featured

- Create all major slide objects: **text, tables, shapes, images, charts**, and more
- Define custom **Slide Masters** for consistent academic or corporate branding
- Supports **SVGs**, **animated GIFs**, **YouTube embeds**, **RTL text**, and **Asian fonts**

### Simple & Powerful

- Ridiculously easy to use - create a presentation in 4 lines of code
- Full **TypeScript definitions** for autocomplete and inline documentation
- Includes **75+ demo slides** covering every feature and usage pattern

### Export Your Way

- Instantly download `.pptx` files from the browser with proper MIME handling
- Export as **base64**, **Blob**, **Buffer**, or **Node stream**
- Supports compression and advanced output options for production use

### HTML to PowerPoint Magic

- Convert any HTML `<table>` to one or more slides with a single line of code → [Explore the HTML-to-PPTX feature](#html-to-powerpoint-magic)

## 🌐 Live Demos

Try PptxGenJS right in your browser - no setup required.

- [Basic Slide Demo](https://gitbrent.github.io/PptxGenJS/demos/) - Build a basic presentation in seconds
- [Full Feature Showcase](https://gitbrent.github.io/PptxGenJS/demo/browser/index.html) - Explore every available feature

> Perfect for testing compatibility or learning by example - all demos run 100% in the browser.

## 📦 Installation

Choose your preferred method to install **PptxGenJS**:

### Quick Install (Node-based)

```bash
npm install pptxgenjs
```

```bash
yarn add pptxgenjs
```

### CDN (Browser Usage)

Use the bundled or minified version via [jsDelivr](https://www.jsdelivr.com/package/gh/gitbrent/pptxgenjs):

```html
<script src="https://cdn.jsdelivr.net/gh/gitbrent/pptxgenjs/dist/pptxgen.bundle.js"></script>
```

> Includes the sole dependency (JSZip) in one file.

📁 Advanced: Separate Files, Direct Download

Download from GitHub: [Latest Release](https://github.com/gitbrent/PptxGenJS/releases/latest)

```html
<script src="PptxGenJS/libs/jszip.min.js"></script>
<script src="PptxGenJS/dist/pptxgen.min.js"></script>
```

## 🚀 Universal Compatibility

PptxGenJS works seamlessly in **modern web and Node environments**, thanks to dual ESM and CJS builds and zero runtime dependencies. Whether you're building a CLI tool, an Electron app, or a web-based presentation builder, the library adapts automatically to your stack.

### Supported Platforms

- **Node.js** – generate presentations in backend scripts, APIs, or CLI tools
- **React / Angular / Vite / Webpack** – just import and go, no config required
- **Electron** – build native apps with full filesystem access and PowerPoint output
- **Browser (Vanilla JS)** – embed in web apps with direct download support
- **Serverless / Edge Functions** – use in AWS Lambda, Vercel, Cloudflare Workers, etc.

> _Vite, Webpack, and modern bundlers automatically select the right build via the `exports` field in `package.json`._

### Builds Provided

- **CommonJS**: [`dist/pptxgen.cjs.js`](./dist/pptxgen.cjs.js)
- **ES Module**: [`dist/pptxgen.es.js`](./dist/pptxgen.es.js)

## 📖 Documentation

### Quick Start Guide

PptxGenJS PowerPoint presentations are created via JavaScript by following 4 basic steps:

#### Angular/React, ES6, TypeScript

```typescript
import pptxgen from "pptxgenjs";

// 1. Create a new Presentation
let pres = new pptxgen();

// 2. Add a Slide
let slide = pres.addSlide();

// 3. Add one or more objects (Tables, Shapes, Images, Text and Media) to the Slide
let textboxText = "Hello World from PptxGenJS!";
let textboxOpts = { x: 1, y: 1, color: "363636" };
slide.addText(textboxText, textboxOpts);

// 4. Save the Presentation
pres.writeFile();
```

#### Script/Web Browser

```javascript
// 1. Create a new Presentation
let pres = new PptxGenJS();

// 2. Add a Slide
let slide = pres.addSlide();

// 3. Add one or more objects (Tables, Shapes, Images, Text and Media) to the Slide
let textboxText = "Hello World from PptxGenJS!";
let textboxOpts = { x: 1, y: 1, color: "363636" };
slide.addText(textboxText, textboxOpts);

// 4. Save the Presentation
pres.writeFile();
```

That's really all there is to it!

## 💥 HTML-to-PowerPoint Magic

Convert any HTML `<table>` into fully formatted PowerPoint slides - automatically and effortlessly.

```javascript
let pptx = new pptxgen();
pptx.tableToSlides("tableElementId");
pptx.writeFile({ fileName: "html2pptx-demo.pptx" });
```

Perfect for transforming:

- Dynamic dashboards and data reports
- Exportable grids in web apps
- Tabular content from CMS or BI tools

[View Full Docs & Live Demo](https://gitbrent.github.io/PptxGenJS/html2pptx/)

## 📚 Full Documentation

Complete API reference, tutorials, and integration guides are available on the official docs site: [https://gitbrent.github.io/PptxGenJS](https://gitbrent.github.io/PptxGenJS)

## 🛠️ Issues / Suggestions

Please file issues or suggestions on the [issues page on github](https://github.com/gitbrent/PptxGenJS/issues/new), or even better, [submit a pull request](https://github.com/gitbrent/PptxGenJS/pulls). Feedback is always welcome!

When reporting issues, please include a code snippet or a link demonstrating the problem.
Here is a small [jsFiddle](https://jsfiddle.net/gitbrent/L1uctxm0/) that is already configured and uses the latest PptxGenJS code.

## 🆘 Need Help?

Sometimes implementing a new library can be a difficult task and the slightest mistake will keep something from working. We've all been there!

If you are having issues getting a presentation to generate, check out the code in the `demos` directory. There
are demos for browser, node and, react that contain working examples of every available library feature.

- Use a pre-configured jsFiddle to test with: [PptxGenJS Fiddle](https://jsfiddle.net/gitbrent/L1uctxm0/)
- [View questions tagged `PptxGenJS` on StackOverflow](https://stackoverflow.com/questions/tagged/pptxgenjs?sort=votes&pageSize=50). If you can't find your question, [ask it yourself](https://stackoverflow.com/questions/ask?tags=PptxGenJS) - be sure to tag it `pptxgenjs`.
- Ask your AI pair programmer! All major LLMs have ingested the pptxgenjs library and have the ability to answer functionality questions and provide code.

## 🙏 Contributors

Thank you to everyone for the contributions and suggestions! ❤️

Special Thanks:

- [Dzmitry Dulko](https://github.com/DzmitryDulko) - Getting the project published on NPM
- [Michal Kacerovský](https://github.com/kajda90) - New Master Slide Layouts and Chart expertise
- [Connor Bowman](https://github.com/conbow) - Adding Placeholders
- [Reima Frgos](https://github.com/ReimaFrgos) - Multiple chart and general functionality patches
- [Matt King](https://github.com/kyrrigle) - Chart expertise
- [Mike Wilcox](https://github.com/clubajax) - Chart expertise
- [Joonas](https://github.com/wyozi) - [react-pptx](https://github.com/wyozi/react-pptx)

PowerPoint shape definitions and some XML code via [Officegen Project](https://github.com/Ziv-Barber/officegen)

## 🌟 Support the Open Source Community

If you find this library useful, consider contributing to open-source projects, or sharing your knowledge on the open social web. Together, we can build free tools and resources that empower everyone.

[@gitbrent@fosstodon.org](https://fosstodon.org/@gitbrent)

## 📜 License

Copyright &copy; 2015-present [Brent Ely](https://github.com/gitbrent/)

[MIT](https://github.com/gitbrent/PptxGenJS/blob/master/LICENSE)
