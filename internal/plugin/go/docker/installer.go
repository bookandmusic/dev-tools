package docker

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

// DockerManager interface
type DockerManager interface {
	Install() error
	Uninstall() error
}

// Config 配置参数
type Config struct {
	DockerVersion  string
	BuildxVersion  string
	ComposeVersion string
	Proxy          string
	Arch           utils.ArchType
	BinDir         string
	PluginDir      string
	ServicePath    string
}

// installer 实现 DockerManager 接口
type installer struct {
	cfg Config
	ui  ui.UI
}

// New 构造函数，自动设置默认值
func New(cfg Config, ui ui.UI) DockerManager {
	return &installer{cfg: cfg, ui: ui}
}

// Install 安装 Docker 及插件
func (i *installer) Install() error {
	if i.cfg.Arch == utils.ArchUnknown {
		return fmt.Errorf("unsupported architecture: %s", i.cfg.Arch)
	}
	i.ui.Info("Detected architecture: %s", i.cfg.Arch)

	if i.cfg.BuildxVersion == "" {
		i.ui.Info("Detecting latest buildx version...")
		tag, err := utils.GetLatestReleaseTag("docker/buildx", i.cfg.Proxy)
		if err != nil {
			return err
		}
		i.cfg.BuildxVersion = tag
		i.ui.Info("Detected buildx version: %s", tag)
	}
	if i.cfg.ComposeVersion == "" {
		i.ui.Info("Detecting latest compose version...")
		tag, err := utils.GetLatestReleaseTag("docker/compose", i.cfg.Proxy)
		if err != nil {
			return err
		}
		i.cfg.ComposeVersion = tag
		i.ui.Info("Detected compose version: %s", tag)
	}

	if err := i.downloadDocker(); err != nil {
		return err
	}
	if err := i.installPlugins(); err != nil {
		return err
	}
	if err := i.setupSystemd(); err != nil {
		return err
	}

	i.ui.Info("Verifying Docker installation...")
	if err := utils.RunCommand("docker", "version"); err != nil {
		return err
	}
	if err := utils.RunCommand("docker", "info"); err != nil {
		return err
	}
	if err := utils.RunCommand("docker", "compose", "version"); err != nil {
		return err
	}
	if err := utils.RunCommand("docker", "buildx", "version"); err != nil {
		return err
	}

	i.ui.Success("Docker installation completed successfully.")
	return nil
}

// Uninstall 卸载 Docker 和插件及 systemd 服务
func (i *installer) Uninstall() error {
	i.ui.Info("Stopping Docker service...")
	if err := utils.RunCommand("systemctl", "stop", "docker"); err != nil {
		i.ui.Warning("Failed to stop docker service: %v", err)
	}

	i.ui.Info("Disabling Docker service...")
	if err := utils.RunCommand("systemctl", "disable", "docker"); err != nil {
		i.ui.Warning("Failed to disable docker service: %v", err)
	}

	i.ui.Info("Removing systemd service file: %s", i.cfg.ServicePath)
	if err := os.Remove(i.cfg.ServicePath); err != nil && !os.IsNotExist(err) {
		i.ui.Warning("Failed to remove systemd service file: %v", err)
	}

	i.ui.Info("Reloading systemd daemon...")
	if err := utils.RunCommand("systemctl", "daemon-reload"); err != nil {
		i.ui.Warning("Failed to reload systemd daemon: %v", err)
	}

	i.ui.Info("Removing Docker binaries from %s...", i.cfg.BinDir)
	binaries := []string{"docker", "dockerd", "docker-init", "docker-proxy", "docker-runc", "docker-containerd", "docker-containerd-shim"}
	if err := utils.RemoveBinaries(i.cfg.BinDir, binaries); err != nil {
		i.ui.Warning("Failed to remove Docker binaries: %v", err)
	}

	i.ui.Info("Removing Docker CLI plugins from %s...", i.cfg.PluginDir)
	if err := os.Remove(filepath.Join(i.cfg.PluginDir, "docker-compose")); err != nil && !os.IsNotExist(err) {
		i.ui.Warning("Failed to remove docker-compose plugin: %v", err)
	}
	if err := os.Remove(filepath.Join(i.cfg.PluginDir, "docker-buildx")); err != nil && !os.IsNotExist(err) {
		i.ui.Warning("Failed to remove docker-buildx plugin: %v", err)
	}

	i.ui.Success("Docker uninstallation completed.")
	return nil
}

func (i *installer) downloadDocker() error {
	url := fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/docker-%s.tgz", i.cfg.Arch, i.cfg.DockerVersion)
	i.ui.Info("Downloading Docker %s from %s...", i.cfg.DockerVersion, url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeReg && strings.HasPrefix(hdr.Name, "docker/") {
			outPath := hdr.Name
			out, err := os.Create(outPath)
			if err != nil {
				return err
			}
			if _, err = io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			os.Chmod(outPath, 0o755)
		}
	}

	i.ui.Info("Installing Docker binaries to %s...", i.cfg.BinDir)
	return copyDir("docker", i.cfg.BinDir)
}

func (i *installer) installPlugin(pluginName, version, arch, urlTemplate string) error {
	pluginPath := filepath.Join(i.cfg.PluginDir, pluginName)
	url := fmt.Sprintf(urlTemplate, version, arch)

	// 如果设置了代理
	if i.cfg.Proxy != "" {
		url = i.cfg.Proxy + "/" + url
	}

	i.ui.Info("Installing plugin: %s", pluginName)
	i.ui.Info("-> Downloading from: %s", url)
	return utils.DownloadFileWithProgress(pluginPath, url, i.ui)
}

func (i *installer) installPluginCompose() error {
	composeURL := "https://github.com/docker/compose/releases/%s/download/docker-compose-linux-%s"
	return i.installPlugin("docker-compose", i.cfg.ComposeVersion, string(i.cfg.Arch), composeURL)
}

func (i *installer) installPluginBuildx() error {
	buildxURL := "https://github.com/docker/buildx/releases/%s/download/buildx-%s.linux-%s"
	arch := ""
	switch i.cfg.Arch {
	case utils.ArchAARCH64:
		arch = "arm64"
	default:
		arch = "amd64"
	}
	return i.installPlugin("docker-buildx", i.cfg.BuildxVersion, arch, buildxURL)
}

func (i *installer) installPlugins() error {
	i.ui.Info("Creating plugin directory: %s", i.cfg.PluginDir)
	if err := os.MkdirAll(i.cfg.PluginDir, 0o755); err != nil {
		return err
	}

	// 安装 docker-compose
	if err := i.installPluginCompose(); err != nil {
		return fmt.Errorf("install compose failed: %w", err)
	}

	// 安装 docker-buildx
	if err := i.installPluginBuildx(); err != nil {
		return fmt.Errorf("install buildx failed: %w", err)
	}

	return nil
}

func (i *installer) setupSystemd() error {
	i.ui.Info("Creating systemd service at %s...", i.cfg.ServicePath)

	content := `[Unit]
Description=Docker Service
After=network.target

[Service]
ExecStart=` + i.cfg.BinDir + `/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
Restart=always
LimitNOFILE=infinity

[Install]
WantedBy=multi-user.target
`

	if err := os.WriteFile(i.cfg.ServicePath, []byte(content), 0o644); err != nil {
		return err
	}

	i.ui.Info("Enabling and starting Docker service...")
	if err := utils.RunCommand("systemctl", "daemon-reexec"); err != nil {
		return err
	}
	if err := utils.RunCommand("systemctl", "daemon-reload"); err != nil {
		return err
	}
	if err := utils.RunCommand("systemctl", "enable", "docker"); err != nil {
		return err
	}
	return utils.RunCommand("systemctl", "start", "docker")
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())

		data, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dstFile, data, 0o755); err != nil {
			return err
		}
	}
	return nil
}
