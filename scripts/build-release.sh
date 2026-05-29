#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Build cross-platform release archives for GitHub Releases.

Usage:
  ./scripts/build-release.sh [options]

Options:
  --version <version>   Binary version to embed. Defaults to util.AppVersion.
  --tag <tag>           Release tag / output folder name. Defaults to v<version>.
  --output <dir>        Output directory. Defaults to ./dist/release.
  --targets <list>      Comma-separated targets, e.g. linux/amd64,windows/amd64.
  --skip-frontend       Skip frontend build.
  --skip-test           Skip go test ./... before packaging.
  --clean               Remove the target release directory before building.
  -h, --help            Show this help message.
EOF
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

first_available_cmd() {
  local cmd
  for cmd in "$@"; do
    if command -v "$cmd" >/dev/null 2>&1; then
      printf '%s\n' "$cmd"
      return 0
    fi
  done
  return 1
}

read_default_version() {
  local version
  version="$(sed -n 's/.*AppVersion *= *"\([^"]*\)".*/\1/p' "${ROOT_DIR}/util/util.go" | head -n 1)"
  if [[ -z "$version" ]]; then
    printf 'failed to read AppVersion from util/util.go\n' >&2
    exit 1
  fi
  printf '%s\n' "$version"
}

create_zip() {
  local archive_path="$1"
  local source_dir="$2"

  if command -v zip >/dev/null 2>&1; then
    (
      cd "$source_dir"
      zip -qr "$archive_path" .
    )
    return 0
  fi

  local python_cmd
  if python_cmd="$(first_available_cmd python3 python)"; then
    "$python_cmd" - "$archive_path" "$source_dir" <<'PY'
import os
import sys
import zipfile

archive_path = sys.argv[1]
source_dir = sys.argv[2]
with zipfile.ZipFile(archive_path, 'w', zipfile.ZIP_DEFLATED) as zf:
    for root, _, files in os.walk(source_dir):
        for file_name in files:
            file_path = os.path.join(root, file_name)
            arc_name = os.path.relpath(file_path, source_dir)
            zf.write(file_path, arc_name)
PY
    return 0
  fi

  if command -v powershell >/dev/null 2>&1; then
    powershell -NoProfile -ExecutionPolicy Bypass -Command \
      "Compress-Archive -Path (Join-Path '${source_dir}' '*') -DestinationPath '${archive_path}' -Force" >/dev/null
    return 0
  fi

  printf 'unable to create zip archive: zip, python, or powershell is required\n' >&2
  exit 1
}

compute_sha256() {
  local file_path="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file_path" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file_path" | awk '{print $1}'
    return 0
  fi
  if command -v certutil >/dev/null 2>&1; then
    certutil -hashfile "$file_path" SHA256 | sed -n '2p' | tr -d ' \r\n'
    return 0
  fi
  printf 'unable to compute sha256: sha256sum, shasum, or certutil is required\n' >&2
  exit 1
}

build_frontend() {
  require_cmd npm
  printf '==> building frontend\n'
  (
    cd "${ROOT_DIR}/frontend"
    if [[ ! -d node_modules ]]; then
      if [[ -f package-lock.json ]]; then
        npm ci
      else
        npm install
      fi
    fi
    npm run build
  )
}

run_tests() {
  require_cmd go
  printf '==> running go test ./...\n'
  (
    cd "${ROOT_DIR}"
    go test ./...
  )
}

build_target() {
  local goos="$1"
  local goarch="$2"
  local artifact_base="my-proxy-${goos}-${goarch}"
  local stage_dir="${STAGE_DIR}/${artifact_base}"
  local binary_name="my-proxy"
  local archive_path

  rm -rf "$stage_dir"
  mkdir -p "$stage_dir"

  if [[ "$goos" == "windows" ]]; then
    binary_name="${binary_name}.exe"
    archive_path="${RELEASE_DIR}/${artifact_base}.zip"
  else
    archive_path="${RELEASE_DIR}/${artifact_base}.tar.gz"
  fi

  printf '==> building %s/%s\n' "$goos" "$goarch"
  (
    cd "${ROOT_DIR}"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
      go build -trimpath -ldflags "-s -w -X github.com/up-zero/my-proxy/util.AppVersion=${VERSION}" \
      -o "${stage_dir}/${binary_name}" .
  )

  cp "${ROOT_DIR}/LICENSE" "$stage_dir/"
  cp "${ROOT_DIR}/README.md" "$stage_dir/"
  cp "${ROOT_DIR}/README_zh.md" "$stage_dir/"

  rm -f "$archive_path"
  if [[ "$goos" == "windows" ]]; then
    create_zip "$archive_path" "$stage_dir"
  else
    tar -C "$stage_dir" -czf "$archive_path" .
  fi

  ARCHIVES+=("$archive_path")
}

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
DEFAULT_VERSION="$(read_default_version)"
VERSION="${DEFAULT_VERSION#v}"
TAG="v${VERSION}"
TAG_EXPLICIT=0
OUTPUT_DIR="${ROOT_DIR}/dist/release"
SKIP_FRONTEND=0
SKIP_TEST=0
CLEAN=0
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2#v}"
      shift 2
      ;;
    --tag)
      TAG="$2"
      TAG_EXPLICIT=1
      shift 2
      ;;
    --output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --targets)
      IFS=',' read -r -a TARGETS <<< "$2"
      shift 2
      ;;
    --skip-frontend)
      SKIP_FRONTEND=1
      shift
      ;;
    --skip-test)
      SKIP_TEST=1
      shift
      ;;
    --clean)
      CLEAN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown option: %s\n\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  printf 'version must not be empty\n' >&2
  exit 1
fi

if [[ "$TAG_EXPLICIT" -eq 0 ]]; then
  TAG="v${VERSION}"
fi

require_cmd go
require_cmd tar

RELEASE_DIR="${OUTPUT_DIR}/${TAG}"
STAGE_DIR="${RELEASE_DIR}/.stage"
CHECKSUM_FILE="${RELEASE_DIR}/checksums.txt"
ARCHIVES=()

if [[ "$CLEAN" -eq 1 && -d "$RELEASE_DIR" ]]; then
  rm -rf "$RELEASE_DIR"
fi
mkdir -p "$RELEASE_DIR" "$STAGE_DIR"

if [[ "$SKIP_FRONTEND" -eq 0 ]]; then
  build_frontend
fi

if [[ "$SKIP_TEST" -eq 0 ]]; then
  run_tests
fi

for target in "${TARGETS[@]}"; do
  if [[ "$target" != */* ]]; then
    printf 'invalid target format: %s\n' "$target" >&2
    exit 1
  fi
  build_target "${target%/*}" "${target#*/}"
done

: > "$CHECKSUM_FILE"
for archive_path in "${ARCHIVES[@]}"; do
  printf '%s  %s\n' "$(compute_sha256 "$archive_path")" "$(basename "$archive_path")" >> "$CHECKSUM_FILE"
done

printf '\nrelease artifacts are ready in %s\n' "$RELEASE_DIR"
printf 'upload these files to the GitHub Release tagged %s:\n' "$TAG"
for archive_path in "${ARCHIVES[@]}"; do
  printf '  - %s\n' "$(basename "$archive_path")"
done
printf '  - %s\n' "$(basename "$CHECKSUM_FILE")"
