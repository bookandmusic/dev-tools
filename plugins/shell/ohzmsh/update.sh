#!/bin/bash

set -e

# 默认值
GITHUB_PROXY=""
INSTALL_DIR="$HOME/.oh-my-zsh"

# 输出信息的函数
info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
warn() { echo -e "\033[0;33m[WARN]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# 使用帮助函数
usage() {
    echo "Usage: $0 [-p github_proxy] [-d install_dir]"
    echo "  -p github 代理，用于下载插件，默认不使用代理"
    echo "  -d  安装目录，默认安装到 ~/.oh-my-zsh"
    exit 1
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -p|--github-proxy) GITHUB_PROXY="$2"; shift 2 ;;
            -d|--install-dir) INSTALL_DIR="$2"; shift 2 ;;
            -h|--help) usage ;;
            *) error "未知选项: $1" ;;
        esac
    done
}

# GitHub 代理 URL 构造函数
proxy_url() {
  if [ -n "$GITHUB_PROXY" ]; then
    echo "${GITHUB_PROXY}$1"
  else
    echo "$1"
  fi
}

# 执行命令的辅助函数
run_cmd() {
  echo "+ $*"
  "$@"
}

# 解析传入的参数
parse_args "$@"

# 判断系统类型
OS_TYPE="$(uname)"
info "检测系统类型: $OS_TYPE"

if [[ "$OS_TYPE" == "Darwin" ]]; then
  IS_MAC=true
else
  IS_MAC=false
fi

# 更新 oh-my-zsh
if [ -d "$INSTALL_DIR" ]; then
    info "更新 oh-my-zsh ..."
    run_cmd git -C "$INSTALL_DIR" pull --rebase
else
    error "oh-my-zsh 未安装，请先运行 install.sh"
fi

# 更新插件
plugins=(
    zsh-users/zsh-autosuggestions
    zsh-users/zsh-history-substring-search
    zsh-users/zsh-completions
    zsh-users/zsh-syntax-highlighting
)

info "更新插件..."

for plugin in "${plugins[@]}"; do
    plugin_name=$(basename "$plugin")
    plugin_dir="$INSTALL_DIR/custom/plugins/$plugin_name"
    
    if [ -d "$plugin_dir" ]; then
        info "更新插件 $plugin_name ..."
        run_cmd git -C "$plugin_dir" pull --rebase
    else
        warn "插件 $plugin_name 未安装，请运行 install.sh 安装"
    fi
done

# 修改 .zshrc，确保插件启用
info "配置 .zshrc 插件..."

# 使用 sed 确保插件行已更新
if grep -q "^plugins=" ~/.zshrc; then
    sed -i.bak 's/^plugins=(.*)/plugins=(git sudo extract z zsh-autosuggestions zsh-history-substring-search zsh-completions zsh-syntax-highlighting)/' ~/.zshrc
else
    # 没有 plugins= 行，追加
    echo "plugins=(git sudo extract z zsh-autosuggestions zsh-history-substring-search zsh-completions zsh-syntax-highlighting)" >> ~/.zshrc
fi

info "更新完成！请重新打开终端或运行：source ~/.zshrc"
