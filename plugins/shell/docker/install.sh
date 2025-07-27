#!/bin/bash
# set -e

info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
warn() { echo -e "\033[0;33m[WARN]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# 全局变量（可根据需要调整）
DOCKER_INSTALL_DIR="/usr/local/docker"
PLUGIN_DIR="$HOME/.docker/plugins"
DOCKER_CONFIG_DIR="/etc/docker"
DAEMON_JSON="$DOCKER_CONFIG_DIR/daemon.json"

GITHUB_PROXY=""
PROXY=""
REGISTRY_MIRRORS=""

check_sudo() {
    if ! sudo -v; then
        error "当前用户无法获得sudo权限，脚本需要sudo权限执行"
    fi
}

usage() {
    echo "Usage: $0 [--install-dir install_dir] [--plugin-dir plugin_dir] [--github-proxy github_proxy] [--proxy proxy] [--registry-mirrors registry_mirrors]"
    echo "  --install-dir        Docker 安装目录，默认: $DOCKER_INSTALL_DIR"
    echo "  --plugin-dir         插件目录，默认: $PLUGIN_DIR"
    echo "  --github-proxy       GitHub 代理，用于下载插件"
    echo "  --proxy              Docker 服务代理（HTTP_PROXY/HTTPS_PROXY），支持多个用逗号分隔"
    echo "  --registry-mirrors   Docker registry 镜像加速地址，支持多个逗号分隔"
    exit 1
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install-dir) DOCKER_INSTALL_DIR="$2"; shift 2 ;;
            --plugin-dir) PLUGIN_DIR="$2"; shift 2 ;;
            --github-proxy) GITHUB_PROXY="$2"; shift 2 ;;
            --proxy) PROXY="$2"; shift 2 ;;
            --registry-mirrors) REGISTRY_MIRRORS="$2"; shift 2 ;;
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

    if [ ! -d "$DOCKER_INSTALL_DIR" ]; then
        info "创建 docker 安装目录 $DOCKER_INSTALL_DIR"
        sudo mkdir -p "$DOCKER_INSTALL_DIR"
        sudo chown "$(whoami)" "$DOCKER_INSTALL_DIR"
    fi

    DOCKER_TGZ="docker-${DOCKER_VERSION#v}.tgz"
    DOCKER_URL="https://mirrors.aliyun.com/docker-ce/linux/static/stable/$ARCH/docker-${DOCKER_VERSION#v}.tgz"

    info "下载 Docker 二进制文件: $DOCKER_URL"
    curl -fSL "$DOCKER_URL" -o "/tmp/$DOCKER_TGZ" -#

    info "解压 Docker 二进制到 $DOCKER_INSTALL_DIR"
    tar -xzf "/tmp/$DOCKER_TGZ" -C /tmp
    sudo cp -r /tmp/docker/* "$DOCKER_INSTALL_DIR"/
    sudo chmod +x "$DOCKER_INSTALL_DIR"/*

    rm -rf /tmp/docker "/tmp/$DOCKER_TGZ"
}

setup_systemd_service() {
    local CURRENT_PATH
    CURRENT_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v "^$DOCKER_INSTALL_DIR\$" | paste -sd: -)
    local FULL_PATH="$DOCKER_INSTALL_DIR:$CURRENT_PATH"

    SERVICE_FILE="/etc/systemd/system/docker.service"
    info "生成 systemd service 文件: $SERVICE_FILE"
    sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Docker Application Container Engine
After=network.target

[Service]
Type=notify
WorkingDirectory=$DOCKER_INSTALL_DIR
Environment="PATH=$FULL_PATH"
ExecStart=$DOCKER_INSTALL_DIR/dockerd
ExecReload=/bin/kill -s HUP \$MAINPID
Restart=always
StartLimitBurst=3
StartLimitIntervalSec=60

[Install]
WantedBy=multi-user.target
EOF

    info "重新加载 systemd 并启用 docker 服务"
    sudo systemctl daemon-reload
    sudo systemctl enable docker
    sudo systemctl restart docker
}

to_json_array() {
    IFS=',' read -ra arr <<< "$1"
    local out="["
    for i in "${arr[@]}"; do
        out+="\"$i\","
    done
    out="${out%,}]"
    echo "$out"
}

setup_docker_daemon_json() {
    sudo mkdir -p "$DOCKER_CONFIG_DIR"

    local http_proxy_val="\"\""
    local https_proxy_val="\"\""
    local no_proxy_val="\"\""

    if [ -n "$PROXY" ]; then
        # 只取第一个代理地址作为字符串（如果多个，用逗号分隔也不适合这里）
        IFS=',' read -r first_proxy _ <<< "$PROXY"
        http_proxy_val="\"$first_proxy\""
        https_proxy_val="\"$first_proxy\""
    fi

    if [ -n "$NO_PROXY" ]; then
        no_proxy_val="\"$NO_PROXY\""
    else
        # 默认 no-proxy，常见写法
        no_proxy_val="\"localhost,127.0.0.1\""
    fi

    local registry_json_val="[]"
    if [ -n "$REGISTRY_MIRRORS" ]; then
        IFS=',' read -ra mirrors <<< "$REGISTRY_MIRRORS"
        local arr="["
        for m in "${mirrors[@]}"; do
            arr+="\"$m\","
        done
        arr="${arr%,}]"
        registry_json_val="$arr"
    fi

    # 生成 daemon.json 内容
    sudo tee "$DAEMON_JSON" > /dev/null << EOF
{
  "registry-mirrors": $registry_json_val,
  "proxies": {
    "http-proxy": $http_proxy_val,
    "https-proxy": $https_proxy_val,
    "no-proxy": $no_proxy_val
  }
}
EOF

    info "生成 daemon.json 完成: $DAEMON_JSON"
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
    if [ ! -d "$PLUGIN_DIR" ]; then
        info "创建插件目录 $PLUGIN_DIR"
        sudo mkdir -p "$PLUGIN_DIR"
    fi

    for plugin in "${PLUGINS[@]}"; do
        local plugin_name=$(basename "$plugin")
        local plugin_path="$PLUGIN_DIR/docker-$plugin_name"
        info "获取插件 $plugin 最新版本..."
        local PLUGIN_VERSION
        PLUGIN_VERSION=$(fetch_latest_release_tag "$plugin")
        info "$plugin 最新版本: $PLUGIN_VERSION"

        local FILE_NAME
        FILE_NAME=$(get_plugin_filename "$plugin_name" "$PLUGIN_VERSION")

        local DOWNLOAD_URL="https://github.com/$plugin/releases/download/$PLUGIN_VERSION/$FILE_NAME"
        DOWNLOAD_URL=$(proxy_url "$DOWNLOAD_URL")


        info "下载插件 $plugin_name: $DOWNLOAD_URL"
        sudo curl -fsSL "$DOWNLOAD_URL" -o "$plugin_path" -#
        sudo chmod +x "$plugin_path"

    done
}

add_docker_client_plugin_dir() {
    local config_dir="$HOME/.docker"
    local config_file="$config_dir/config.json"
    local key="cliPluginsExtraDirs"

    # 创建配置目录
    if [ ! -d "$config_dir" ]; then
        info "创建 Docker Client 配置目录: $config_dir"
        mkdir -p "$config_dir"
    fi

    # 创建空 JSON 文件
    if [ ! -f "$config_file" ]; then
        info "创建空的 Docker Client 配置文件: $config_file"
        echo "{}" > "$config_file"
    fi

    # 用 jq 确保 cliPluginsExtraDirs 存在且包含 $PLUGIN_DIR
    local tmp_file
    tmp_file=$(mktemp)
    jq --arg dir "$PLUGIN_DIR" '
        .cliPluginsExtraDirs |=
        (if . == null then [$dir]
         else (if index($dir) == null then . + [$dir] else . end)
         end)
    ' "$config_file" > "$tmp_file" && mv "$tmp_file" "$config_file"

    info "已将插件目录添加到 Docker Client 配置: $PLUGIN_DIR"
}



update_path_in_profile() {
    local shell_name="$1"
    local profile_file="$2"
    local path_entry="$DOCKER_INSTALL_DIR"

    if [ ! -f "$profile_file" ]; then
        touch "$profile_file"
    fi

    if grep -q "$path_entry" "$profile_file"; then
        info "$profile_file 已包含 $path_entry，跳过添加PATH"
    else
        info "向 $profile_file 添加 $path_entry 到 PATH"
        echo "" >> "$profile_file"
        echo "# Added by docker install script to include docker binaries" >> "$profile_file"

        case "$shell_name" in
            bash|zsh)
                echo "export PATH=\"$path_entry:\$PATH\"" >> "$profile_file"
                ;;
            fish)
                echo "set -gx PATH $path_entry \$PATH" >> "$profile_file"
                ;;
            *)
                echo "# 你的 shell $shell_name 可能需要手动添加 PATH: $path_entry" >> "$profile_file"
                ;;
        esac
    fi
}

detect_and_update_shell_path() {
    local user_shell
    user_shell=$(basename "$SHELL")
    local home_dir="$HOME"
    case "$user_shell" in
        bash)
            update_path_in_profile "bash" "$home_dir/.bashrc"
            ;;
        zsh)
            update_path_in_profile "zsh" "$home_dir/.zshrc"
            ;;
        fish)
            update_path_in_profile "fish" "$home_dir/.config/fish/config.fish"
            ;;
        *)
            warn "未知 shell 类型 $user_shell，无法自动添加 PATH，请手动添加 $DOCKER_INSTALL_DIR 到 PATH"
            ;;
    esac
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

    info "系统架构: $ARCH"
    info "Docker 安装路径（固定）: $DOCKER_INSTALL_DIR"
    info "插件路径（固定）: $PLUGIN_DIR"
    info "GitHub代理: ${GITHUB_PROXY:-无}"
    info "Docker 服务代理: ${PROXY:-无}"
    info "Registry 镜像加速: ${REGISTRY_MIRRORS:-无}"

    download_docker
    setup_docker_daemon_json
    setup_systemd_service


    PLUGINS=(
        "docker/compose"
        "docker/buildx"
    )
    download_plugins

    add_docker_client_plugin_dir

    detect_and_update_shell_path

    info "安装完成！docker 服务已启动。"
}

main "$@"