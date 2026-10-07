#!/usr/bin/env bash
set -euo pipefail
mkdir -p .cache/security
case "${1:?usage: scan-source.sh vulnerabilities|secrets}" in
  vulnerabilities)
    GOOS="$(go env GOHOSTOS)" GOARCH="$(go env GOHOSTARCH)" GOBIN="$PWD/.cache/security/bin" go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
    CGO_ENABLED=0 .cache/security/bin/govulncheck -format json ./... > ".cache/security/govulncheck-${GOOS:-host}.json"
    # JSON mode always returns success for findings; inspect reachable findings.
    jq -s -e '[.[] | .finding? // empty | select(.trace[0].function != null)] | length == 0' ".cache/security/govulncheck-${GOOS:-host}.json" >/dev/null
    ;;
  secrets)
    GOOS="$(go env GOHOSTOS)" GOARCH="$(go env GOHOSTARCH)" GOBIN="$PWD/.cache/security/bin" go install github.com/zricethezav/gitleaks/v8@v8.24.3
    .cache/security/bin/gitleaks dir --redact --no-banner --config release/gitleaks.toml --report-format json --report-path .cache/security/gitleaks.json .
    ;;
  *) exit 2 ;;
esac
