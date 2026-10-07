# Portable projects and packaged browsing decks: scope record

This file recorded Ri's initial 2026-10-07 request before the follow-up choices.
The authoritative maintained requirements, implemented command contracts and
acceptance evidence are now in
[portable projects and packaged browsing decks](../portable-projects.md).

Accepted follow-up choices:

- Numbered versions retain complete source and generated PowerPoint snapshots.
  All versions share one deck-owned immutable asset store; each changed asset
  has its own exact hash revision. Unchanged image bytes are stored once in the
  colleague ZIP, with ordinary relative files and explicit legacy path aliases.
- The reusable-slide browsing PowerPoint includes only the latest approved,
  nonexpired revision per identity, respecting newer withdrawals/deprecation.
  It does not show every retained revision or resurrect an older approval.
- Private GitLab CI generates release browsing decks. All CLI CI jobs run in
  GitLab; GitHub remains a source/merge mirror and hosts no release artifacts.

The canonical document distinguishes implemented project commands from remaining
browsing-deck release/native qualification. This historical record is a redirect,
not a separate or conflicting implementation contract.
