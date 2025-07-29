package docker

import (
	"fmt"
	"time"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

func (i *installer) Install() error {
	step := 1

	// Step 1: 架构检测
	i.ui.Info("[%d/8] Detecting system architecture...", step)
	step++
	if i.cfg.Arch == utils.ArchUnknown {
		return fmt.Errorf("unsupported architecture: %s", i.cfg.Arch)
	}
	i.ui.Success("Detected architecture: %s", i.cfg.Arch)

	// Step 2: 检测 Docker 及插件版本
	i.ui.Info("[%d/8] Detecting Docker and plugin versions...", step)
	step++
	if err := i.detectVersions(); err != nil {
		return err
	}

	// Step 3: 停止并禁用旧 Docker 服务
	i.ui.Info("[%d/8] Stopping and disabling any existing Docker service...", step)
	step++
	i.stopAndDisableDockerService()

	// Step 4: 安装 Docker
	i.ui.Info("[%d/8] Installing Docker binaries...", step)
	step++
	if err := i.installDocker(); err != nil {
		return err
	}
	i.ui.Success("Docker binaries installed successfully.")

	// Step 5: 配置守护进程
	i.ui.Info("[%d/8] Ensuring daemon configuration...", step)
	step++
	i.ensureDaemonConfig()

	// Step 6: 配置 systemd 并启动
	i.ui.Info("[%d/8] Setting up and starting Docker systemd service...", step)
	step++
	if err := i.setupSystemdService(); err != nil {
		return err
	}

	// Step 7: 等待守护进程就绪
	i.ui.Info("[%d/8] Waiting for Docker daemon to be ready...", step)
	step++
	if err := i.waitForDockerDaemon(5 * time.Second); err != nil {
		return err
	}
	i.ui.Success("Docker daemon is running.")

	// Step 8: 安装插件并验证
	i.ui.Info("[%d/8] Installing Docker plugins (compose, buildx)...", step)
	if err := i.installPlugins(); err != nil {
		return err
	}
	i.ensureClientConfig()

	i.ui.Info("Verifying Docker plugins...")
	if err := utils.RunCommand("docker", "compose", "version"); err != nil {
		return err
	}
	if err := utils.RunCommand("docker", "buildx", "version"); err != nil {
		return err
	}
	i.ui.Success("Docker installation completed successfully.")

	return nil
}
