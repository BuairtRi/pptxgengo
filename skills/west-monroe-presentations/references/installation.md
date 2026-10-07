# Installation and recovery

Use a CLI containing the new installation manager. The signed `v4.1.0` CLI
predates these commands; check the package help instead of assuming availability.

Run `pptxgengo installation doctor` before repairing a colleague's setup. Read
its selected CLI, resource pins, skill digest, PATH conflicts and pending receipt.
Check `pptxgengo --version` and `pptxgengo paths` too. A CLI-only package omits the
presentation resources and skill; it is not a full presentation installation.

After authenticating the release's signed manifest and archive digest, use
`pptxgengo installation install --from EXTRACTED_PACKAGE` to install or upgrade.
`--stage-only` verifies without activation; `--skip-skill` and `--no-path` leave
those user settings alone. Full Windows packages retain `-SkipFonts` separately.

`installation rollback` selects the recorded previous release and matching skill;
`installation recover` restores the predecessor after an interrupted activation.
Use the original custom `--root`, `--bin-dir` and `--skill-dir` flags when present.
Recovery must preserve conflicting user edits. Never delete a receipt or remove a
lock to make recovery proceed without understanding its recorded predecessors.

Open a new terminal and restart the agent after changing selection. Installer
success does not qualify fonts, PowerPoint or supplied content. Run
`pptxgengo design render-doctor --json` for native readiness and review exports.

`installation uninstall` removes owned activation bindings, restores an unchanged
original skill when recorded, and retains releases/backups. Edited user skills,
fonts, projects and authored decks are preserved. `--keep-skill` preserves skills
explicitly. Do not remove a colleague's projects as part of toolkit maintenance.
