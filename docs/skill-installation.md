# Install or update the presentation skill

`scripts/install-skill.py` copies the complete `skills/west-monroe-presentations`
folder from this checkout into Codex or Claude Code's normal skill discovery
folder. It uses Python's standard library; it does not fetch code, ask for a
credential, install a CLI, or change the CLI release selected by a project.

Obtain this repository through the authenticated private GitLab checkout you
already use. From the checkout root, fast-forward to the reviewed revision, then
run the installer again:

```sh
git pull --ff-only
python3 scripts/install-skill.py --agent codex
```

The source skill folder is the update payload; the helper never runs `git pull`
or accesses the network.

## Install globally

From the repository root:

```sh
python3 scripts/install-skill.py --agent codex --dry-run
python3 scripts/install-skill.py --agent codex
```

A new Codex install goes to `~/.agents/skills/west-monroe-presentations`. If that
folder is absent but an existing skill is at `${CODEX_HOME:-$HOME/.codex}/skills/`,
the helper updates that existing legacy location instead of creating a duplicate.
If both Codex locations already contain the skill, the helper stops and asks you
to choose one with `--dest`. To install for Claude Code instead:

```sh
python3 scripts/install-skill.py --agent claude
```

On Windows, use `py -3 scripts\install-skill.py --agent codex` or
`py -3 scripts\install-skill.py --agent claude` from PowerShell.

That destination is `~/.claude/skills/west-monroe-presentations`. Use
`--agent both` to install for Codex and Claude Code. Codex still selects only one
Codex discovery location; this option does not duplicate the skill into both
current and legacy Codex roots.

## Install for one project

```sh
python3 scripts/install-skill.py --agent codex --project /path/to/project
python3 scripts/install-skill.py --agent claude --project /path/to/project
```

On Windows, use `py -3 scripts\install-skill.py --agent codex --project C:\Work\deck`
or change `codex` to `claude`.

Project installs go into `.agents/skills/west-monroe-presentations` for Codex
or `.claude/skills/west-monroe-presentations` for Claude Code, under the project
root. Pass `--agent both` to install into both project locations. Restart or
refresh the agent session if it does not discover the new skill immediately.

## Choose another destination

Use `--dest` for a single agent when an existing setup intentionally uses another
skill root. For a legacy Codex installation, for example:

```sh
python3 scripts/install-skill.py --agent codex \
  --dest "${CODEX_HOME:-$HOME/.codex}/skills/west-monroe-presentations"
```

A new Codex install follows the current `~/.agents/skills` discovery location.
An existing legacy `$CODEX_HOME/skills` destination is reused when the current
location is absent; the helper never creates a new legacy copy implicitly.

## Update, backup and rollback

The helper compares relative paths and file bytes. An identical install is a
no-op. A changed install is copied to a staging folder beside the destination
and verified before same-parent renames swap it in; it never copies files over
the live skill in place. Existing contents are saved before the
replacement under `${XDG_DATA_HOME:-$HOME/.local/share}/pptxgengo/skill-backups`.
Set `--backup-dir` to choose another location; it must be outside the source,
the skill discovery roots and the destination. The helper prints the backup
location on update.

```sh
python3 scripts/install-skill.py --agent codex --backup-dir "$HOME/skill-backups"
```

To roll back manually, move the current skill folder outside the discovery root,
then copy the chosen saved folder back. For example:

```sh
backup="$HOME/.local/share/pptxgengo/skill-backups"
mkdir -p "$backup"
mv "$HOME/.agents/skills/west-monroe-presentations" \
  "$backup/west-monroe-presentations.failed-<unique-name>"
cp -a "$backup/<saved-folder>" \
  "$HOME/.agents/skills/west-monroe-presentations"
```

For Windows PowerShell, use the matching paths under `$HOME\.agents\skills` or
`$HOME\.claude\skills`, and the backup path printed by the installer. Move the
current folder to a name outside the agent's `skills` folder, then copy the saved
folder back with `Copy-Item -Recurse`.

Use the matching `.claude/skills` or project destination for those installs.
Keep the saved folder so rollback remains repeatable. A saved top-level symlink
is preserved as a symlink pointing to the same target. Replacing a symlink never
writes into its target.

A per-destination lock prevents two updates from replacing the same skill at
once. If interrupted execution leaves an `.install-lock` folder, first confirm
that no installer is running, then remove that lock folder and rerun. Failed
updates restore the previous destination; any external backup already written
is retained.

## Safety boundaries

- The source must be a real folder containing a `SKILL.md` that declares the
`west-monroe-presentations` skill name; source symlinks are refused.
- Source/destination overlaps, resolved destination aliases and backup paths inside skill discovery roots are refused. Choose one destination explicitly if Codex and Claude currently share a linked folder.
- The full skill folder is copied. No unrelated skills or agent settings are changed.
- `--dry-run` reports install, update or no-op without writing files.
- A symlink at the destination is backed up as a symlink; its target is never mutated.
- The archive or GitLab checkout should be authenticated and reviewed before use.

Codex discovers user skills from `~/.agents/skills` and project skills from
`.agents/skills`; see the [Codex skills guide](https://learn.chatgpt.com/docs/build-skills).
Claude Code uses `~/.claude/skills` and project `.claude/skills`; see the
[Claude Code skills guide](https://code.claude.com/docs/en/skills).
