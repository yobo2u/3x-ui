#!/usr/bin/env bash
# Print one unsynced stable upstream tag, or nothing. Never opens a PR.
set -euo pipefail

upstream="${UPSTREAM_REPO:-MHSanaei/3x-ui}"
limit="${UPSTREAM_RELEASE_LIMIT:-100}"

if ! command -v gh >/dev/null || ! command -v jq >/dev/null; then
  echo "upstream-sync: gh and jq are required" >&2
  exit 1
fi

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

git fetch --no-tags "https://github.com/${upstream}.git" '+refs/tags/*:refs/tags/upstream/*'

rows=$(mktemp)
trap 'rm -f "$rows"' EXIT

gh release list --repo "$upstream" --limit "$limit" --json tagName,isDraft,isPrerelease \
  | jq -r '.[] | [.tagName, (.isDraft|tostring), (.isPrerelease|tostring)] | @tsv' \
  | while IFS=$'\t' read -r tag draft pre; do
      sha=""
      if git rev-parse -q --verify "refs/tags/upstream/${tag}^{commit}" >/dev/null; then
        sha=$(git rev-parse "refs/tags/upstream/${tag}^{commit}")
      fi
      ancestor=false
      if [[ -n "$sha" ]] && git merge-base --is-ancestor "$sha" HEAD; then
        ancestor=true
      fi
      exists=false
      if git ls-remote --exit-code --heads origin "sync/upstream-${tag}" >/dev/null 2>&1; then
        exists=true
      fi
      printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$tag" "$sha" "$draft" "$pre" "$ancestor" "$exists"
    done >"$rows"

tag=$(go run ./tools/upstreamsync <"$rows")
printf '%s' "$tag"
