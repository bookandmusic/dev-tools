#!/bin/bash

set -e

# 默认值
INSTALL_DIR="$HOME/.oh-my-zsh"

# 输出信息的函数
info() { echo -e "\033[0;32m[INFO]\033[0m $*"; }
warn() { echo -e "\033[0;33m[WARN]\033[0m $*"; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $*"; exit 1; }

# 使用帮助函数
usage() {
    echo "Usage: $0 [-d install_dir]"
    echo "  -d  安装目录，默认安装到 ~/.oh-my-zsh"
    exit 1
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -d|--install-dir) INSTALL_DIR="$2"; shift 2 ;;
            -h|--help) usage ;;
            *) error "未知选项: $1" ;;
        esac
    done
}

# 解析传入的参数
parse_args "$@"

# 检查安装目录是否存在
if [ -d "$INSTALL_DIR" ]; then
    info "删除 oh-my-zsh 目录 $INSTALL_DIR ..."
    rm -rf "$INSTALL_DIR"
else
    warn "oh-my-zsh 安装目录 $INSTALL_DIR 不存在"
fi

# 删除 .zshrc 配置文件
if [ -e "$HOME/.zshrc" ]; then
    info "删除 ~/.zshrc 配置文件"
    rm -f "$HOME/.zshrc"
else
    warn "~/.zshrc 配置文件不存在"
fi

info "卸载完成！"
