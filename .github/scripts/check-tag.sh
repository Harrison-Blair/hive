#!/usr/bin/env bash
# Check a release tag. Usage: check-tag.sh vX.Y.Z[-rc.N]
# Prints "release" or "prerelease". Exits 1 with a reason for a refused tag.
set -euo pipefail

tag=${1:?release tag required}
refuse() {
  echo "Refused tag $tag: $1" >&2
  exit 1
}

if [[ ! "$tag" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)(-rc\.([0-9]+))?$ ]]; then
  refuse "the name does not match vX.Y.Z or vX.Y.Z-rc.N"
fi
numbers=("${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}" "${BASH_REMATCH[5]}")
suffix=${BASH_REMATCH[4]}
for number in "${numbers[@]}"; do
  if [[ "$number" =~ ^0[0-9] ]]; then
    refuse "a number has a leading zero"
  fi
done
kind=release
if [[ -n "$suffix" ]]; then
  kind=prerelease
fi

# Fetch the tag object and main again. A checkout can keep a tag as a lightweight ref.
if ! git fetch --quiet --force origin "+refs/heads/main:refs/remotes/origin/main" "+refs/tags/$tag:refs/tags/$tag"; then
  refuse "the tag or main was not found on origin"
fi
if [[ "$(git cat-file -t "refs/tags/$tag")" != tag ]]; then
  refuse "it is a lightweight tag; push an annotated tag"
fi
if ! git merge-base --is-ancestor "refs/tags/$tag^{commit}" refs/remotes/origin/main; then
  refuse "the tag commit is not on origin/main"
fi
echo "$kind"
