package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

// ========== 辅助子方法 ==========

// 下载Docker包
func (d *DockerManager) downloadIfNotExists(ui ui.UI, arch utils.ArchType, tarFilePath string) error {
	ui.Info("下载Docker安装包...")
	tmpPath := filepath.Dir(tarFilePath)
	tarFile := filepath.Base(tarFilePath)
	if !utils.PathExists(tarFilePath) {
		url := fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/%s", arch, tarFile)
		if err := utils.DownloadFileWithProgress(url, tarFilePath, ui, ""); err != nil {
			ui.Error("下载Docker失败")
			return err
		}
	} else {
		ui.Info("缓存目录%s已存在%s,跳过下载", tmpPath, tarFile)
	}
	return nil
}

// 设置执行权限
func (d *DockerManager) chmodFiles(ctx context.Context, ui ui.UI, env map[string]string, dir string) error {
	ui.Info("添加可执行权限...")
	if err := utils.RunCommand(
		ctx, ui, env,
		"sudo", "bash", "-c",
		fmt.Sprintf("find %s -maxdepth 1 -type f -exec chmod +x {} \\;", dir),
	); err != nil {
		ui.Error("设置执行权限失败")
		return err
	}
	return nil
}

// 生成 systemd service
func (d *DockerManager) generateSystemdService(ctx context.Context, ui ui.UI, env map[string]string, binPath string) error {
	serviceFile := "/etc/systemd/system/docker.service"
	ui.Info("正在生成%s文件...", serviceFile)
	serviceTem := fmt.Sprintf(`
[Unit]
Description=Docker Application Container Engine
After=network.target

[Service]
Type=notify
WorkingDirectory=%s
Environment="PATH=%s:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ExecStart=%s/dockerd
ExecReload=/bin/kill -s HUP \$MAINPID
Restart=always
StartLimitBurst=3
StartLimitIntervalSec=60

[Install]
WantedBy=multi-user.target
`, binPath, binPath, binPath)
	cmd := fmt.Sprintf("cat <<'EOF' | sudo tee %s > /dev/null\n%s\nEOF", serviceFile, serviceTem)
	return utils.RunCommand(ctx, ui, env, "bash", "-c", cmd)
}

// 合并JSON并写入文件
func (d *DockerManager) mergeJSONToFile(
	ctx context.Context,
	ui ui.UI,
	env map[string]string,
	file string,
	defaults map[string]interface{},
	useSudo bool,
) error {
	var content []byte
	var orig map[string]interface{}

	if utils.PathExists(file) {
		ui.Info("%s 已存在，读取并合并内容...", file)
		var err error
		content, err = os.ReadFile(file)
		if err != nil {
			ui.Error("读取 %s 失败: %v", file, err)
			return err
		}
		if len(content) > 0 {
			if err := json.Unmarshal(content, &orig); err != nil {
				ui.Warning("解析 %s 失败，使用空配置: %v", file, err)
				orig = make(map[string]interface{})
			}
		}
	}

	// 合并配置
	jsonStr, err := utils.MergeJSON(orig, defaults)
	if err != nil {
		ui.Error("合并 %s 失败: %v", file, err)
		return err
	}

	// 写入文件
	writeCmd := fmt.Sprintf("cat <<'EOF' | %s tee %s > /dev/null\n%s\nEOF",
		func() string {
			if useSudo {
				return "sudo"
			}
			return ""
		}(),
		file, jsonStr,
	)

	// 去掉多余空格（当 useSudo = false 时）
	writeCmd = strings.ReplaceAll(writeCmd, "|  tee", "| tee")

	return utils.RunCommand(ctx, ui, env, "bash", "-c", writeCmd)
}
