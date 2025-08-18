package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
)

type uiWriter struct {
	log func(msg string, args ...interface{})
}

func (w *uiWriter) Write(p []byte) (n int, err error) {
	// 去掉末尾换行符
	s := strings.TrimRight(string(p), "\r\n")
	if s != "" {
		w.log("%s", s)
	}
	return len(p), nil
}

// RunCmd 执行命令，实时输出到 UI
func RunCommand(ctx context.Context, u ui.UI, env map[string]string, name string, args ...string) error {
	u.Info(name + " " + strings.Join(args, " "))

	for k, v := range env {
		if k == envPath {
			os.Setenv(envPath, fmt.Sprintf("%s:%s", v, os.Getenv(envPath)))
		} else {
			os.Setenv(k, v)
		}
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &uiWriter{log: u.Println}
	cmd.Stderr = &uiWriter{log: u.Warning}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v, err: %w", name, args, err)
	}
	return nil
}
