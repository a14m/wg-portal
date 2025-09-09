#!/bin/bash
set -euo pipefail

abort() {
    printf "\033[31m%s\033[0m\n" "$@" >&2
    exit 1
}

log() {
    printf "%s\n" "$@"
}

ensure_visudo() {
    if ! command -v visudo >/dev/null 2>&1; then
        log "visudo was not found"
        log "Please install visudo for your distro"
        abort "sudo must be installed"
    fi
}

ensure_curl() {
    if ! command -v curl >/dev/null 2>&1; then
        log "curl was not found"
        log "Please install curl for your distro"
        abort "curl must be installed"
    fi
}

ensure_acl() {
    if ! command -v setfacl >/dev/null 2>&1; then
        log "ACL tools not found"
        log "Please install acl for your distro"
        abort "setfacl must be installed"
    fi
}

ensure_wg() {
    if ! command -v wg >/dev/null 2>&1; then
        log "Wireguard tools not found"
        log "Please install wireguard for your distro"
        abort "wg/wg-quick must be installed"
    fi
    WIREGUARD_PATH="$(which wg)"
    WIREGUARD_QUICK_PATH="$(which wg-quick)"
}

ensure_systemd() {
    if ! systemctl --version >/dev/null 2>&1; then
        abort "systemd is required but not available"
    fi
}

ensure_sudo() {
    if [ "$EUID" -ne 0 ]; then
        abort "Must be run as sudo"
    fi
}

get_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) abort "Unsupported architecture: $ARCH" ;;
    esac
    log "Architecture: $ARCH"
}

get_version() {
    if [ "$VERSION" = "latest" ]; then
        RELEASE_URL="https://api.github.com/repos/a14m/wg-portal/releases/latest"
        VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name":' | cut -d '"' -f 4)
        log "Latest version: $VERSION"
    fi
}

validate_version() {
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] && [ "$VERSION" != "latest" ]; then
        abort "Invalid version format: $VERSION (valid format example v1.0.0)"
    fi
}

create_tmp() {
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf $TMP_DIR' EXIT
}

download_binary() {
    log "Downloading wg-portal binary"
    DOWNLOAD_URL="https://github.com/a14m/wg-portal/releases/download/$VERSION/wg-portal-linux-${ARCH}"
    if ! timeout 120 curl -s -L -o "$TMP_DIR/wg-portal" "$DOWNLOAD_URL"; then
        log "Failed to download $DOWNLOAD_URL"
        abort "Failed to download wg-portal binary"
    fi
    log "Downloaded $TMP_DIR/wg-portal"
}

download_systemd_service() {
    log "Downloading wg-portal.service"
    DOWNLOAD_SERVICE_URL="https://git.sr.ht/~a14m/wg-portal/blob/$VERSION/deployment/wg-portal.service"
    if ! timeout 30 curl -s -L -o "$TMP_DIR/wg-portal.service" "$DOWNLOAD_SERVICE_URL"; then
        log "Failed to download $DOWNLOAD_SERVICE_URL"
        abort "Failed to download wg-portal.service"
    fi
    log "Downloaded $TMP_DIR/wg-portal.service"
}

create_system_user() {
  if ! id -u wg-portal >/dev/null 2>&1; then
      log "Creating wg-portal system user and group..."
      useradd --system --no-create-home --shell /usr/sbin/nologin wg-portal
  else
      log "User wg-portal already exists"
  fi
}

create_etc_directory() {
    log "Creating /etc/wg-portal directory..."
    mkdir -p /etc/wg-portal
    chown root:wg-portal /etc/wg-portal
    chmod 750 /etc/wg-portal
}

configure_access_control() {
    log "Configure ACL for /etc/wireguard"
    setfacl -R -m g:wg-portal:r /etc/wireguard

    log "Setting up wg-portal user/group sudo permissions"
    cat > "$TMP_DIR/wg-portal-sudoers" << EOF
%wg-portal ALL=(ALL) NOPASSWD: ${WIREGUARD_QUICK_PATH} up *, ${WIREGUARD_QUICK_PATH} down *
%wg-portal ALL=(ALL) NOPASSWD: ${WIREGUARD_PATH} show
EOF
    # Validate before installing
    if visudo -c -f "$TMP_DIR/wg-portal-sudoers"; then
        mv "$TMP_DIR/wg-portal-sudoers" /etc/sudoers.d/wg-portal
        chmod 440 /etc/sudoers.d/wg-portal
    else
        abort "Invalid sudoers configuration"
    fi
}

install_binary() {
    log "Installing wg-portal binary to /etc/wg-portal"
    chmod +x "$TMP_DIR/wg-portal"
    mv "$TMP_DIR/wg-portal" /etc/wg-portal
}

install_systemd_service() {
    log "Installing wg-portal systemd service..."
    mv "$TMP_DIR/wg-portal.service" /etc/systemd/system/wg-portal.service
    systemctl daemon-reload
    systemctl enable wg-portal
    log "Service installed and enabled"
}

show_completion_message() {
  log ""
  log "wg-portal installed successfully!"
  log ""
  log "Configuration:"
  log "  - Configuration file: /etc/wg-portal/config"
  log ""
  log "Service Information:"
  log "  - Binary installed: /etc/wg-portal/wg-portal"
  log "  - Service: wg-portal.service (enabled)"
  log "  - User: wg-portal (system user)"
  log ""
  log "Useful Commands:"
  log "  - Check status: systemctl status wg-portal"
  log "  - View logs: journalctl -u wg-portal -f"
  log "  - Restart: systemctl restart wg-portal"
  log "  - Stop: systemctl stop wg-portal"
  log ""
  log "Next steps:"
  log "  - Configure: /etc/wg-portal/config"
  log "  - Start Service: systemctl start wg-portal"
}


# Configuration
VERSION=${1:-latest}
# Main
ensure_visudo
ensure_curl
ensure_acl
ensure_wg
ensure_systemd
ensure_sudo
get_arch
get_version
validate_version
create_tmp
download_binary
download_systemd_service
create_system_user
create_etc_directory
configure_access_control
install_binary
install_systemd_service
show_completion_message
