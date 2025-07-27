package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

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
