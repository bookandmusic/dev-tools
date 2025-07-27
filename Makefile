.PHONY: fmt lint vulncheck test all

# 1. 格式化代码 & 导包规范化
fmt:
	goimports-reviser -rm-unused -set-alias -format ./...

# 2. 静态检查 (golangci-lint)
lint:
	golangci-lint run ./...

# 3. 漏洞检查 (govulncheck)
vulncheck:
	govulncheck ./...

# 4. 单元测试
test:
	go test ./... -cover

# 一键全跑
all: fmt lint vulncheck test
