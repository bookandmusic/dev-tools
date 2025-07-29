package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// AddPathToUserEnv ensures the dir is in PATH both for current session and future shells.
func AddPathToUserEnv(dir string) error {
	if dir == "" {
		return fmt.Errorf("empty directory path")
	}

	absDir := os.ExpandEnv(dir)
	if strings.HasPrefix(absDir, "~") {
		usr, _ := user.Current()
		absDir = filepath.Join(usr.HomeDir, absDir[1:])
	}

	// 更新 shell profile
	profileFile := detectProfileFile()
	if err := ensureProfileHasPath(profileFile, absDir); err != nil {
		return err
	}

	// 更新当前环境变量
	return AddToCurrentEnvPath(absDir)
}

// detectProfileFile returns the most likely profile file for current shell.
func detectProfileFile() string {
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

// ensureProfileHasPath appends PATH export if not exists.
func ensureProfileHasPath(profileFile, absDir string) error {
	content, _ := os.ReadFile(profileFile)
	if strings.Contains(string(content), absDir) {
		return nil // 已存在
	}

	exportLine := fmt.Sprintf("\n# Added by dev-tools\nexport PATH=\"%s:$PATH\"\n", absDir)
	f, err := os.OpenFile(profileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open profile file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(exportLine); err != nil {
		return fmt.Errorf("write to profile: %w", err)
	}
	return nil
}

// AddToCurrentEnvPath prepends dir to the current process PATH if it's not already present.
func AddToCurrentEnvPath(dir string) error {
	if dir == "" {
		return fmt.Errorf("empty dir")
	}

	// 获取当前的 PATH 环境变量
	currentPath := os.Getenv("PATH")
	newPath := prependToPath(currentPath, dir)

	// 只有在路径不同的情况下才更新 PATH
	if currentPath != newPath {
		return os.Setenv("PATH", newPath)
	}

	return nil
}

// BuildEnvPath builds a default PATH with binDir prepended if not already present.
func BuildEnvPath(binDir string) string {
	defaultPath := "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	return prependToPath(defaultPath, binDir)
}

// prependToPath prepends the given dir to the provided PATH if not already present.
func prependToPath(currentPath, dir string) string {
	// 如果路径已经存在于当前的 PATH 中，则不再添加
	paths := strings.Split(currentPath, ":")
	if len(paths) > 0 && paths[0] == dir {
		return currentPath
	}

	// 将新的路径插入到 PATH 前面
	return fmt.Sprintf("%s:%s", dir, currentPath)
}
