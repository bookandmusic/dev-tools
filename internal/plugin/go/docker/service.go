package docker

import (
	"fmt"
	"strings"
	"time"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

func (i *installer) setupSystemdService() error {
	i.ui.Info("Setting up Docker systemd service...")

	envPath := utils.BuildEnvPath(i.cfg.BinDir)
	serviceContent := `[Unit]
Description=Docker Service
After=network.target

[Service]
WorkingDirectory=` + i.cfg.BinDir + `
Environment="PATH=` + envPath + `"
ExecStart=` + i.cfg.BinDir + `/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
Restart=always
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
`

	// 1. 写入 systemd service 文件
	if err := utils.WriteSystemdServiceFile(i.cfg.ServicePath, "docker.service", serviceContent); err != nil {
		i.ui.Error("Failed to write Docker systemd service file: %v", err)
		return err
	}
	i.ui.Success("Docker systemd service file created at %s", i.cfg.ServicePath)

	// 2. systemctl 操作
	cmds := [][]string{
		{"daemon-reload"},
		{"enable", "docker"},
		{"start", "docker"},
	}

	for _, cmd := range cmds {
		action := fmt.Sprintf("systemctl %s %s", cmd[0], strings.Join(cmd[1:], " "))
		i.ui.Info("Executing: %s...", action)

		if err := utils.RunSystemctlCommand(cmd...); err != nil {
			i.ui.Error("Failed to run %s: %v", action, err)
			return fmt.Errorf("failed to run %s: %w", action, err)
		}
		i.ui.Success("%s succeeded.", action)
	}

	return nil
}

func (i *installer) waitForDockerDaemon(timeout time.Duration) error {
	// 等待 systemd 激活
	i.ui.Info("Waiting up to %s for docker.service to become active...", timeout)
	if err := utils.WaitForSystemdServiceActive("docker", timeout); err != nil {
		i.ui.Error("docker.service did not become active in time: %v", err)
		return err
	}
	i.ui.Success("docker.service is active.")

	// 等待 docker info 成功
	i.ui.Info("Waiting up to %s for docker daemon status succeed...", timeout)
	if err := utils.WaitForCommandSuccess("docker", []string{"info"}, timeout); err != nil {
		i.ui.Error("`docker info` did not succeed in time: %v", err)
		return err
	}
	i.ui.Success("Docker daemon is fully operational.")
	return nil
}

func (i *installer) stopAndDisableDockerService() error {
	i.ui.Info("Stopping and disabling docker.service (if running)...")
	if err := utils.StopAndDisableService("docker"); err != nil {
		// 非致命，继续安装流程
		i.ui.Warning("Failed to stop or disable docker.service: %v", err)
		return err
	}
	i.ui.Success("docker.service stopped and disabled successfully.")
	return nil
}
