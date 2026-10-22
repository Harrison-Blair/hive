#!/usr/bin/env bash
# Build reproducible Linux release archives. Usage: package.sh vX.Y.Z[-rc.N] DIST
# Build from the tag commit: the archive time comes from HEAD.
set -euo pipefail

tag=${1:?release tag required}
dist=${2:?output directory required}
if [[ ! "$tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-rc\.(0|[1-9][0-9]*))?$ ]]; then
  echo "Invalid release tag: $tag" >&2
  exit 1
fi
mkdir -p "$dist"
dist=$(cd "$dist" && pwd)
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
module=$(go list -m)
timestamp=$(git show -s --format=%ct HEAD)

for arch in amd64 arm64; do
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -buildvcs=true \
    -ldflags "-s -w -X $module/internal/lib/version.Version=$tag" -o "$stage/hive" .
  cp LICENSE "$stage/"
  # Run the archive's binary on a matching Linux build host.
  if [[ "$(go env GOHOSTOS)/$(go env GOHOSTARCH)" == "linux/$arch" ]]; then
    version=$("$stage/hive" version)
    if [[ "$version" != "hive $tag ("* ]]; then
      echo "hive version printed \"$version\", want hive $tag" >&2
      exit 1
    fi
  fi
  tar --sort=name --mtime="@$timestamp" --owner=0 --group=0 --numeric-owner --mode=a+rX,go-w \
    -cf - -C "$stage" LICENSE hive | gzip -n -9 > "$dist/hive_${tag}_linux_${arch}.tar.gz"
done
(
  cd "$dist"
  sha256sum "hive_${tag}_linux_amd64.tar.gz" "hive_${tag}_linux_arm64.tar.gz" > checksums.txt
  sha256sum --check checksums.txt
)
