#!/usr/bin/env bash
set -e
cd "$(dirname -- "${BASH_SOURCE[0]}")"

VERSION="$(<VERSION)"
TS="$(date +%s)"
GIT_HASH="$(git rev-parse HEAD)"

go build -trimpath -a -ldflags "
  -s
  -w
  -X 'feedmash/src.appVersion=$VERSION'
  -X 'feedmash/src.appBuildTimestamp=$TS'
  -X 'feedmash/src.appBuildGitHash=$GIT_HASH'
"
