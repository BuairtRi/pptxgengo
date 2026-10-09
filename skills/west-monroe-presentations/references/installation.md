# Installation and recovery

## Skill setup for Codex and Claude Code

From an authenticated source checkout or an extracted v4.3.0 archive, run
`python3 scripts/install-skill.py --agent both` (Windows: `py -3`). Select
`--agent codex` or `--agent claude` for one agent; rerun after updating the checkout
to update the skill. The helper copies the whole reference folder, preserves the
previous installation and reports destinations/backups. Use `--dry-run` first
when inspecting an existing setup; `--project PATH` installs for that project.
Restart the agent and invoke `west-monroe-presentations` (Claude Code:
`/west-monroe-presentations`). CLI and project pins are unchanged.

v4.3.0 archives include the complete skill, helper and `SKILL-INSTALL.md`;
v4.2.1 archives require the source checkout for these materials. If given the skill folder alone, copy
the complete folder to `~/.agents/skills/west-monroe-presentations` for Codex or
`~/.claude/skills/west-monroe-presentations` for Claude Code. Preserve the previous
folder outside agent discovery roots before replacement. For an older Codex
installation using `CODEX_HOME/skills`, the helper reuses the existing folder
when the current location is absent. If both locations exist, select the active
one with `--dest`; avoid two copies of the same skill.

## CLI installation and recovery

Check `pptxgengo --version`, `pptxgengo paths` and `pptxgengo installation doctor`
before changing an installation. Read the selected executable, resource pins,
PATH conflicts and pending receipt. Paths can describe optional files that are
not installed; check availability before copying an example or opening a site.

The current template archives contain three executables, the V11 source bundle,
SQLite index, specimen gallery, fonts and `browsing/template-library.pptx`.
The v4.3.0 archive also carries skill installation/update materials and selected
registered SVG diagram icons/arrows. The upstream docs site, external private
media originals and curated reusable-slide inventory are separate resources. Font files in an archive are not proof that
PowerPoint has those fonts installed.

After authenticating the private GitLab release's signed manifest and archive
digest, use `pptxgengo installation install --from EXTRACTED_PACKAGE`.
`--stage-only` verifies and copies without activation; `--skip-skill` and
`--no-path` preserve those settings. Custom `--root`, `--bin-dir` and `--skill-dir`
locations must be separate; retain the same flags for later recovery/rollback.
Use the installed command's help for platform-specific choices.

`installation rollback` selects the recorded predecessor; `installation recover`
restores it after interrupted activation. Preserve conflicting edits and inspect
the receipt before retrying. Restart the terminal and agent after changing the
active release. A skill-only documentation update does not require project
migration; changing the executable or bundle can require explicit migration.

Run `pptxgengo design render-doctor --json` for native readiness. Resolve the
reported font, activation or folder-access problem and review actual exports.
On macOS use the same PowerPoint-accessible staging folder for doctor and render.
On Windows follow [Windows workflows](windows.md).

`installation uninstall` removes owned activation bindings and retains releases
and backups. Edited skills, fonts and projects are preserved; `--keep-skill`
explicitly retains skills. Keep authored decks outside toolkit cleanup.
