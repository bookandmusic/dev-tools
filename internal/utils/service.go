package utils

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

var ErrTimeout = errors.New("operation timed out")

// WriteSystemdServiceFile 创建 systemd 服务文件目录（如果不存在）并写入内容
func WriteSystemdServiceFile(serviceDir, serviceName, content string) error {
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return err
	}
	serviceFile := filepath.Join(serviceDir, serviceName)
	return os.WriteFile(serviceFile, []byte(content), 0o644)
}

// WaitForSystemdServiceActive 等待 systemd 服务激活，直到超时，成功返回 nil，超时返回 ErrTimeout
func WaitForSystemdServiceActive(service string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		err := RunSystemctlCommand("is-active", "--quiet", service)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ErrTimeout
}

// WaitForCommandSuccess 等待指定命令成功返回，直到超时
func WaitForCommandSuccess(command string, args []string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if _, err := RunCommandCapture(command, args...); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ErrTimeout
}

// StopAndDisableService 先停止再禁用指定 systemd 服务
func StopAndDisableService(name string) error {
	if err := RunSystemctlCommand("stop", name); err != nil {
		return err
	}
	if err := RunSystemctlCommand("disable", name); err != nil {
		return err
	}
	return nil
}
