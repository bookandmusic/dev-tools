# Dev-tools

## 🚀 开发环境准备

### 1. 安装 Go
请使用 Go >= 1.21  
[Go 官方安装指南](https://go.dev/dl/)

### 2. 安装工具
项目依赖以下工具（确保 `$HOME/go/bin` 在 PATH 中）：

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 导入规范化工具
go install github.com/incu6us/goimports-reviser/v3@latest

# 官方安全漏洞检查
go install golang.org/x/vuln/cmd/govulncheck@latest
