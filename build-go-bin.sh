#!/usr/bin/env sh
set -eu
NOM=squeeze-empty-lines
CGO_ENABLED=0 \
go build \
  -trimpath \
  -ldflags="-s -w" \
  -o "$NOM" .
if command -v strip >/dev/null 2>&1; then
  strip "$NOM" || true
fi
