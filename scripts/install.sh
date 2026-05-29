#!/usr/bin/env bash
set -euo pipefail

APP_NAME="my-proxy"
REPO="${MY_PROXY_REPO:-up-zero/my-proxy}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${MY_PROXY_VERSION:-latest}"
DOWNLOAD_BASE_URL="${MY_PROXY_DOWNLOAD_BASE_URL:-}"
SERVICE_NAME="${MY_PROXY_SERVICE_NAME:-my-proxy}"
SERVICE_LABEL="${MY_PROXY_SERVICE_LABEL:-com.up-zero.my-proxy}"
SERVICE_PORT="${MY_PROXY_SERVICE_PORT:-12312}"
SERVICE_SCOPE="${MY_PROXY_SERVICE_SCOPE:-auto}"
SYSTEMD_DIR="${MY_PROXY_SYSTEMD_DIR:-/etc/systemd/system}"
LAUNCHD_SYSTEM_DIR="${MY_PROXY_LAUNCHD_SYSTEM_DIR:-/Library/LaunchDaemons}"
LAUNCHD_USER_DIR="${MY_PROXY_LAUNCHD_USER_DIR:-$HOME/Library/LaunchAgents}"
USE_SUDO=1
INSTALL_SERVICE=1

usage() {
  cat <<'EOF'
Install my-proxy on Linux or macOS from GitHub Releases.

Usage:
  ./scripts/install.sh [options]

Options:
  --version <version>      Install a specific release tag or version. Defaults to latest.
  --install-dir <dir>      Install directory. Defaults to /usr/local/bin.
  --repo <owner/name>      GitHub repository. Defaults to up-zero/my-proxy.
  --service-port <port>    Service port passed to `my-proxy serve`. Defaults to 12312.
  --service-scope <scope>  Service scope: auto, system, user. Defaults to auto.
                           Linux supports system only. macOS auto chooses LaunchDaemon
                           when root/sudo is available, otherwise LaunchAgent.
  --skip-service           Only install the binary, do not configure auto-start service.
  --no-sudo                Do not use sudo even when privileged operations are required.
  -h, --help               Show this help message.

Environment variables:
  MY_PROXY_VERSION         Same as --version.
  INSTALL_DIR              Same as --install-dir.
  MY_PROXY_REPO            Same as --repo.
  MY_PROXY_DOWNLOAD_BASE_URL
                            Override the release asset base URL.
  MY_PROXY_SERVICE_PORT    Same as --service-port.
  MY_PROXY_SERVICE_SCOPE   Same as --service-scope.
EOF
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

run_privileged() {
  if [[ "$EUID" -eq 0 ]]; then
    "$@"
    return 0
  fi

  if [[ "$USE_SUDO" -eq 1 ]]; then
    need_cmd sudo
    sudo "$@"
    return 0
  fi

  printf 'this operation requires root privileges; rerun without --no-sudo or use --skip-service\n' >&2
  exit 1
}

needs_sudo_for_dir() {
  local dir_path="$1"
  if [[ -d "$dir_path" ]]; then
    [[ ! -w "$dir_path" ]]
    return $?
  fi

  local parent_dir
  parent_dir="$(dirname "$dir_path")"
  while [[ ! -d "$parent_dir" && "$parent_dir" != "/" ]]; do
    parent_dir="$(dirname "$parent_dir")"
  done

  [[ ! -w "$parent_dir" ]]
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64\n' ;;
    aarch64|arm64) printf 'arm64\n' ;;
    *)
      printf 'unsupported architecture: %s\n' "$(uname -m)" >&2
      exit 1
      ;;
  esac
}

detect_os() {
  case "$(uname -s)" in
    Linux) printf 'linux\n' ;;
    Darwin) printf 'darwin\n' ;;
    *)
      printf 'unsupported operating system: %s\n' "$(uname -s)" >&2
      exit 1
      ;;
  esac
}

xml_escape() {
  local value="$1"
  value="${value//&/&amp;}"
  value="${value//</&lt;}"
  value="${value//>/&gt;}"
  value="${value//\"/&quot;}"
  value="${value//\'/&apos;}"
  printf '%s\n' "$value"
}

download() {
  local url="$1"
  local output="$2"
  if has_cmd curl; then
    curl -fsSL "$url" -o "$output"
    return $?
  fi
  if has_cmd wget; then
    wget -qO "$output" "$url"
    return $?
  fi
  printf 'curl or wget is required to download release assets\n' >&2
  exit 1
}

url_exists() {
  local url="$1"
  if has_cmd curl; then
    curl -fsI "$url" >/dev/null 2>&1
    return $?
  fi
  if has_cmd wget; then
    wget --spider -q "$url"
    return $?
  fi
  return 1
}

verify_checksum() {
  local archive_path="$1"
  local checksum_path="$2"
  local archive_name
  archive_name="$(basename "$archive_path")"

  if [[ ! -f "$checksum_path" ]]; then
    printf 'warning: checksums.txt not found, skipping checksum verification\n' >&2
    return 0
  fi

  local verify_cmd=()
  if has_cmd sha256sum; then
    verify_cmd=(sha256sum -c -)
  elif has_cmd shasum; then
    verify_cmd=(shasum -a 256 -c -)
  else
    printf 'warning: sha256sum/shasum not found, skipping checksum verification\n' >&2
    return 0
  fi

  local checksum_line
  checksum_line="$(awk -v name="$archive_name" 'NF >= 2 && $NF == name { print $1 "  " $NF; exit }' "$checksum_path")"
  if [[ -z "$checksum_line" ]]; then
    printf 'warning: checksum for %s not found, skipping checksum verification\n' "$archive_name" >&2
    return 0
  fi

  printf '%s\n' "$checksum_line" | (cd "$(dirname "$archive_path")" && "${verify_cmd[@]}")
}

resolve_service_scope() {
  local requested="$1"

  case "$requested" in
    auto)
      case "$OS" in
        linux)
          printf 'system\n'
          ;;
        darwin)
          if [[ "$EUID" -eq 0 ]]; then
            printf 'system\n'
          elif [[ "$USE_SUDO" -eq 1 ]] && has_cmd sudo; then
            printf 'system\n'
          else
            printf 'user\n'
          fi
          ;;
      esac
      ;;
    system|user)
      if [[ "$OS" == 'linux' && "$requested" == 'user' ]]; then
        printf 'Linux only supports --service-scope system\n' >&2
        exit 1
      fi
      printf '%s\n' "$requested"
      ;;
    *)
      printf 'invalid service scope: %s\n' "$requested" >&2
      exit 1
      ;;
  esac
}

install_linux_service() {
  need_cmd systemctl

  local service_file="${SYSTEMD_DIR}/${SERVICE_NAME}.service"
  local temp_file="${TMP_DIR}/${SERVICE_NAME}.service"

  cat > "$temp_file" <<EOF
[Unit]
Description=my-proxy service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=${INSTALL_PATH} serve --port ${SERVICE_PORT}
Restart=always
RestartSec=3
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF

  printf '==> installing systemd service %s\n' "$service_file"
  run_privileged mkdir -p "$SYSTEMD_DIR"
  run_privileged install -m 0644 "$temp_file" "$service_file"

  printf '==> enabling and starting %s\n' "${SERVICE_NAME}.service"
  run_privileged systemctl daemon-reload
  run_privileged systemctl enable --now "${SERVICE_NAME}.service"
}

install_darwin_service() {
  need_cmd launchctl

  local scope="$1"
  local service_dir service_path domain home_dir plist_file label
  label="$SERVICE_LABEL"
  plist_file="${TMP_DIR}/${label}.plist"

  if [[ "$scope" == 'system' ]]; then
    service_dir="$LAUNCHD_SYSTEM_DIR"
    service_path="${service_dir}/${label}.plist"
    domain='system'
    home_dir='/var/root'
  else
    service_dir="$LAUNCHD_USER_DIR"
    service_path="${service_dir}/${label}.plist"
    domain="gui/$(id -u)"
    home_dir="$HOME"
  fi

  cat > "$plist_file" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>$(xml_escape "$label")</string>
  <key>ProgramArguments</key>
  <array>
    <string>$(xml_escape "$INSTALL_PATH")</string>
    <string>serve</string>
    <string>--port</string>
    <string>$(xml_escape "$SERVICE_PORT")</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>EnvironmentVariables</key>
  <dict>
    <key>HOME</key>
    <string>$(xml_escape "$home_dir")</string>
  </dict>
</dict>
</plist>
EOF

  printf '==> installing launchd service %s\n' "$service_path"
  if [[ "$scope" == 'system' ]]; then
    run_privileged mkdir -p "$service_dir"
    run_privileged install -m 0644 "$plist_file" "$service_path"
    run_privileged chown root:wheel "$service_path"
    run_privileged chmod 0644 "$service_path"
    run_privileged launchctl bootout "$domain" "$service_path" >/dev/null 2>&1 || true
    run_privileged launchctl bootstrap "$domain" "$service_path"
    run_privileged launchctl enable "${domain}/${label}"
    run_privileged launchctl kickstart -k "${domain}/${label}"
  else
    mkdir -p "$service_dir"
    install -m 0644 "$plist_file" "$service_path"
    launchctl bootout "$domain" "$service_path" >/dev/null 2>&1 || true
    launchctl bootstrap "$domain" "$service_path"
    launchctl enable "${domain}/${label}"
    launchctl kickstart -k "${domain}/${label}"
  fi
}

install_service() {
  case "$OS" in
    linux)
      install_linux_service
      ;;
    darwin)
      install_darwin_service "$RESOLVED_SERVICE_SCOPE"
      ;;
  esac
}

resolve_base_url() {
  local requested="$1"
  local exact="$requested"
  local stripped="${requested#v}"
  local prefixed="v${stripped}"
  local candidate

  if [[ -n "$DOWNLOAD_BASE_URL" ]]; then
    printf '%s\n' "${DOWNLOAD_BASE_URL%/}"
    return 0
  fi

  if [[ "$requested" == "latest" ]]; then
    printf 'https://github.com/%s/releases/latest/download\n' "$REPO"
    return 0
  fi

  for candidate in "$exact" "$prefixed" "$stripped"; do
    local base_url="https://github.com/${REPO}/releases/download/${candidate}"
    if url_exists "${base_url}/${ARCHIVE_NAME}"; then
      printf '%s\n' "$base_url"
      return 0
    fi
  done

  printf 'unable to locate release asset for version/tag: %s\n' "$requested" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --install-dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    --repo)
      REPO="$2"
      shift 2
      ;;
    --service-port)
      SERVICE_PORT="$2"
      shift 2
      ;;
    --service-scope)
      SERVICE_SCOPE="$2"
      shift 2
      ;;
    --skip-service)
      INSTALL_SERVICE=0
      shift
      ;;
    --no-sudo)
      USE_SUDO=0
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

OS="$(detect_os)"

if [[ -z "$INSTALL_DIR" ]]; then
  printf 'install dir must not be empty\n' >&2
  exit 1
fi

if [[ ! "$SERVICE_PORT" =~ ^[0-9]+$ ]] || (( SERVICE_PORT < 1 || SERVICE_PORT > 65535 )); then
  printf 'service port must be an integer between 1 and 65535\n' >&2
  exit 1
fi

need_cmd tar
need_cmd install
ARCH="$(detect_arch)"
ARCHIVE_NAME="${APP_NAME}-${OS}-${ARCH}.tar.gz"
CHECKSUM_NAME="checksums.txt"
BASE_URL="$(resolve_base_url "$VERSION")"
TMP_DIR="$(mktemp -d)"
ARCHIVE_PATH="${TMP_DIR}/${ARCHIVE_NAME}"
CHECKSUM_PATH="${TMP_DIR}/${CHECKSUM_NAME}"
INSTALL_PATH="${INSTALL_DIR}/${APP_NAME}"
RESOLVED_SERVICE_SCOPE=''

if [[ "$INSTALL_SERVICE" -eq 1 ]]; then
  RESOLVED_SERVICE_SCOPE="$(resolve_service_scope "$SERVICE_SCOPE")"
fi

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

printf '==> downloading %s\n' "$ARCHIVE_NAME"
download "${BASE_URL}/${ARCHIVE_NAME}" "$ARCHIVE_PATH"

if download "${BASE_URL}/${CHECKSUM_NAME}" "$CHECKSUM_PATH" 2>/dev/null; then
  verify_checksum "$ARCHIVE_PATH" "$CHECKSUM_PATH"
else
  printf 'warning: checksums.txt not found, skipping checksum verification\n' >&2
fi

printf '==> extracting archive\n'
tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
if [[ ! -f "${TMP_DIR}/${APP_NAME}" ]]; then
  printf 'failed to find %s in the downloaded archive\n' "$APP_NAME" >&2
  exit 1
fi
chmod +x "${TMP_DIR}/${APP_NAME}"

INSTALL_CMD=(install -m 0755 "${TMP_DIR}/${APP_NAME}" "$INSTALL_PATH")
MKDIR_CMD=(mkdir -p "$INSTALL_DIR")
if needs_sudo_for_dir "$INSTALL_DIR"; then
  if [[ "$EUID" -eq 0 ]]; then
    printf '==> installing to %s\n' "$INSTALL_PATH"
    "${MKDIR_CMD[@]}"
    "${INSTALL_CMD[@]}"
  elif [[ "$USE_SUDO" -eq 1 ]]; then
    if ! has_cmd sudo; then
      printf 'install dir is not writable and sudo is not available; try --install-dir or run as root\n' >&2
      exit 1
    fi
    printf '==> installing to %s with sudo\n' "$INSTALL_PATH"
    sudo "${MKDIR_CMD[@]}"
    sudo "${INSTALL_CMD[@]}"
  else
    printf 'install dir is not writable; rerun without --no-sudo, use --install-dir, or run as root\n' >&2
    exit 1
  fi
else
  printf '==> installing to %s\n' "$INSTALL_PATH"
  "${MKDIR_CMD[@]}"
  "${INSTALL_CMD[@]}"
fi

if [[ "$INSTALL_SERVICE" -eq 1 ]]; then
  install_service
fi

printf '\n%s installed successfully\n' "$APP_NAME"
printf 'binary path: %s\n' "$INSTALL_PATH"

if [[ "$INSTALL_SERVICE" -eq 1 ]]; then
  case "$OS:$RESOLVED_SERVICE_SCOPE" in
    linux:system)
      printf 'service: %s.service (systemd, enabled at boot)\n' "$SERVICE_NAME"
      printf 'status: sudo systemctl status %s.service\n' "$SERVICE_NAME"
      ;;
    darwin:system)
      printf 'service: %s (launchd LaunchDaemon, enabled at boot)\n' "$SERVICE_LABEL"
      printf 'status: sudo launchctl print system/%s\n' "$SERVICE_LABEL"
      ;;
    darwin:user)
      printf 'service: %s (launchd LaunchAgent, starts on user login)\n' "$SERVICE_LABEL"
      printf 'status: launchctl print gui/%s/%s\n' "$(id -u)" "$SERVICE_LABEL"
      ;;
  esac
else
  printf 'service installation skipped\n'
fi

printf 'run `%s version` to verify the installation\n' "$APP_NAME"
