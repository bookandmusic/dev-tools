#!/bin/bash

set -e

# 默认值
GITHUB_PROXY=""
INSTALL_DIR="$HOME/.oh-my-zsh"

# 使用帮助函数
usage() {
    echo "Usage: $0 [-p github_proxy] [-p proxy] [-rm registry_mirrors] [-d install_dir]"
    echo "  -p github 代理，用于下载插件，默认不使用代理"
    echo "  -d  安装目录，默认安装到 ~/.oh-my-zsh"
    echo "  -p  Docker 服务代理（HTTP_PROXY/HTTPS_PROXY），支持多个用逗号分隔"
    echo "  -rm Docker registry 镜像加速地址，支持多个逗号分隔"
    exit 1
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -p|--github-proxy) GITHUB_PROXY="$2"; shift 2 ;;
            -p|--proxy) PROXY="$2"; shift 2 ;;
            -rm|--registry-mirrors) REGISTRY_MIRRORS="$2"; shift 2 ;;
            -d|--install-dir) INSTALL_DIR="$2"; shift 2 ;;
            -h|--help) usage ;;
            *) echo "未知选项: $1" && usage ;;
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
echo "检测系统类型: $OS_TYPE"

if [[ "$OS_TYPE" == "Darwin" ]]; then
  IS_MAC=true
else
  IS_MAC=false
fi

# 安装 zsh 和依赖
if ! $IS_MAC; then
  if ! command -v zsh &> /dev/null; then
      echo "zsh 未安装，正在安装..."
      if [ -f /etc/debian_version ]; then
          run_cmd sudo apt update
          run_cmd sudo apt install -y zsh git curl
      elif [ -f /etc/redhat-release ]; then
          run_cmd sudo yum install -y zsh git curl
      elif [ -f /etc/arch-release ]; then
          run_cmd sudo pacman -S --noconfirm zsh git curl
      else
          echo "未知或不支持的 Linux 发行版，请手动安装 zsh/git/curl"
          exit 1
      fi
  else
      echo "zsh 已安装"
  fi

  # 设置默认 shell
  CURRENT_SHELL=$(basename "$SHELL")
  if [ "$CURRENT_SHELL" != "zsh" ]; then
      echo "修改默认 shell 为 zsh"
      run_cmd chsh -s "$(which zsh)"
  else
      echo "默认 shell 已是 zsh"
  fi
else
  echo "macOS 系统，跳过安装 zsh 和修改默认 shell"
  # macOS 默认自带 zsh，一般不修改默认 shell
fi

# 检查并创建安装目录
if [ ! -d "$INSTALL_DIR" ]; then
    echo "通过代理克隆 oh-my-zsh 仓库到 $INSTALL_DIR ..."
    run_cmd git clone "$(proxy_url "https://github.com/ohmyzsh/ohmyzsh.git")" "$INSTALL_DIR"
else
    echo "oh-my-zsh 已存在于 $INSTALL_DIR"
fi

# 生成 .zshrc，如果不存在
if [ ! -e "$HOME/.zshrc" ]; then
    echo "生成 ~/.zshrc 配置文件"
    run_cmd cp "$INSTALL_DIR/templates/zshrc.zsh-template" "$HOME/.zshrc"
else
    echo "~/.zshrc 已存在，跳过生成"
fi

# 安装插件
plugins=(
    zsh-users/zsh-autosuggestions
    zsh-users/zsh-history-substring-search
    zsh-users/zsh-completions
    zsh-users/zsh-syntax-highlighting
)

echo "安装常用插件..."

for plugin in "${plugins[@]}"; do
    plugin_name=$(basename "$plugin")
    plugin_dir="$INSTALL_DIR/custom/plugins/$plugin_name"
    if [ ! -d "$plugin_dir" ]; then
        echo "通过代理克隆插件 $plugin ..."
        run_cmd git clone "$(proxy_url "https://github.com/$plugin.git")" "$plugin_dir"
    else
        echo "插件 $plugin_name 已安装"
    fi
done

# 修改 .zshrc，启用插件
echo "配置 .zshrc 插件..."

# 使用 sed 替换 plugins=() 行，保留这些插件（含 oh-my-zsh 内置插件 z sudo git extract）
if grep -q "^plugins=" ~/.zshrc; then
    sed -i.bak 's/^plugins=(.*)/plugins=(git sudo extract z zsh-autosuggestions zsh-history-substring-search zsh-completions zsh-syntax-highlighting)/' ~/.zshrc
else
    # 没有 plugins= 行，追加
    echo "plugins=(git sudo extract z zsh-autosuggestions zsh-history-substring-search zsh-completions zsh-syntax-highlighting)" >> ~/.zshrc
fi

echo "安装完成！请重新打开终端或运行：source ~/.zshrc"
