package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

// AddEnvToUserProfile ensures the given envVars are added to the user profile and current process.
// - Supports ~, $VAR, and relative paths
// - If key is PATH, the value is prepended to PATH instead of overriding
// - Writes lines to the user's shell profile (~/.bashrc, ~/.zshrc, or ~/.profile)
// - Updates current process environment as well
func AddEnvToUserProfile(envVars map[string]string) error {
	if len(envVars) == 0 {
		return fmt.Errorf("no environment variables provided")
	}

	profileFile := DetectProfileFile()

	// 1️⃣ 生成写入 profile 的行
	var lines []string
	for k, v := range envVars {
		absValue := ExpandHomeAndEnv(v)

		if k == "PATH" {
			// PATH 特殊处理：前置 PATH
			lines = append(lines, fmt.Sprintf("export PATH=\"%s:$PATH\"", absValue))
		} else {
			// 普通环境变量
			lines = append(lines, fmt.Sprintf("export %s=%s", k, absValue))
		}
	}

	// 写入 profile（幂等）
	if err := ensureProfileHasLines(profileFile, lines); err != nil {
		return err
	}

	// 2️⃣ 更新当前进程环境
	for k, v := range envVars {
		absValue := ExpandHomeAndEnv(v)
		if k == "PATH" {
			// prepend PATH 并去重
			if err := AddToCurrentEnvPath(absValue); err != nil {
				return err
			}
		} else {
			os.Setenv(k, absValue)
		}
	}

	return nil
}

func ExpandHomeAndEnv(path string) string {
	path = os.ExpandEnv(path)
	if strings.HasPrefix(path, "~") {
		usr, _ := user.Current()
		path = filepath.Join(usr.HomeDir, path[1:])
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	return path
}

func ensureProfileHasLines(profileFile string, lines []string) error {
	content, _ := os.ReadFile(profileFile)
	updated := false

	f, err := os.OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open profile file: %w", err)
	}
	defer f.Close()

	for _, line := range lines {
		if !strings.Contains(string(content), line) {
			if _, err := f.WriteString("\n# Added by dev-tools\n" + line + "\n"); err != nil {
				return fmt.Errorf("write to profile: %w", err)
			}
			updated = true
		}
	}

	if updated {
		ui.Console.Info("Updated shell profile:", profileFile)
	}
	return nil
}

// detectProfileFile returns the most likely shell profile file for the current user.
func DetectProfileFile() string {
	shell := filepath.Base(os.Getenv("SHELL"))
	homeDir, _ := os.UserHomeDir()

	switch shell {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc")
	case "bash":
		return filepath.Join(homeDir, ".bashrc")
	default:
		return filepath.Join(homeDir, ".profile")
	}
}

func prependToPath(currentPath, dir string) string {
	if currentPath == "" {
		return dir
	}
	paths := strings.Split(currentPath, ":")
	for _, p := range paths {
		if p == dir {
			return currentPath // 已存在
		}
	}
	return fmt.Sprintf("%s:%s", dir, currentPath)
}

// BuildEnvPath 获取当前 PATH，并将 binDir 放在最前面（如果不存在的话）。
func BuildEnvPath(binDir string) string {
	currentPath := os.Getenv("PATH")
	return prependToPath(currentPath, binDir)
}

func AddToCurrentEnvPath(dir string) error {
	newPath := BuildEnvPath(dir)
	return os.Setenv("PATH", newPath)
}
