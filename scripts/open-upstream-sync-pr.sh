#!/usr/bin/env bash
# Merge one upstream release tag onto a sync branch and open a draft PR.
# A conflict, an empty merge, or a tree that still points at upstream stops here.
set -euo pipefail

tag="${1:?release tag required}"
upstream="${UPSTREAM_REPO:-MHSanaei/3x-ui}"
base="${BASE_BRANCH:-main}"
branch="sync/upstream-${tag}"

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

if git ls-remote --exit-code --heads origin "$branch" >/dev/null 2>&1; then
  echo "branch ${branch} already exists; not opening another PR" >&2
  exit 0
fi

git fetch --no-tags "https://github.com/${upstream}.git" "refs/tags/${tag}:refs/tags/upstream-sync"
git checkout -B "$branch"
if ! git merge --no-edit refs/tags/upstream-sync; then
  echo "conflict merging ${upstream}@${tag}; stopping without a PR" >&2
  git merge --abort || true
  exit 1
fi

if git diff --quiet "${base}...HEAD"; then
  echo "merge of ${tag} produced no changes; not opening a PR" >&2
  exit 0
fi

forbidden=$(mktemp)
trap 'rm -f "$forbidden"' EXIT
found=0
if git grep -n -E 'donate\.sanaei\.dev|t\.me/XrayUI|MHSanaei/sponsors|ghcr\.io/mhsanaei/3x-ui|hsanaeii/3x-ui' -- . >"$forbidden"; then
  found=1
fi
for f in install.sh update.sh x-ui.sh internal/web/service/panel/panel.go docs/lib/shared.ts docs/lib/xray/install.ts frontend/src/layouts/AppSidebar.tsx; do
  if git grep -n -E 'MHSanaei/3x-ui|mhsanaei/3x-ui' -- "$f" >>"$forbidden"; then
    found=1
  fi
done
if [[ "$found" -eq 1 ]]; then
  echo "merged tree still has upstream promo or release URLs; not opening a PR" >&2
  cat "$forbidden" >&2
  exit 1
fi

git push -u origin "$branch"
gh pr create --draft --base "$base" --head "$branch" \
  --title "sync: merge upstream ${tag}" \
  --body "$(cat <<EOF
合并上游 \`${upstream}\` 的 release \`${tag}\`。

这个分支从默认分支拉出后再合并该 tag。默认分支上的去广告和自更新指向应当还在；若上面的检查或 CI 失败，不要合并。

不自动合并。
EOF
)"
