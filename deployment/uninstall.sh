#!/bin/bash
set -euo pipefail

abort() {
    printf "\033[31m%s\033[0m\n" "$@" >&2
    exit 1
}

log() {
    printf "%s\n" "$@"
}

ensure_sudo() {
    if [ "$EUID" -ne 0 ]; then
        abort "Must be run as sudo"
    fi
}

uninstall_systemd_configurations() {
    log "Removing wg-portal systemd configurations"
    systemctl stop wg-portal 2>/dev/null || true
    systemctl disable wg-portal 2>/dev/null || true
    rm -f /etc/systemd/system/wg-portal.service
    systemctl daemon-reload
}

remove_directories() {
    log "Removing wg-portal directories..."
    rm -rf /etc/wg-portal
}

remove_sudo_configurations() {
    log "Removing wg-portal sudo permissions..."
    rm -f /etc/sudoers.d/wg-portal
}

remove_acl_permissions() {
    if [ -d /etc/wireguard ]; then
        log "Removing wg-portal ACL permissions from /etc/wireguard..."
        setfacl -R -x g:wg-portal /etc/wireguard 2>/dev/null || true
    fi
}

remove_system_user() {
    log "Removing wg-portal system user and group..."
    userdel wg-portal 2>/dev/null || true
    groupdel wg-portal 2>/dev/null || true
}

show_completion_message() {
    log ""
    log "wg-portal uninstalled successfully!"
}

# Main
ensure_sudo
uninstall_systemd_configurations
remove_directories
remove_sudo_configurations
remove_acl_permissions
remove_system_user
show_completion_message
