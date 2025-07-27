#!/bin/bash
set -e

info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
warn() { echo -e "\033[0;33m[WARN]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# 默认目录
DOCKER_INSTALL_DIR="/usr/local/docker"
PLUGIN_DIR="$HOME/.docker/plugins"
DOCKER_CONFIG_DIR="/etc/docker"
USER_DOCKER_DIR="$HOME/.docker"
SERVICE_FILE="/etc/systemd/system/docker.service"

usage() {
    echo "Usage: $0 [--install-dir docker_install_dir] [--plugin-dir plugin_dir]"
    echo "  --install-dir Docker 安装目录，默认: $DOCKER_INSTALL_DIR"
    echo "  --plugin-dir 插件目录，默认: $PLUGIN_DIR"
    exit 1
}

check_sudo() {
    if ! sudo -v; then
        error "当前用户无法获得sudo权限，脚本需要sudo权限执行"
    fi
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install-dir) DOCKER_INSTALL_DIR="$2"; shift 2 ;;
            --plugin-dir) PLUGIN_DIR="$2"; shift 2 ;;
            -h|--help) usage ;;
            *) error "未知选项: $1" ;;
        esac
    done
}

main() {
    check_sudo
    parse_args "$@"

    info "停止 Docker 服务..."
    sudo systemctl stop docker || warn "Docker 服务未运行"
    sudo systemctl disable docker || true

    info "删除 Docker 二进制目录: $DOCKER_INSTALL_DIR"
    sudo rm -rf "$DOCKER_INSTALL_DIR"

    info "删除插件目录: $PLUGIN_DIR"
    sudo rm -rf "$PLUGIN_DIR"

    info "删除用户 Docker 数据目录: $USER_DOCKER_DIR"
    rm -rf "$USER_DOCKER_DIR"

    info "删除 Docker 配置目录: $DOCKER_CONFIG_DIR"
    sudo rm -rf "$DOCKER_CONFIG_DIR"

    info "删除 systemd 服务文件: $SERVICE_FILE"
    sudo rm -f "$SERVICE_FILE"

    info "重新加载 systemd..."
    sudo systemctl daemon-reload

    info "Docker 已卸载完成！"
}

main "$@"