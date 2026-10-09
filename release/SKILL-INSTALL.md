# Install or update the presentation skill

After authenticating the private GitLab archive against its signed release manifest,
extract it and run these commands from the extracted package root:

```sh
python3 scripts/install-skill.py --agent both --dry-run
python3 scripts/install-skill.py --agent both
```

On Windows use `py -3` instead of `python3`. Select `--agent codex` or
`--agent claude` to update one agent. The helper copies the packaged skill and
all references, verifies the staged copy, and preserves an existing installation
outside skill discovery roots. It prints the selected destinations and backup.
Rerunning the same payload is a no-op. No network access or extra Python package
is required. Restart the agent to discover the updated skill.

For a deck-specific installation, add `--project /path/to/deck-project`. Use
`--dest` only with one selected agent when choosing an explicit existing skill
folder. If both modern and legacy Codex locations exist, choose the active one
instead of creating a duplicate. CLI installation and project toolchain migration
are separate operations; a skill update does not migrate authored decks.

Read `skills/west-monroe-presentations/references/installation.md` for CLI
installation, recovery, backups and version selection. Keep projects and edited
decks outside the extracted toolkit directory.
