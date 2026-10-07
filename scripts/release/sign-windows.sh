#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/common.sh"
require_release_identity
target="${1:?usage: sign-windows.sh windows-ARCH}"
[[ "$target" == windows-amd64 || "$target" == windows-arm64 ]] || exit 2
[[ "$(uname -s)" == Linux ]] || exit 1
input="dist/unsigned/$target"; output="dist/signed/$target"
release_ci verify-binaries "$input" "$target" "$release_version" "$release_commit"
[[ ! -e "$output" ]] || exit 1
for name in AZURE_FEDERATED_TOKEN AZURE_CLIENT_ID AZURE_TENANT_ID AZURE_SUBSCRIPTION_ID AZURE_ARTIFACT_SIGNING_ENDPOINT AZURE_ARTIFACT_SIGNING_ACCOUNT_NAME AZURE_ARTIFACT_SIGNING_CERTIFICATE_PROFILE_NAME; do
  [[ -n "${!name:-}" ]] || { echo "Missing protected signing variable $name" >&2; exit 1; }
done
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
# Azure CLI credentials live only in a job-private directory.
export AZURE_CONFIG_DIR="$tmp/azure"
az login --service-principal --username "$AZURE_CLIENT_ID" --tenant "$AZURE_TENANT_ID" --federated-token "$AZURE_FEDERATED_TOKEN" --allow-no-subscriptions --output none
az account set --subscription "$AZURE_SUBSCRIPTION_ID"
signer_token="$(az account get-access-token --resource https://codesigning.azure.net --query accessToken -o tsv)"
[[ -n "$signer_token" ]] || exit 1
bash scripts/release/install-tools.sh --bin-dir "$tmp/tools" --only jsign
bash scripts/release/install-windows-ca.sh --bundle-out "$tmp/ca.pem"
mkdir -p "$output/bin"
cp "$input"/bin/*.exe "$output/bin/"
endpoint="${AZURE_ARTIFACT_SIGNING_ENDPOINT#https://}"; endpoint="${endpoint%/}"
for binary in "$output"/bin/*.exe; do
  java -jar "$tmp/tools/jsign.jar" sign --storetype TRUSTEDSIGNING --keystore "$endpoint" --storepass "$signer_token" \
    --alias "$AZURE_ARTIFACT_SIGNING_ACCOUNT_NAME/$AZURE_ARTIFACT_SIGNING_CERTIFICATE_PROFILE_NAME" \
    --alg SHA-256 --name pptxgengo --url https://gitlab.samcott.com/riscott/pptxgengo --replace "$binary"
  bash scripts/release/verify-windows.sh "$binary" "$tmp/ca.pem"
done
unset signer_token
release_ci evidence "$output" "$target" "$release_version" "$release_commit" authenticode "$input/build-evidence.json"
