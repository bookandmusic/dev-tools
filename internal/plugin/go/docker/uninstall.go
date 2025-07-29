package docker

import (
	"github.com/bookandmusic/dev-tools/internal/utils"
)

func (i *installer) Uninstall() error {
	step := 1

	// Step 1: 停止并禁用服务
	i.ui.Info("[%d/5] Stopping and disabling docker.service...", step)
	step++
	i.stopAndDisableDockerService()

	// Step 2: 删除 systemd 服务文件
	i.ui.Info("[%d/5] Removing systemd service file from %s...", step, i.cfg.ServicePath)
	step++
	if err := utils.RemoveBinaries(i.cfg.ServicePath, []string{"docker.service"}); err != nil {
		i.ui.Warning("Failed to remove docker.service: %v", err)
	} else {
		i.ui.Success("Removed docker.service from %s", i.cfg.ServicePath)
	}

	// 重新加载 systemd
	if err := utils.RunCommand("systemctl", "daemon-reload"); err != nil {
		i.ui.Warning("Failed to reload systemd daemon: %v", err)
	} else {
		i.ui.Success("systemd daemon reloaded.")
	}

	// Step 3: 删除 Docker 二进制文件
	i.ui.Info("[%d/5] Removing Docker binaries from %s...", step, i.cfg.BinDir)
	step++
	binaries := []string{
		"docker", "dockerd", "docker-init", "docker-proxy",
		"docker-runc", "docker-containerd", "docker-containerd-shim",
	}
	if err := utils.RemoveBinaries(i.cfg.BinDir, binaries); err != nil {
		i.ui.Warning("Failed to remove Docker binaries: %v", err)
	} else {
		i.ui.Success("Docker binaries removed from %s", i.cfg.BinDir)
	}

	// Step 4: 删除插件
	i.ui.Info("[%d/5] Removing Docker CLI plugins from %s...", step, i.cfg.PluginDir)
	step++
	plugins := []string{"docker-compose", "docker-buildx"}
	if err := utils.RemoveBinaries(i.cfg.PluginDir, plugins); err != nil {
		i.ui.Warning("Failed to remove CLI plugins: %v", err)
	} else {
		i.ui.Success("Docker CLI plugins removed from %s", i.cfg.PluginDir)
	}

	// Step 5: 卸载完成
	i.ui.Info("[%d/5] Finalizing uninstallation...", step)
	i.ui.Success("Docker uninstallation completed successfully.")
	return nil
}
