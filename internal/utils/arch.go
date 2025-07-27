package utils

import (
	"runtime"
)

// ArchType 定义支持的架构类型
type ArchType string

const (
	ArchX86_64  ArchType = "x86_64"
	ArchAARCH64 ArchType = "aarch64"
	ArchUnknown ArchType = "unknown"
)

// DetectArch 检测当前系统的 CPU 架构
func DetectArch() ArchType {
	switch runtime.GOARCH {
	case "amd64":
		return ArchX86_64
	case "arm64":
		return ArchAARCH64
	default:
		return ArchUnknown
	}
}
