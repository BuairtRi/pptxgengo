# Testing and historical PptxGenJS checks

Current Go test lanes, private GitLab CI policy and platform qualification are
documented in [docs/testing.md](docs/testing.md). GitHub Actions is disabled;
GitHub run links elsewhere in this repository are dated historical evidence,
not current CI gates. The sections below preserve the original writer baseline
and inherited upstream JavaScript guide.

Use Go 1.27.1 or newer:

```sh
go build ./...
go test -race -cover ./...
```

Writer tests live in `pptx/`; eight golden cases compare generated package parts with
the retained PptxGenJS 4.0.1 output, including embedded workbooks. See
[`pptx/testdata/golden/README.md`](pptx/testdata/golden/README.md).

The historical 2026-09-25 baseline passed both commands on Go 1.27.1, darwin/arm64,
with 78.1% statement coverage. These are code and XML checks, not rendered
presentation acceptance. See [PRODUCT_PLAN.md](PRODUCT_PLAN.md) for proposed
visual, narrative, reuse, and round-trip experiments.

Later source work added authoring, project, template and packaging tests; do not
use the baseline figures above as current aggregate coverage. The experimental
scene, anchor and pixel tools also have tests under
`internal/nativepkg/` and `cmd/`. Their latest prior `go test -race ./...`
run passed after the highlight implementation. The 78.1% figure above belongs
to the earlier writer baseline, not an updated aggregate for these tools.
Native UHG benchmark scripts require separately supplied local sources and
ignored scene/measurement artifacts; see [the checkpoint](planning/RECONSTRUCTION_CHECKPOINT.md).

## Inherited PptxGenJS testing guide

The following instructions apply to the retained JavaScript project.

This document outlines how to manually test PptxGenJS across supported platforms and environments prior to release.

> ✅ Run these tests to ensure compatibility with major bundlers, runtimes, and front-end frameworks.

Config Notes

> ⚠️ Disable VPN on the server machine, otherwise, clients using the local IP address cannot connect.

Testing Steps

1. Run `npm run ship`
2. Execute tests from each section below

## 🧪 Test Suites Overview

| Platform        | Tooling              | Status |
| --------------- | -------------------- | ------ |
| Browser         | Standalone HTML demo | ✅      |
| Node.js         | Native CLI           | ✅      |
| Web Worker      | JS Worker demo       | ✅      |
| Vite/TypeScript | Modern front-end SPA | ✅      |
| Webpack         | SharePoint Framework | ✅      |

---

## 🌐 Browser Tests

**Purpose:** Validate browser compatibility using the standalone bundle as script.

### Desktop & Mobile Browsers

Run local test server:

```bash
cd demos
node browser_server.mjs
```

1. Open the [Demo Page](http://localhost:8000/browser/index.html).
2. In DevTools, confirm the latest `pptxgen.bundle.js` is loaded (`Sources` tab).
3. Run all UI-driven demos and verify demo presentation render correctly.
4. Open the [Demo Page](http://192.168.254.x:8000/browser/index.html) on iPhone & test.

### Web Worker API

1. Open the [Web Worker Demo Page](localhost:8000/browser/worker_test.html).
2. Note: Use Chrome (Safari *will not work*)
3. Run the test; verify result & library version

### Microsoft 365 Check

1. Upload the full demo output from above to M365/Office/OneDrive.
2. Use web viewer to validate file

---

## 📦 Node.js Tests

**Purpose:** Validate functionality of CommonJS module in pure Node environments.

### CLI Tests

Run the following test commands:

```bash
cd demos/node
npm run demo
npm run demo-all
```

1. Confirm console output and exported PPTX files are correct.

### Stream Test

```bash
npm run demo-stream
```

1. Confirm stream download PPTX file is correct.
2. Open the [Stream URL](http://192.168.254.x:3000/) on iPhone & test.

---

## ⚛️ Vite + TypeScript Tests

**Purpose:** Validate integration in modern front-end SPA toolchains (Vite, TypeScript, React-compatible).

Ensure the latest files below are copied to local `node_modules`:

- `dist/pptxgen.es.js`
- `types/index.d.ts`

1. Update `package.json` (and `package-lock.json` if needed) in `demos/vite-demo/`
2. Check for TS errors in files:

- Open `src/tstest/Test.tsx`
- Use IntelliSense to autocomplete things like `pptxgen.ChartType.`

Start the app:

```bash
cd demos/vite-demo
npm install (?)
npm run dev
```

From your network:

- MacBook..: [Demo](http://localhost:8080/PptxGenJS/)
- iPhone...: [Demo](http://192.168.254.x:8080/PptxGenJS/)
- Android..: [Demo](http://192.168.254.x:8080/PptxGenJS/)

1. Run test slides, export PowerPoint files.
2. Open files on each device to verify:

- MIME type is valid
- File renders as expected in PowerPoint or previewer

---

## 🚀 Build for gh-pages (Manual)

After confirming the above:

```bash
npm run build
```

1. Copy the entire `dist` folder from `demos/vite-demo/` to a safe location.
2. Use this copy when updating the `gh-pages` branch after the release.

> ⚠️ DO NOT use the "deploy" script displayed onscreen by Vite. Manual copying ensures full control over final content.

---

## 🏁 Test Completion Checklist

| Dist File         | Test       | Tested Via             | Result |
| ----------------- | ---------- | ---------------------- | ------ |
| pptxgen.es.js     | Webpack 4  | SPFx (v1.16.1) project | ✅?🟡    |
| pptxgen.es.js     | Webpack 5  | SPFx (v1.19.1) project | ✅?🟡    |
| pptxgen.es.js     | Rollup 4   | Vite (v6) demo         | ✅?🟡    |
| pptxgen.es.js     | Webworkers | worker_test demo       | ✅?🟡    |
| pptxgen.cjs.js    | Node/CJS   | Node demo              | ✅?🟡    |
| pptxgen.bundle.js | Script     | Browser demo (desktop) | ✅?🟡    |
| pptxgen.bundle.js | Script     | Browser demo (iOS)     | ✅?🟡    |
