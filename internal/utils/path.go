package utils

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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

func CreateIfNotExists(ctx context.Context, ui ui.UI, env map[string]string, path string, sudo bool) error {
	ui.Info("检测目录 %s ...", path)

	if _, err := os.Stat(path); err == nil {
		ui.Info("目录 %s 已经存在，跳过创建", path)
		return nil
	} else if !os.IsNotExist(err) {
		ui.Error("检查目录 %s 失败: %v", path, err)
		return err
	}

	ui.Info("创建目录 %s ...", path)
	var err error
	if sudo {
		err = RunCommand(ctx, ui, env, "sudo", "mkdir", "-p", path)
	} else {
		err = os.MkdirAll(path, 0o700)
	}

	if err != nil {
		ui.Error("创建目录 %s 失败: %v", path, err)
		return err
	}

	ui.Info("创建目录 %s 成功", path)
	return nil
}
