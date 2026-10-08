# Windows workflows

Use the private GitLab Windows archive matching the machine architecture. It
contains `pptxgengo.exe`, `pptxdesign.exe`, `wmdsdocs.exe`, the template source
bundle/SQLite index/fonts/gallery and template browsing deck. The skill is
installed separately. Go and Python are not required for ordinary CLI authoring.
Use the shared [installation workflow](installation.md); check command help for
this package rather than assuming a legacy preview ZIP has the same installer.

## Authoring

The CLI verbs and YAML contracts are shared. Resolve packaged paths with
`pptxgengo paths`; keep `/` in project-relative YAML references and pass Windows
filesystem paths as quoted PowerShell arguments. `catalog --open` uses the
Windows file handler; `docs` serves the documentation locally.

```powershell
pptxgengo design project check --project 'C:\Work\My deck'
pptxgengo design project build --project 'C:\Work\My deck'
pptxgengo design project measure --report 'C:\Work\My deck\builds\BUILD-ID\layout-report.json'
```

A project copied from macOS retains an OS/compiler lock. Work on a copy, run
`project migrate --dry-run`, then `project migrate`, rebuild and visually review.
The original lock is preserved; incompatible custom compositions require repair.
Never silently delete the lock to get around a drift error.

Private photo/branding originals are separate. Set `WMDS_BRANDING_ROOT` to the
matching collection when a selected registered asset requires it, or register an
operator-provided image in the project. Template browsing media are illustrative
placeholders, not originals to copy into a final client deck.

## Native review

Windows CLI binaries are signed, but interactive PowerPoint qualification
remains pending. Record an actual export and visual review for the current deck. It needs desktop PowerPoint in the
signed-in user session, Windows PowerShell 5.1 and the bundled IBM Plex fonts.
PowerPoint COM exports PDF and slide PNGs; Go creates the contact sheet and signs
the receipt. PNGs are direct PowerPoint slide exports, not PDFKit rasterizations.

```powershell
pptxgengo design render-doctor --json --timeout 60s
pptxgengo design render --pptx 'C:\Work\My deck\builds\BUILD-ID\deck.pptx' --out 'C:\Work\review-1' --pdf --png --contact-sheet
```

Inspect every full-size PNG before attaching review decisions. Font substitutions,
Office version differences and policy prompts can change appearance. The doctor
uses the same staging folder as render (default `%LOCALAPPDATA%\pptxgengo\native`).
It tests actual PowerPoint open, PDF/PNG export and font availability. NTFS is
required for task ownership and the private receipt issuer. The source PPTX and
unrelated open decks are preserved.

On failure, inspect `render-error.txt` and the PowerPoint window. Ask the operator
to resolve reported setup, activation, Protected View or enterprise policy issues;
do not change global policy or repeatedly retry. Manual PowerPoint export is a
fallback for visual review and does not produce a signed CLI receipt. Do not use
macOS Grant File Access instructions on Windows, run AppleScript/Swift commands,
or claim that macOS source specimens qualify a new Windows-rendered deck.
