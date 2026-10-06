// Bounded browser regression for the four density contrast repairs.
// Build first: python3 tools/build_explorations.py --only components --out /tmp/wmds-board
// Run: NODE_PATH=<checkout>/.ds-sync/node_modules node tools/check_density_contrast.js /tmp/wmds-board/components.html
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {pathToFileURL} = require('node:url');
const {chromium} = require('playwright');
(async () => {
  const browser = await chromium.launch({executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
  try {
    const page = await browser.newPage({viewport: {width: 1920, height: 1080}});
    const keys = ['pillars/two-categories-six-magenta', 'pillars/two-categories-six-stacked-magenta', 'capability-heat/annotated', 'risk-heat/annotated'];
    const rows = [];
    for (const key of keys) {
      const [id, variant] = key.split('/');
      await page.goto(pathToFileURL(process.argv[2]).href + '?preview=' + id + '&variant=' + variant);
      await page.waitForSelector('body[data-ready="1"]');
      await page.evaluate(() => document.fonts.ready);
      for (const density of ['comfortable', 'compact', 'dense']) {
        const result = await page.evaluate(async (density) => {
          document.body.setAttribute('data-dview', density);
          // Use the board's density control so the same warning refresh path runs.
          const select = document.querySelector('#densityView');
          if (select) { select.value = density; select.dispatchEvent(new Event('change')); }
          // Preview removes the toolbar; redraw warnings through the actual public board mode below.
          await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
          const stage = document.querySelector('.stage'), scale = stage.getBoundingClientRect().width / 960;
          const tiles = [...stage.querySelectorAll('.num-tile')].map(t => {
            const cs = getComputedStyle(t), b = t.getBoundingClientRect(), title = t.parentElement.nextElementSibling, tb = title.getBoundingClientRect(), ts = getComputedStyle(title);
            return {w: b.width/scale, h: b.height/scale, size: parseFloat(cs.fontSize)/scale, weight: cs.fontWeight, color: cs.color, fill: cs.backgroundColor,
              gap: (tb.left-b.right)/scale, centerOffset: (b.top+b.height/2-tb.top-parseFloat(ts.lineHeight)/2)/scale, border: parseFloat(cs.borderTopWidth), warning: t.dataset.warn || ''};
          });
          const checked = [...stage.querySelectorAll('[data-contrast-fg]')].map(t => ({text:t.textContent, fg:t.dataset.contrastFg, bg:t.dataset.contrastBg, size:parseFloat(getComputedStyle(t).fontSize)/scale, weight:getComputedStyle(t).fontWeight}));
          return {tiles, checked};
        }, density);
        if (key.startsWith('pillars/')) {
          assert.equal(result.tiles.length, 6, key);
          for (const t of result.tiles) {
            for (const [field, expected] of [['w',27],['h',27],['size',14],['gap',9],['centerOffset',0],['border',0]]) assert.ok(Math.abs(t[field]-expected)<0.02, key+' '+density+' '+field+' '+t[field]);
            assert.equal(t.weight,'600'); assert.equal(t.color,'rgb(7, 1, 84)'); assert.equal(t.fill,'rgb(249, 0, 211)'); assert.equal(t.warning,'');
          }
        } else assert.ok(result.checked.some(t => t.bg==='#F900D3' && t.fg==='#070154'), key+' Grounded callout title');
        rows.push({key,density,tiles:result.tiles});
        if (process.argv[3]) {fs.mkdirSync(process.argv[3],{recursive:true});await page.screenshot({path:path.join(process.argv[3],key.replace('/','-')+'-'+density+'.png')});}
      }
    }
    // Native contrast audit exposed two previously uninstrumented primitives.
    // Assert actual colors and checker coverage for all three approved specimens at every density.
    const primitiveRows = [];
    for (const key of ['maturity/ai-beyond', 'maturity/insights', 'road-fork/decision-chosen']) {
      const [id, variant] = key.split('/');
      await page.goto(pathToFileURL(process.argv[2]).href + '?preview=' + id + '&variant=' + variant);
      await page.waitForSelector('body[data-ready="1"]');
      await page.evaluate(() => document.fonts.ready);
      for (const density of ['comfortable', 'compact', 'dense']) {
        const checks = await page.evaluate(async ({density, key}) => {
          document.body.setAttribute('data-dview', density);
          await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r)));
          const stage = document.querySelector('.stage'), scale = stage.getBoundingClientRect().width / 960;
          return [...stage.querySelectorAll('[data-contrast-fg]')].filter(el => key.startsWith('maturity/') ? el.textContent === 'Inflection' : el.dataset.contrastNote === 'roadfork pin numeral').map(el => {
            const cs = getComputedStyle(el);
            return {text:el.textContent, fg:el.dataset.contrastFg, bg:el.dataset.contrastBg, color:cs.color, size:parseFloat(cs.fontSize)/scale, ring:cs.boxShadow, warning:el.dataset.warn||''};
          });
        }, {density,key});
        assert.ok(checks.length, key + ' contrast instrumentation');
        if (key.startsWith('maturity/')) {
          assert.equal(checks.length,1); assert.equal(checks[0].fg,'#0047FF'); assert.equal(checks[0].color,'rgb(0, 71, 255)');
          assert.equal(checks[0].bg,'#FFFFFF'); assert.equal(checks[0].warning,'');
        } else {
          const inactive = checks.filter(t => t.ring.includes('rgb(151, 164, 186)'));
          assert.equal(inactive.length,2, 'two inactive branch pins retain gray rings');
          for (const t of inactive) {assert.equal(t.fg,'#070154');assert.equal(t.color,'rgb(7, 1, 84)');assert.equal(t.bg,'#FFFFFF');assert.equal(t.size,10);assert.equal(t.warning,'');}
        }
        primitiveRows.push({key,density,checks});
      }
    }
    // Exercise the board control and warning checker on an ordinary NUMBER (600): 18pt passes3:1, 16/14 need4.5.
    await page.goto(pathToFileURL(process.argv[2]).href + '#templates');
    await page.evaluate(async () => {
      await document.fonts.ready;
      const stage = document.querySelector('#panel-templates .stage');
      const probe = document.createElement('div'); probe.className = 'st-number density-contrast-regression'; probe.textContent = '01';
      probe.style.cssText = 'left:0;top:0;color:#F900D3;background:#E8EEF8';
      Object.assign(probe.dataset,{contrastFg:'#F900D3',contrastBg:'#E8EEF8',contrastKind:'large',contrastNote:'contrast'});stage.appendChild(probe);
    });
    for (const density of ['comfortable', 'compact', 'dense']) {
      await page.selectOption('#densityView', density);
      await page.evaluate(() => new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r))));
      const checks = await page.evaluate(() => [...document.querySelectorAll('.stage [data-contrast-fg]')].filter(t => t.classList.contains('density-contrast-regression')).map(t => ({minimum:+t.dataset.contrastMinimum, warning:t.dataset.warn||''})));
      assert.ok(checks.length, 'existing Magenta NUMBER checks');
      for (const t of checks) {assert.equal(t.minimum,density==='comfortable'?3:4.5); assert.equal(!!t.warning,density!=='comfortable');}
    }
    // Reject invalid generic tile authoring just as the strict native renderer does.
    // Mutate parsed fixture data only, before rendering; the committed templates stay untouched.
    for (const invalid of ['surface', 'inlineNumber', 'title', 'band']) {
      const bad = await browser.newPage({viewport:{width:1920,height:1080}});
      await bad.addInitScript((invalid) => {
        const parse = JSON.parse;
        JSON.parse = function(...args) {
          const value = parse.apply(this,args);
          if (Array.isArray(value)) {
            const template = value.find(t => t && t.id === 'pillars' && t.variant === 'two-categories-six-magenta');
            if (template) {
              const card = template.slide.body.find(n => n.type === 'card' && n.numTile);
              if (invalid === 'surface') card.numTile = 'inverse';
              else if (invalid === 'band') card.band = {surface:'callout'};
              else delete card[invalid];
            }
          }
          return value;
        };
      },invalid);
      const error = bad.waitForEvent('pageerror',{timeout:10000});
      await bad.goto(pathToFileURL(process.argv[2]).href+'?preview=pillars&variant=two-categories-six-magenta');
      assert.match((await error).message,/unsupported card numTile/,invalid);
      await bad.close();
    }
    console.log(JSON.stringify({status:'passed',specimens:rows.length,invalidTileCases:4,primitiveSpecimens:primitiveRows.length,primitiveRows,rows},null,2));
  } finally {await browser.close();}
})().catch(e => {console.error(e);process.exit(1);});
