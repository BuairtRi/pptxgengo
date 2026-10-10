# Release runbook

Protected release tags are built and published by private GitLab project 17.
GitHub remains the source and review mirror; GitHub Actions and GitHub release
assets are disabled.

For current package scope, signing gates, resource inventory, verification
commands and tag procedure, use the [private GitLab release runbook](release/CI.md).
For PR, main and nightly job selection, use [CI cadence](docs/ci-cadence.md).

The v4.3.1 source selects V12 (735 templates). Start the release from a reviewed
`main` commit, choose a new immutable `vX.Y.Z` or `vX.Y.Z-rc.N` tag, and follow
the runbook. Never move or reuse an existing release tag.

The inherited Node/npm checklist is retained in
[upstream release history](docs/upstream-releasing.md).
