#!/usr/bin/env bash
set -euo pipefail
release_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
release_version="${CI_COMMIT_TAG:-${PPTXGENGO_RELEASE_VERSION:-}}"
release_commit="${CI_COMMIT_SHA:-$(git -C "$release_root" rev-parse HEAD)}"
[[ "$release_version" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*)?$ ]] || { echo 'Expected vX.Y.Z or vX.Y.Z-rc.N' >&2; exit 1; }
release_ci() {
  env GOOS="$(go env GOHOSTOS)" GOARCH="$(go env GOHOSTARCH)" CGO_ENABLED=0 go run "$release_root/scripts/cmd/release-ci" "$@"
}
require_release_identity() {
  [[ "${CI_PROJECT_ID:-}" == 17 && "${CI_PROJECT_PATH:-}" == riscott/pptxgengo && "${CI_SERVER_URL:-}" == https://gitlab.samcott.com && "${CI_COMMIT_REF_PROTECTED:-}" == true && "${CI_COMMIT_TAG:-}" == "$release_version" && "$(git -C "$release_root" rev-parse HEAD)" == "$release_commit" ]] || { echo 'Refusing release outside this project and its exact protected tag' >&2; exit 1; }
}
