# Windows tester package

This is an experimental Windows build of pptxgengo, with the V11 template
library, SQLite search index, visual gallery, documentation, and current West
Monroe presentations skill. Go/Python are not needed to use it. No administrator
rights are required for the user installation.

Open **`guides/00-start.html`** in your browser for the illustrated installation,
existing-deck and agent revision guides. Every HTML file runs offline without a
server or external assets; keep them together for navigation.

## Install

1. Extract the ZIP completely to a local folder. Do not run from inside the ZIP.
2. Open PowerShell in the extracted `pptxgengo-windows-amd64` folder and run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\install-windows.ps1
```

The installer verifies the complete package, native architecture and required
resources, checks a copied stage and all three executable versions, then activates
through the shared installation manager. Releases are immutable; retrying an
identical package is supported. A different package cannot overwrite a release.
User PATH uses one stable `%LOCALAPPDATA%\pptxgengo\bin` entry across upgrades.
The skill is installed into `%USERPROFILE%\.codex\skills` (or `CODEX_HOME`), with
existing user skills backed up. Fonts are installed separately for this user.

Activation records a durable receipt before changing settings. A failed activation
restores the previous selection; an interrupted operation can be recovered with:

```powershell
pptxgengo installation doctor
pptxgengo installation recover
pptxgengo installation rollback
```

Recovery preserves conflicting edits and reports the receipt requiring attention.
Open a new terminal and restart the agent after activation or rollback. Font
registration is a separate best-effort step and is not rolled back. Actual Windows
installer/PATH and PowerPoint qualification remain pending the desktop runner.

`-SkipSkill`, `-SkipFonts` and `-NoPath` opt out separately; `-StageOnly` checks and
stages without activation. A custom `-Destination` must be absent or contain the
same verified package. A custom destination is an additional checked staging copy;
activation still uses the managed `%LOCALAPPDATA%\pptxgengo\releases\VERSION`
selection. Use the Go command's `--root` for a different managed installation root.
x64 packages require an x64 PC and ARM64 packages an ARM64
PC. Enterprise policy can still block scripts or signed executables; this command
does not change machine execution policy.

`pptxgengo installation uninstall` removes owned activation bindings, restores an
unchanged original user skill when recorded, and retains releases/backups. Edited
skills, user projects, authored decks and fonts are preserved. `--keep-skill` keeps
the current skill explicitly. These commands require a package containing the new
installation manager; the earlier `v4.1.0` CLI does not include them.

For portable use, skip installation and invoke `bin\pptxgengo.exe` directly.
Copy `skills\west-monroe-presentations` into your own skill directory and install
the TTFs in `library\wm-design-system\v11\fonts` before native review.

## Try it

```powershell
.\bin\pptxgengo.exe --version
.\bin\pptxgengo.exe catalog --open
.\bin\pptxgengo.exe docs
powershell -NoProfile -ExecutionPolicy Bypass -File .\smoke-test-windows.ps1
```

The smoke test creates a new project in a path containing spaces, searches SQLite,
checks a gallery preview, initializes/splits/builds a two-slide deck, adds a native
section, checks measurements, runs migration dry-run, and exports a client ZIP.
It saves `summary.json`, stdout and stderr logs. Send the result folder back to
the maintainer. Do not use real client content for your first smoke test.

Native export needs **desktop Microsoft PowerPoint** in a signed-in Windows
session and Windows PowerShell 5.1. It uses PowerPoint COM for PDF and PNG export;
contact sheets are produced in Go. It preserves the input PPTX and closes only
its exact task copy, never unrelated presentations. This backend is implemented
and checked with simulated exports, but has not yet passed a real Windows native
qualification run.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\smoke-test-windows.ps1 -Native
```

If COM is blocked by activation, Protected View, application policy, missing fonts
or a pending PowerPoint prompt, core authoring can still be tested. Open the
generated PPTX in PowerPoint and export/review manually. Manual export does not
produce the CLI's signed native review receipt. Inspect the full-size PNGs after
a successful native export; successful automation alone is not visual acceptance.

## What is included

All 649 template definitions, searchable SQLite metadata, reviewed source previews,
icons/logos/core artwork, fonts and documentation are included. The smaller tester
ZIP omits original photograph bytes while keeping searchable photo metadata and
gallery thumbnails. Choosing one of those photos for a slide needs the full ZIP
or the original `West Monroe Photos` directory through `WMDS_BRANDING_ROOT`.
The manifest states `original_photographs_included` explicitly.

Projects from a Mac pin their compiler, OS and architecture. Preserve a copy of
the project, then use `project migrate --dry-run`, `project migrate`, rebuild and
review on Windows. Migration preserves an exact backup of the original lock and
does not rewrite incompatible custom compositions. Keep forward slashes in YAML
project-relative file references; PowerShell CLI arguments can use Windows paths.

Windows x64 and ARM64 binaries are separate. The folder name indicates the target.
The existing library specimen qualification was performed on macOS; new Windows
decks need their own visual review. A Windows CI runner will qualify the native
backend separately from ordinary Go tests.
