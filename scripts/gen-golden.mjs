#!/usr/bin/env node
// scripts/gen-golden.mjs
//
// Generates golden reference .pptx files with the EXISTING JavaScript
// PptxGenJS library (dist/pptxgen.cjs.js, currently v4.0.1) and extracts
// their XML parts into pptx/testdata/golden/<case>/ so the Go port
// (pptx/) can be tested for byte-identical XML output.
//
// Regenerate with:
//   /opt/node22/bin/node scripts/gen-golden.mjs
//
// Requires node_modules/ to be installed at the repo root:
//   /opt/node22/bin/npm install --omit=dev
//
// This script does NOT modify src/, dist/, or package.json, and does not
// leave any .pptx binaries behind — only the extracted XML/rels/media parts
// are written under pptx/testdata/golden/.

import { fileURLToPath } from 'node:url'
import path from 'node:path'
import fs from 'node:fs/promises'
import os from 'node:os'

import PptxGenJS from '../dist/pptxgen.cjs.js'
import JSZip from 'jszip'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = path.resolve(__dirname, '..')
const GOLDEN_ROOT = path.join(REPO_ROOT, 'pptx', 'testdata', 'golden')
const PKG_VERSION = JSON.parse(await fs.readFile(path.join(REPO_ROOT, 'package.json'), 'utf8')).version

// ---------------------------------------------------------------------------
// A tiny 5x5 solid-red PNG, generated once via node:zlib deflate (level 9) of
// raw RGB scanlines (no filters) and hand-built IHDR/IDAT/IEND chunks with a
// manually-computed CRC32 — no external tools or network calls involved.
// Kept as a literal constant so downstream Go tests can byte-compare against
// the identical decoded PNG bytes.
// ---------------------------------------------------------------------------
export const TINY_RED_PNG_BASE64 =
	'iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEUlEQVR42mP4z8CAjBgo5AMA/XwY6DI3DH0AAAAASUVORK5CYII='

const CASES = []
function defineCase(name, build) {
	CASES.push({ name, build })
}

// ---------------------------------------------------------------------------
// 01-basic: default 16x9 layout, one slide, one text box, fixed doc props.
// ---------------------------------------------------------------------------
defineCase('01-basic', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()
	slide.addText('Hello World', { x: 1, y: 1, w: 8, h: 1, fontSize: 24, color: '363636' })
})

// ---------------------------------------------------------------------------
// 02-text-rich: multiple runs, bold/italic/underline, bullets, alignment,
// line breaks, hyperlink, font face change, subscript/superscript.
// ---------------------------------------------------------------------------
defineCase('02-text-rich', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()

	// Multi-run text box: bold / italic / underline / font-face-change / hyperlink
	slide.addText(
		[
			{ text: 'Bold ', options: { bold: true } },
			{ text: 'Italic ', options: { italic: true } },
			{ text: 'Underline ', options: { underline: { style: 'sng' } } },
			{ text: 'CourierFace ', options: { fontFace: 'Courier New' } },
			{ text: 'ExampleLink', options: { hyperlink: { url: 'https://example.com', tooltip: 'Visit Example' } } },
		],
		{ x: 0.5, y: 0.5, w: 9, h: 1, fontSize: 18, align: 'center' }
	)

	// Bulleted list with line breaks
	slide.addText(
		[
			{ text: 'First bullet point', options: { bullet: true, breakLine: true } },
			{ text: 'Second bullet point', options: { bullet: true, breakLine: true } },
			{ text: 'Numbered item one', options: { bullet: { type: 'number' }, breakLine: true } },
			{ text: 'Numbered item two', options: { bullet: { type: 'number' } } },
		],
		{ x: 0.5, y: 2, w: 5, h: 2.5, fontSize: 14 }
	)

	// Alignment variants + subscript/superscript
	slide.addText(
		[
			{ text: 'Left aligned line', options: { align: 'left', breakLine: true } },
			{ text: 'Center aligned line', options: { align: 'center', breakLine: true } },
			{ text: 'Right aligned line', options: { align: 'right', breakLine: true } },
			{ text: 'H', options: {} },
			{ text: '2', options: { subscript: true } },
			{ text: 'O and E=mc', options: {} },
			{ text: '2', options: { superscript: true } },
		],
		{ x: 6, y: 2, w: 3.5, h: 2.5, fontSize: 12 }
	)
})

// ---------------------------------------------------------------------------
// 03-shapes: rect, roundRect, oval w/ fill+line, line w/ arrows, rotated
// triangle, shape w/ shadow.
// ---------------------------------------------------------------------------
defineCase('03-shapes', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()

	slide.addShape('rect', { x: 0.5, y: 0.5, w: 2, h: 1, fill: { color: '2E86AB' }, line: { color: '1B4965', width: 1 } })

	slide.addShape('roundRect', {
		x: 3, y: 0.5, w: 2, h: 1, rectRadius: 0.15,
		fill: { color: 'A23B72' }, line: { color: '6A2049', width: 1 },
	})

	slide.addShape('ellipse', {
		x: 5.5, y: 0.5, w: 2, h: 1,
		fill: { color: 'F18F01' }, line: { color: 'C46F00', width: 2, dashType: 'dash' },
	})

	slide.addShape('line', {
		x: 0.5, y: 2, w: 3, h: 0,
		line: { color: '363636', width: 2, beginArrowType: 'triangle', endArrowType: 'arrow' },
	})

	slide.addShape('triangle', {
		x: 4.5, y: 2, w: 1.5, h: 1.5, rotate: 45,
		fill: { color: '4CAF50' }, line: { color: '2E7D32', width: 1 },
	})

	slide.addShape('rect', {
		x: 7, y: 2, w: 2, h: 1.5,
		fill: { color: 'FFFFFF' }, line: { color: '333333', width: 1 },
		shadow: { type: 'outer', color: '000000', opacity: 0.5, blur: 3, angle: 45, offset: 3 },
	})
})

// ---------------------------------------------------------------------------
// 04-table: 3x3 data, colW array, borders, cell fill, merged (colspan) cell,
// bold+fill header row.
// ---------------------------------------------------------------------------
defineCase('04-table', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()

	const border = { type: 'solid', color: '999999', pt: 1 }

	const rows = [
		[
			{ text: 'Merged Header', options: { colspan: 3, bold: true, fill: { color: '2E86AB' }, color: 'FFFFFF', align: 'center', border } },
		],
		[
			{ text: 'Col A', options: { bold: true, fill: { color: 'DDDDDD' }, border } },
			{ text: 'Col B', options: { bold: true, fill: { color: 'DDDDDD' }, border } },
			{ text: 'Col C', options: { bold: true, fill: { color: 'DDDDDD' }, border } },
		],
		[
			{ text: 'r1c1', options: { border } },
			{ text: 'r1c2', options: { border, fill: { color: 'FFF3CD' } } },
			{ text: 'r1c3', options: { border } },
		],
		[
			{ text: 'r2c1', options: { border } },
			{ text: 'r2c2', options: { border } },
			{ text: 'r2c3', options: { border } },
		],
	]

	slide.addTable(rows, { x: 0.5, y: 0.5, w: 9, colW: [3, 3, 3], border })
})

// ---------------------------------------------------------------------------
// 05-chart-bar: single-series bar chart, title, cat/val axis titles shown,
// fixed chartColors.
// ---------------------------------------------------------------------------
defineCase('05-chart-bar', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()

	const data = [
		{ name: 'Revenue', labels: ['Q1', 'Q2', 'Q3', 'Q4'], values: [100, 150, 130, 175] },
	]

	slide.addChart(pptx.ChartType.bar, data, {
		x: 0.5, y: 0.5, w: 9, h: 5,
		showTitle: true, title: 'Quarterly Revenue',
		showCatAxisTitle: true, catAxisTitle: 'Quarter',
		showValAxisTitle: true, valAxisTitle: 'USD (thousands)',
		chartColors: ['2E86AB', 'A23B72', 'F18F01', '4CAF50'],
	})
})

// ---------------------------------------------------------------------------
// 06-chart-multi: line chart w/ 2 series on slide 1, pie chart on slide 2.
// ---------------------------------------------------------------------------
defineCase('06-chart-multi', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide1 = pptx.addSlide()
	const lineData = [
		{ name: 'Series A', labels: ['Jan', 'Feb', 'Mar', 'Apr'], values: [10, 20, 15, 25] },
		{ name: 'Series B', labels: ['Jan', 'Feb', 'Mar', 'Apr'], values: [5, 12, 18, 9] },
	]
	slide1.addChart(pptx.ChartType.line, lineData, {
		x: 0.5, y: 0.5, w: 9, h: 5,
		showTitle: true, title: 'Two Series Line Chart',
		showLegend: true,
		chartColors: ['2E86AB', 'A23B72'],
	})

	const slide2 = pptx.addSlide()
	const pieData = [
		{ name: 'Share', labels: ['Alpha', 'Beta', 'Gamma'], values: [40, 35, 25] },
	]
	slide2.addChart(pptx.ChartType.pie, pieData, {
		x: 1, y: 0.5, w: 7, h: 5,
		showTitle: true, title: 'Market Share',
		showLegend: true,
		chartColors: ['2E86AB', 'A23B72', 'F18F01'],
	})
})

// ---------------------------------------------------------------------------
// 07-image: addImage with an embedded base64 tiny PNG.
// ---------------------------------------------------------------------------
defineCase('07-image', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	const slide = pptx.addSlide()
	slide.addImage({
		data: `image/png;base64,${TINY_RED_PNG_BASE64}`,
		x: 1, y: 1, w: 2, h: 2,
	})
})

// ---------------------------------------------------------------------------
// 08-master: defineSlideMaster w/ title, background color, placeholder,
// slideNumber; a slide using it; plus addSection.
// ---------------------------------------------------------------------------
defineCase('08-master', pptx => {
	pptx.author = 'PptxGenGo Test Author'
	pptx.company = 'PptxGenGo Test Co'
	pptx.subject = 'PptxGenGo Golden Test Subject'
	pptx.title = 'PptxGenGo Golden Test Title'

	pptx.defineSlideMaster({
		title: 'GOLDEN_MASTER',
		background: { color: 'F1F1F1' },
		slideNumber: { x: 9, y: 5.5, color: '363636' },
		objects: [
			{
				placeholder: {
					options: { name: 'title', type: 'title', x: 0.5, y: 0.3, w: 9, h: 1 },
					text: 'Click to add title',
				},
			},
			{ rect: { x: 0, y: 5.3, w: '100%', h: 0.3, fill: { color: '2E86AB' } } },
		],
	})

	pptx.addSection({ title: 'Golden Section' })

	const slide = pptx.addSlide({ masterName: 'GOLDEN_MASTER', sectionTitle: 'Golden Section' })
	slide.addText('Slide using GOLDEN_MASTER', { placeholder: 'title' })
})

// ---------------------------------------------------------------------------
// Runner
// ---------------------------------------------------------------------------

async function unzipToDir(pptxBuffer, destDir) {
	const zip = await JSZip.loadAsync(pptxBuffer)
	const entries = Object.keys(zip.files).sort()
	const written = []
	for (const entryName of entries) {
		const entry = zip.files[entryName]
		if (entry.dir) continue
		const destPath = path.join(destDir, entryName)
		await fs.mkdir(path.dirname(destPath), { recursive: true })
		const content = await entry.async('nodebuffer')
		await fs.writeFile(destPath, content)
		written.push(entryName)
	}
	return written
}

async function main() {
	const tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'pptxgengo-golden-'))
	const results = []

	for (const { name, build } of CASES) {
		const golden = path.join(GOLDEN_ROOT, name)
		try {
			await fs.rm(golden, { recursive: true, force: true })
			await fs.mkdir(golden, { recursive: true })

			const pptx = new PptxGenJS()
			build(pptx)

			const outFile = path.join(tmpDir, `${name}.pptx`)
			await pptx.writeFile({ fileName: outFile })
			const buf = await fs.readFile(outFile)

			const written = await unzipToDir(buf, golden)
			results.push({ name, ok: true, files: written.length, list: written })
		} catch (err) {
			results.push({ name, ok: false, error: err && err.stack ? err.stack : String(err) })
		}
	}

	await fs.rm(tmpDir, { recursive: true, force: true })

	// README.md for the golden directory
	const readmeLines = []
	readmeLines.push('# Golden reference test data')
	readmeLines.push('')
	readmeLines.push(`Generated by \`scripts/gen-golden.mjs\` using pptxgenjs **v${PKG_VERSION}** (dist/pptxgen.cjs.js) under Node ${process.version}.`)
	readmeLines.push('')
	readmeLines.push('Each subdirectory contains the extracted XML/rels/media parts of a .pptx produced by the real JS library, for byte-for-byte comparison against the Go port\'s output. The .pptx binaries themselves are not kept — only the extracted parts.')
	readmeLines.push('')
	readmeLines.push('To regenerate: `/opt/node22/bin/node scripts/gen-golden.mjs` from the repo root (requires `npm install --omit=dev` first).')
	readmeLines.push('')
	readmeLines.push('## Cases')
	readmeLines.push('')
	for (const r of results) {
		if (r.ok) {
			readmeLines.push(`- \`${r.name}/\` — ${r.files} files`)
		} else {
			readmeLines.push(`- \`${r.name}/\` — FAILED: ${r.error.split('\n')[0]}`)
		}
	}
	readmeLines.push('')
	await fs.writeFile(path.join(GOLDEN_ROOT, 'README.md'), readmeLines.join('\n'))

	// Print inventory
	console.log('=== Golden generation report ===')
	console.log('pptxgenjs version:', PKG_VERSION)
	console.log('node version:', process.version)
	console.log('')
	for (const r of results) {
		if (r.ok) {
			console.log(`[OK] ${r.name}: ${r.files} files`)
			for (const f of r.list) console.log(`       ${f}`)
		} else {
			console.log(`[FAIL] ${r.name}:`)
			console.log(r.error.split('\n').map(l => '       ' + l).join('\n'))
		}
	}

	const failed = results.filter(r => !r.ok)
	if (failed.length > 0) {
		console.error(`\n${failed.length} case(s) failed.`)
		process.exitCode = 1
	} else {
		console.log(`\nAll ${results.length} cases generated successfully.`)
	}
}

main()
