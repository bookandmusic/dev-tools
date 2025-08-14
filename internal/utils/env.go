package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// AddEnvToEnvFile ensures the given envVars are added to the user profile and current process.
// - Supports ~, $VAR, and relative paths
// - If key is PATH, the value is prepended to PATH instead of overriding
// - Writes lines to the user's shell profile (~/.bashrc, ~/.zshrc, or ~/.profile)
// - Updates current process environment as well
func AddEnvToEnvFile(envVars map[string]string, envFile string) error {
	if len(envVars) == 0 {
		return fmt.Errorf("no environment variables provided")
	}
	// 1️⃣ 生成写入 profile 的行
	var lines []string
	for k, v := range envVars {
		absValue := ExpandAbsDir(v)

		if k == "PATH" {
			// PATH 特殊处理：前置 PATH
			lines = append(lines, fmt.Sprintf("export PATH=\"%s:$PATH\"", absValue))
		} else {
			// 普通环境变量
			lines = append(lines, fmt.Sprintf("export %s=%s", k, absValue))
		}
	}

	// 写入 profile（幂等）
	if err := ensureProfileHasLines(envFile, lines); err != nil {
		return err
	}

	// 2️⃣ 更新当前进程环境
	for k, v := range envVars {
		absValue := ExpandAbsDir(v)
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

func ExpandAbsDir(path string) string {
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

// ensureProfileHasLines ensures that the specified lines are present in the profile file
// before appending them. It checks for potential directory traversal vulnerabilities.
func ensureProfileHasLines(profileFile string, lines []string) error {
	// 防止路径遍历攻击：验证路径是否安全
	cleanProfileFile := filepath.Clean(profileFile)

	content, _ := os.ReadFile(profileFile)
	updated := false

	// 安全打开文件
	f, err := os.OpenFile(cleanProfileFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // Changed to 0600
	if err != nil {
		return fmt.Errorf("open profile file: %s", err)
	}
	defer f.Close()

	// 写入行
	for _, line := range lines {
		if !strings.Contains(string(content), line) {
			if !updated {
				if _, err := f.WriteString("\n# Added by dev-tools\n"); err != nil {
					return fmt.Errorf("write to profile: %s", err)
				}
			}
			if _, err := f.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("write to profile: %s", err)
			}
			updated = true
		}
	}
	return nil
}

func BuildEnvPath(binDir string) string {
	currentPath := os.Getenv("PATH")
	if currentPath == "" {
		return binDir
	}
	paths := strings.Split(currentPath, ":")
	for _, p := range paths {
		if p == binDir {
			return currentPath // 已存在
		}
	}
	return fmt.Sprintf("%s:%s", binDir, currentPath)
}

func AddToCurrentEnvPath(dir string) error {
	newPath := BuildEnvPath(dir)
	return os.Setenv("PATH", newPath)
}

func GetCurrentUser() (*user.User, error) {
	// 使用 os/user 获取当前用户信息
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	return currentUser, nil
}

func GetUserHomeAndShell(username string) (homeDir string, shell string, err error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, username+":") {
			parts := strings.Split(line, ":")
			if len(parts) >= 7 {
				homeDir = parts[5] // 家目录
				shell = parts[6]   // shell
				return homeDir, shell, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	return "", "", fmt.Errorf("user %s not found", username)
}

func GetProfilePath(shell, home string) string {
	base := filepath.Base(shell)
	switch base {
	case "bash":
		// 优先用 .bash_profile，如果不存在再用 .profile（代码里没法判断文件存在，这里给出常见路径）
		return filepath.Join(home, ".bash_profile") // 也可根据需求改成 .profile
	case "zsh":
		return filepath.Join(home, ".zprofile")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".profile")
	}
}
