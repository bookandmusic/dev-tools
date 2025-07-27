#!/bin/bash
set -e

info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# 默认目录
DOCKER_INSTALL_DIR="/usr/local/docker"
PLUGIN_DIR="$HOME/.docker/plugins"
GITHUB_PROXY=""
ARCH=""

usage() {
    echo "Usage: $0 [--install-dir docker_install_dir] [--plugin-dir plugin_dir] [--github-proxy github_proxy]"
    echo "  --install-dir        Docker 安装目录，默认: $DOCKER_INSTALL_DIR"
    echo "  --plugin-dir         插件目录，默认: $PLUGIN_DIR"
    echo "  --github-proxy       GitHub 代理，用于下载插件"
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
            --github-proxy) GITHUB_PROXY="$2"; shift 2 ;;
            -h|--help) usage ;;
            *) error "未知选项: $1" ;;
        esac
    done
}

proxy_url() {
    if [ -z "$GITHUB_PROXY" ]; then
        echo "$1"
    else
        local proxy="${GITHUB_PROXY}"
        [[ "$proxy" != */ ]] && proxy="${proxy}/"
        echo "${proxy}$1"
    fi
}

fetch_latest_release_tag() {
    local repo="$1"
    local api_url="https://api.github.com/repos/$repo/releases/latest"
    local url
    url=$(proxy_url "$api_url")
    local tag
    tag=$(curl -sS "$url" | grep -oP '"tag_name":\s*"\K(.*)(?=")')
    if [ -z "$tag" ]; then
        error "获取 $repo 最新版本失败"
    fi
    echo "$tag"
}

download_docker() {
    info "获取 Docker 最新版本号..."
    DOCKER_VERSION=$(fetch_latest_release_tag "moby/moby")
    info "最新 Docker 版本: $DOCKER_VERSION"

    DOCKER_TGZ="docker-${DOCKER_VERSION#v}.tgz"
    DOCKER_URL="https://mirrors.aliyun.com/docker-ce/linux/static/stable/$ARCH/docker-${DOCKER_VERSION#v}.tgz"

    info "下载 Docker 二进制文件: $DOCKER_URL"
    curl -fsSL "$DOCKER_URL" -o "/tmp/$DOCKER_TGZ"

    info "解压 Docker 二进制到 $DOCKER_INSTALL_DIR"
    sudo rm -rf "$DOCKER_INSTALL_DIR"
    sudo mkdir -p "$DOCKER_INSTALL_DIR"
    tar -xzf "/tmp/$DOCKER_TGZ" -C /tmp
    sudo cp -r /tmp/docker/* "$DOCKER_INSTALL_DIR"/
    sudo chmod +x "$DOCKER_INSTALL_DIR"/*

    rm -rf /tmp/docker "/tmp/$DOCKER_TGZ"
}

get_plugin_filename() {
    local plugin_name="$1"
    local plugin_version="$2"

    declare -A ARCH_FILE_MAP_X86_64=(
        [compose]="docker-compose-linux-x86_64"
        [buildx]="buildx-${plugin_version}.linux-amd64"
    )
    declare -A ARCH_FILE_MAP_ARM64=(
        [compose]="docker-compose-linux-aarch64"
        [buildx]="buildx-${plugin_version}.linux-arm64"
    )

    if [ "$ARCH" = "x86_64" ]; then
        echo "${ARCH_FILE_MAP_X86_64[$plugin_name]}"
    elif [ "$ARCH" = "arm64" ]; then
        echo "${ARCH_FILE_MAP_ARM64[$plugin_name]}"
    else
        error "未知架构 $ARCH"
    fi
}

download_plugins() {
    sudo mkdir -p "$PLUGIN_DIR"

    for plugin in "${PLUGINS[@]}"; do
        local plugin_name
        plugin_name=$(basename "$plugin")
        local plugin_path="$PLUGIN_DIR/$plugin_name"

        info "获取插件 $plugin 最新版本..."
        local PLUGIN_VERSION
        PLUGIN_VERSION=$(fetch_latest_release_tag "$plugin")
        info "$plugin 最新版本: $PLUGIN_VERSION"

        local FILE_NAME
        FILE_NAME=$(get_plugin_filename "$plugin_name" "$PLUGIN_VERSION")

        local DOWNLOAD_URL="https://github.com/$plugin/releases/download/$PLUGIN_VERSION/$FILE_NAME"
        DOWNLOAD_URL=$(proxy_url "$DOWNLOAD_URL")

        info "下载插件 $plugin_name: $DOWNLOAD_URL"
        sudo curl -fsSL "$DOWNLOAD_URL" -o "$plugin_path"
        sudo chmod +x "$plugin_path"
    done
}

main() {
    check_sudo
    parse_args "$@"

    if [[ "$(uname)" != "Linux" ]]; then
        error "只支持 Linux 系统"
    fi

    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64) ARCH="x86_64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "不支持的架构: $ARCH" ;;
    esac

    PLUGINS=(
        "docker/compose"
        "docker/buildx"
    )

    info "使用 Docker 安装目录: $DOCKER_INSTALL_DIR"
    info "使用插件目录: $PLUGIN_DIR"
    info "GitHub代理: ${GITHUB_PROXY:-无}"

    download_docker
    download_plugins

    info "Docker 与插件已更新完成！"
}

main "$@"