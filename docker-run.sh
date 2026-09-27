#!/usr/bin/env bash
set -euo pipefail

cookies_dir="${MR_COOKIES_DIR:-$(pwd)/cookies}"
mkdir -p "$cookies_dir"
cookies_dir="$(cd "$cookies_dir" && pwd)"
docker run -p 3000:3000 \
  -v "$(pwd)/download:/download" \
  -v "$cookies_dir:/app/cookies" \
  media-roller
