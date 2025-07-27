#!/bin/bash

# 检查并安装工具
check_and_install_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo "[ERROR] $1 未安装，正在安装..."
        case "$1" in
            "goimports-reviser")
                echo "[INFO] 安装 goimports-reviser..."
                go install github.com/incu6us/goimports-reviser/v2@latest
                ;;
            "gofumpt")
                echo "[INFO] 安装 gofumpt..."
                go install github.com/mvdan/gofumpt@latest
                ;;
            "golangci-lint")
                echo "[INFO] 安装 golangci-lint..."
                go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
                ;;
            "staticcheck")
                echo "[INFO] 安装 staticcheck..."
                go install honnef.co/go/tools/cmd/staticcheck@latest
                ;;
            "gosec")
                echo "[INFO] 安装 gosec..."
                go install github.com/securego/gosec/v2/cmd/gosec@latest
                ;;
            "gocyclo")
                echo "[INFO] 安装 gocyclo..."
                go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
                ;;
            "ineffassign")
                echo "[INFO] 安装 ineffassign..."
                go install github.com/gordonklaus/ineffassign@latest
                ;;
            "errcheck")
                echo "[INFO] 安装 errcheck..."
                go install github.com/kisielk/errcheck@latest
                ;;
            *)
                echo "[ERROR] 未知工具：$1"
                exit 1
                ;;
        esac
    else
        echo "[INFO] $1 已安装"
    fi
}

# 格式化代码
format_code() {
    echo "[INFO] 格式化 Go 代码..."

    # 使用 gofumpt 格式化
    gofumpt -l -w .
    # 使用 goimports-reviser 格式化和整理导入
    goimports-reviser -rm-unused -set-alias -format ./...
}

# 执行静态检查
check_code() {
    echo "[INFO] 执行静态检查..."

    # 检查 go vet
    echo "[INFO] 运行 go vet ..."
    go vet ./...

    # 检查 staticcheck
    echo "[INFO] 运行 staticcheck ..."
    staticcheck ./...

    # 检查 golangci-lint
    echo "[INFO] 运行 golangci-lint ..."
    golangci-lint run

    # # 检查 gosec
    # echo "[INFO] 运行 gosec ..."
    # gosec ./...

    # 检查 gocyclo
    echo "[INFO] 运行 gocyclo ..."
    gocyclo .

    # 检查 ineffassign
    echo "[INFO] 运行 ineffassign ..."
    ineffassign ./...

    # 检查 errcheck
    echo "[INFO] 运行 errcheck ..."
    errcheck ./...
}

# 主函数
main() {
    # 检查并安装工具
    check_and_install_tool "goimports-reviser"
    check_and_install_tool "gofumpt"
    check_and_install_tool "golangci-lint"
    check_and_install_tool "staticcheck"
    check_and_install_tool "gosec"
    check_and_install_tool "gocyclo"
    check_and_install_tool "ineffassign"
    check_and_install_tool "errcheck"

    # 格式化代码
    format_code

    # 执行静态检查
    check_code

    echo "[INFO] 完成！"
}

# 执行脚本
main
