#!/usr/bin/env bash
# Merge one upstream release tag onto a sync branch and open a draft PR.
# A clean empty merge or a tree that still points at upstream stops here.
# A conflict still opens a draft PR pointed at the tag so GitHub shows it.
set -euo pipefail

tag="${1:?release tag required}"
upstream="${UPSTREAM_REPO:-MHSanaei/3x-ui}"
upstream_url="${UPSTREAM_GIT_URL:-https://github.com/${upstream}.git}"
base="${BASE_BRANCH:-main}"
branch="sync/upstream-${tag}"

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

if git ls-remote --exit-code --heads origin "$branch" >/dev/null 2>&1; then
  echo "branch ${branch} already exists; not opening another PR" >&2
  exit 0
fi

git fetch --no-tags "$upstream_url" "refs/tags/${tag}:refs/tags/upstream-sync"
git checkout -B "$branch"

conflicted=""
if ! git merge --no-edit refs/tags/upstream-sync; then
  conflicted=$(git diff --name-only --diff-filter=U)
  if [[ -z "$conflicted" ]]; then
    conflicted="(git could not complete the merge)"
  fi
  echo "conflict merging ${upstream}@${tag}; opening a draft PR for: ${conflicted}" >&2
  git merge --abort || true
  git checkout -B "$branch" refs/tags/upstream-sync
fi

title="sync: merge upstream ${tag}"
body="$(cat <<EOF
合并上游 \`${upstream}\` 的 release \`${tag}\`。

这个分支从默认分支拉出后再合并该 tag。默认分支上的去广告和自更新指向应当还在；若上面的检查或 CI 失败，不要合并。

不自动合并。
EOF
)"

if [[ -n "$conflicted" ]]; then
  title="sync: merge upstream ${tag} (conflicts)"
  body="$(cat <<EOF
合并上游 \`${upstream}\` 的 release \`${tag}\` 时本地产生冲突。

这个 PR 的分支指向该 tag 本身，GitHub 会标出与默认分支的冲突。冲突文件：

${conflicted}

请在 PR 里解决冲突，并保留本仓库的去广告和自更新指向。不要自动合并。
EOF
)"
else
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
fi

git push -u origin "$branch"
gh pr create --draft --base "$base" --head "$branch" \
  --title "$title" \
  --body "$body"
