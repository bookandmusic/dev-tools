package docker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type DockerManager struct{}

func (d *DockerManager) Install(ctx context.Context) error {
	var err error
	var params *BaseParams
	params, err = soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}

	ui := params.UI
	cfg := params.Cfg
	global := params.Global
	env := params.Env
	version := cfg.Version

	ui.Info("检测sudo权限")
	if err = utils.RunCommand(ctx, ui, env, "sudo", "-v"); err != nil {
		ui.Error("获取sudo权限失败")
		return err
	}
	ui.Info("尝试停止正在运行的Docker服务...")
	_ = utils.RunCommand(ctx, ui, env, "sudo", "systemctl", "stop", "docker")

	// ========= 第一步：获取版本 & 架构 =========
	arch := utils.DetectArch()
	if arch == utils.ArchUnknown {
		ui.Error("工具只支持AMD64、AARCH64, 不支持当前主机架构: %s", runtime.GOARCH)
		return errors.New("unknown architecture")
	}

	if version == "" {
		ui.Info("获取Docker最新版本...")
		versionStr, err := utils.GetLatestReleaseTag(ui, "moby/moby", global.GithubProxy, global.HttpProxy)
		if err != nil {
			ui.Error("获取Docker最新版本失败")
			return err
		}
		version = strings.TrimPrefix(versionStr, "v")
	}
	ui.Info("准备安装 Docker %s, 架构: %s", version, arch)

	// 获取插件版本
	composeVersion, err := utils.GetLatestReleaseTag(ui, "docker/compose", global.GithubProxy, global.HttpProxy)
	if err != nil {
		ui.Error("获取docker-compose最新版本失败")
		return err
	}
	buildxVersion, err := utils.GetLatestReleaseTag(ui, "docker/buildx", global.GithubProxy, global.HttpProxy)
	if err != nil {
		ui.Error("获取docker-buildx最新版本失败")
		return err
	}
	var buildxArch string
	switch arch {
	case utils.ArchX86_64:
		buildxArch = "amd64"
	default:
		buildxArch = "arm64"
	}
	ui.Info("插件版本: compose=%s, buildx=%s (%s)", composeVersion, buildxVersion, buildxArch)

	// ========= 第二步：下载 & 安装Docker =========
	installPath := cfg.InstallDir
	if installPath == "" {
		installPath = path.Join(global.RootDir, "docker")
	}
	installPath = utils.ExpandAbsDir(installPath)
	binPath := path.Join(installPath, "bin")
	ui.Info("安装目录 %s ...", binPath)
	if err = utils.CreateIfNotExists(ctx, ui, env, binPath, true); err != nil {
		return err
	}

	tmpPath := path.Join(global.CacheDir, "docker", version)
	tarFile := fmt.Sprintf("docker-%s.tgz", version)
	tarFilePath := path.Join(tmpPath, tarFile)
	if err := d.downloadIfNotExists(ui, arch, tarFilePath); err != nil {
		return err
	}
	ui.Info("解压文件:%s 到 %s", tarFile, binPath)
	if err = utils.ExtractTarGzWithProgress(ui, tarFilePath, binPath, 1); err != nil {
		ui.Error("解压Docker失败")
		return err
	}
	if err = d.chmodFiles(ctx, ui, env, binPath); err != nil {
		return err
	}

	// ========= 第三步：生成配置文件 =========
	if err = d.generateSystemdService(ctx, ui, env, binPath); err != nil {
		return err
	}
	if err = d.mergeJSONToFile(ctx, ui, env, "/etc/docker/daemon.json", map[string]interface{}{
		"fixed-cidr-v6":       "fd00::/80",
		"insecure-registries": []string{},
		"ipv6":                true,
		"proxies": map[string]string{
			"http-proxy":  cfg.HttpProxy,
			"https-proxy": cfg.HttpProxy,
		},
		"registry-mirrors": cfg.RegistryMirrors,
	}, true); err != nil {
		return err
	}

	// ========= 第四步：安装插件 =========
	pluginDir := path.Join(installPath, "plugins")
	if err := utils.CreateIfNotExists(ctx, ui, env, pluginDir, true); err != nil {
		ui.Error("创建插件目录失败")
		return err
	}

	// helper function
	installPlugin := func(name, version, pluginFile, baseUrl string) error {
		cachePath := path.Join(global.CacheDir, "docker", fmt.Sprintf("%s-%s", name, version))
		if err := utils.CreateIfNotExists(ctx, ui, env, cachePath, false); err != nil {
			return err
		}

		cachedFilePath := path.Join(cachePath, pluginFile)
		if _, err := os.Stat(cachedFilePath); os.IsNotExist(err) {
			// 缓存不存在，下载
			downloadUrl := fmt.Sprintf("%s/%s/%s", baseUrl, version, pluginFile)
			ui.Info("下载插件 %s 到缓存: %s", name, cachedFilePath)
			if global.HttpProxy == "" {
				downloadUrl = utils.ProxyURL(ui, global.GithubProxy, downloadUrl)
			}
			if err := utils.DownloadFileWithProgress(downloadUrl, cachedFilePath, ui, global.HttpProxy); err != nil {
				return err
			}
		} else {
			ui.Info("缓存已存在插件 %s, 跳过下载", name)
		}

		// 从缓存复制到插件目录
		destPath := path.Join(pluginDir, name)
		ui.Info("复制插件 %s 到 %s", name, destPath)
		if err := utils.CopyFile(cachedFilePath, destPath); err != nil {
			return err
		}

		return nil
	}

	// compose
	if err := installPlugin(
		"docker-compose",
		composeVersion,
		fmt.Sprintf("docker-compose-linux-%s", arch),
		"https://github.com/docker/compose/releases/download",
	); err != nil {
		return err
	}

	// buildx
	if err := installPlugin(
		"docker-buildx",
		buildxVersion,
		fmt.Sprintf("buildx-%s.linux-%s", buildxVersion, buildxArch),
		"https://github.com/docker/buildx/releases/download",
	); err != nil {
		return err
	}

	// 更新 ~/.docker/config.json
	if err = d.mergeJSONToFile(ctx, ui, env, utils.ExpandAbsDir("~/.docker/config.json"), map[string]interface{}{
		"cliPluginsExtraDirs": []string{pluginDir},
	}, false); err != nil {
		return err
	}
	if err = d.chmodFiles(ctx, ui, env, pluginDir); err != nil {
		return err
	}

	// ========= 第五步：启动服务 =========
	ui.Info("启动服务...")
	if err = utils.RunCommand(
		ctx, ui, env,
		"bash", "-c",
		"sudo systemctl daemon-reload && sudo systemctl enable docker && sudo systemctl start docker",
	); err != nil {
		ui.Error("启动服务失败")
		return err
	}
	ui.Info("检测docker信息...")
	if err = utils.RunCommand(ctx, ui, map[string]string{"PATH": binPath}, "docker", "info"); err != nil {
		ui.Error("等待服务启动成功失败")
		return err
	}

	envFile := path.Join(global.RootDir, ".env")
	err = utils.UpdateEnvFile(envFile, map[string]string{
		"PATH": binPath,
	}, "add")
	if err != nil {
		ui.Error("添加环境变量失败")
		return err
	}
	ui.Success("成功安装Docker: %s", installPath)
	currentUser, _ := utils.GetCurrentUser()
	homeDir, shell, _ := utils.GetUserHomeAndShell(currentUser.Username)
	profilePath := utils.GetProfilePath(shell, homeDir)
	ui.Success("安装完成！请执行以下操作：")
	ui.Info("1. 重新打开终端")
	ui.Info("2. 运行: source %s", profilePath)
	return nil
}

func (d *DockerManager) Uninstall(ctx context.Context) error {
	params, err := soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}
	ui := params.UI
	cfg := params.Cfg
	global := params.Global
	env := params.Env

	installPath := cfg.InstallDir
	if installPath == "" {
		installPath = path.Join(global.RootDir, "docker")
	}
	installPath = utils.ExpandAbsDir(installPath)

	ui.Info("停止并禁用 Docker 服务...")
	_ = utils.RunCommand(ctx, ui, env, "sudo", "systemctl", "stop", "docker")
	_ = utils.RunCommand(ctx, ui, env, "sudo", "systemctl", "disable", "docker")

	// 删除 service 文件
	ui.Info("删除 docker.service 文件...")
	serviceFile := "/etc/systemd/system/docker.service"
	if utils.PathExists(serviceFile) {
		ui.Info("删除 service 文件: %s", serviceFile)
		_ = utils.RunCommand(ctx, ui, env, "sudo", "rm", "-f", serviceFile)
	}
	_ = utils.RunCommand(ctx, ui, env, "sudo", "systemctl", "daemon-reload")

	// 删除 daemon.json
	daemonFile := "/etc/docker/daemon.json"
	if utils.PathExists(daemonFile) {
		ui.Info("删除 daemon.json 文件: %s", daemonFile)
		_ = utils.RunCommand(ctx, ui, env, "sudo", "rm", "-f", daemonFile)
	}

	// 删除安装目录
	if utils.PathExists(installPath) {
		ui.Info("删除安装目录: %s", installPath)
		_ = utils.RunCommand(ctx, ui, env, "sudo", "rm", "-rf", installPath)
	}

	// 删除用户配置 ~/.docker/config.json 中的插件路径
	clientConfigFile := utils.ExpandAbsDir("~/.docker/config.json")
	if utils.PathExists(clientConfigFile) {
		ui.Info("清理用户配置文件: %s", clientConfigFile)
		_ = utils.RunCommand(ctx, ui, env, "sudo", "rm", "-rf", clientConfigFile)
	}
	envFile := path.Join(global.RootDir, ".env")
	err = utils.UpdateEnvFile(envFile, map[string]string{
		"PATH": path.Join(installPath, "bin"),
	}, "remove")
	if err != nil {
		ui.Error("删除环境变量失败")
		return err
	}

	ui.Success("Docker 卸载完成")
	return nil
}

// 更新 Docker
func (d *DockerManager) Update(ctx context.Context) error {
	params, err := soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}
	ui := params.UI

	ui.Info("开始更新 Docker...")

	// 这里直接走 Install，Install 会覆盖已有文件 & 合并配置
	if err := d.Install(ctx); err != nil {
		ui.Error("更新 Docker 失败: %v", err)
		return err
	}

	ui.Success("Docker 更新完成")
	return nil
}

func init() {
	soft.Register("docker", &DockerManager{})
}
