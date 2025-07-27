package ohmyzsh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/manager/soft"
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

// OhMyzshManager 实现OhMyZsh管理
type OhMyzshManager struct{}

// installDependencies 安装系统依赖
func (o *OhMyzshManager) installDependencies(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	if runtime.GOOS == "darwin" {
		ui.Info("macOS系统，跳过依赖安装")
		return nil
	}

	if _, err := exec.LookPath("zsh"); err == nil {
		ui.Info("zsh已安装")
		return nil
	}

	ui.Info("安装zsh和依赖...")
	switch {
	case utils.PathExists("/etc/debian_version"):
		return utils.RunCommand(ctx, ui, env, "sudo", "apt", "update", "-y")
	case utils.PathExists("/etc/redhat-release"):
		return utils.RunCommand(ctx, ui, env, "sudo", "yum", "install", "-y", "zsh", "git", "curl")
	case utils.PathExists("/etc/arch-release"):
		return utils.RunCommand(ctx, ui, env, "sudo", "pacman", "-S", "--noconfirm", "zsh", "git", "curl")
	default:
		return fmt.Errorf("未知或不支持的Linux发行版")
	}
}

// setDefaultShell 设置zsh为默认shell
func (o *OhMyzshManager) setDefaultShell(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	if runtime.GOOS == "darwin" {
		ui.Info("macOS系统，跳过修改默认shell")
		return nil
	}
	currentShell := filepath.Base(os.Getenv("SHELL"))
	if currentShell == "zsh" {
		ui.Info("默认shell已是zsh")
		return nil
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		return fmt.Errorf("找不到zsh: %w", err)
	}

	ui.Info("修改默认shell为zsh")
	return utils.RunCommand(ctx, ui, env, "chsh", "-s", zshPath)
}

// cloneRepository 克隆OhMyZsh仓库
func (o *OhMyzshManager) cloneRepository(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	path := cfg.InstallDir
	return utils.CloneRepoWithProxy(
		ctx,
		ui,
		"https://github.com/ohmyzsh/ohmyzsh.git",
		path, global.HttpProxy, global.GithubProxy, env,
	)
}

// setupZshrc 配置.zshrc文件
func (o *OhMyzshManager) setupZshrc(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")

	if utils.PathExists(zshrcPath) {
		ui.Info(".zshrc已存在，跳过生成")
		return nil
	}

	templatePath := filepath.Join(cfg.InstallDir, "templates", "zshrc.zsh-template")
	ui.Info("生成.zshrc配置文件")
	return utils.CopyFile(templatePath, zshrcPath)
}

// installPlugins 安装所有插件
func (o *OhMyzshManager) installPlugins(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	pluginsDir := filepath.Join(cfg.InstallDir, "custom", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o700); err != nil {
		return err
	}

	ui.Info("安装插件...")
	for _, plugin := range cfg.Plugins {
		pluginName := filepath.Base(plugin.Name)
		pluginPath := filepath.Join(pluginsDir, pluginName)
		url := fmt.Sprintf("https://github.com/%s.git", plugin.Repo)
		if err := utils.CloneRepoWithProxy(
			ctx, ui, url, pluginPath, global.HttpProxy, global.GithubProxy, env,
		); err != nil {
			return err
		}
	}
	return nil
}

// configureZshrc 配置.zshrc文件
func (o *OhMyzshManager) configureZshrc(ctx context.Context, ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig, env map[string]string) error {
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	zshLine := fmt.Sprintf(`export ZSH=%q`, cfg.InstallDir)
	zshSet := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "export ZSH=") {
			lines[i] = zshLine
			zshSet = true
			break
		}
	}
	if !zshSet {
		lines = append([]string{zshLine}, lines...)
	}

	themeLine := fmt.Sprintf("ZSH_THEME=%q", cfg.Theme)
	themeSet := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "ZSH_THEME=") {
			lines[i] = themeLine
			themeSet = true
			break
		}
	}
	if !themeSet {
		lines = append([]string{themeLine}, lines...)
	}

	plugins := ""
	for _, plugin := range cfg.Plugins {
		plugins += fmt.Sprintf(" %s", plugin.Name)
	}
	pluginLine := fmt.Sprintf(`plugins=(git sudo extract z %s)`, plugins)
	pluginSet := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "plugins=(") {
			lines[i] = pluginLine
			pluginSet = true
			break
		}
	}
	if !pluginSet {
		lines = append(lines, pluginLine)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(zshrcPath, []byte(newContent), 0o600); err != nil {
		return err
	}

	ui.Info("已配置主题: %s", cfg.Theme)
	ui.Info("已启用插件: %v", plugins)
	return nil
}

// Install 实现SoftManage接口的安装方法
func (o *OhMyzshManager) Install(ctx context.Context) error {
	params, err := soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}
	console := params.UI
	console.Info("开始安装Oh My Zsh...")

	steps := []func(context.Context, ui.UI, *config.OhMyzshConfig, *config.CommonConfig, map[string]string) error{
		o.installDependencies,
		o.setDefaultShell,
		o.cloneRepository,
		o.setupZshrc,
		o.installPlugins,
		o.configureZshrc,
	}

	for _, step := range steps {
		if err := step(ctx, params.UI, params.Cfg, params.Global, params.Env); err != nil {
			console.Error("安装失败: %v", err)
			return fmt.Errorf("安装失败: %w", err)
		}
	}

	console.Success("安装完成！请执行以下操作：")
	console.Info("1. 重新打开终端")
	console.Info("2. 运行: source ~/.zshrc")
	return nil
}

// Update 实现SoftManage接口的更新方法
func (o *OhMyzshManager) Update(ctx context.Context) error {
	params, err := soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}
	cfg := params.Cfg
	ui := params.UI
	env := params.Env
	ui.Info("更新Oh My Zsh...")
	if err := utils.UpdateRepo(ctx, ui, cfg.InstallDir, env, params.Global.HttpProxy); err != nil {
		return err
	}

	ui.Info("更新插件...")
	pluginsDir := filepath.Join(cfg.InstallDir, "custom", "plugins")
	for _, plugin := range cfg.Plugins {
		pluginName := filepath.Base(plugin.Name)
		pluginPath := filepath.Join(pluginsDir, pluginName)
		if err := utils.UpdateRepo(ctx, ui, pluginPath, env, params.Global.HttpProxy); err != nil {
			return err
		}
	}

	ui.Success(`更新完成！请运行: source ~/.zshrc`)
	return nil
}

// Uninstall 实现SoftManage接口的卸载方法
func (o *OhMyzshManager) Uninstall(ctx context.Context) error {
	params, err := soft.Parse[BaseParams](ctx)
	if err != nil {
		return err
	}
	cfg := params.Cfg
	ui := params.UI
	ui.Info("卸载Oh My Zsh...")

	if utils.PathExists(cfg.InstallDir) {
		ui.Info("删除目录: %s", cfg.InstallDir)
		if err := os.RemoveAll(cfg.InstallDir); err != nil {
			return err
		}
	}

	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
	if utils.PathExists(zshrcPath) {
		ui.Info("备份.zshrc: %s -> %s.bak", zshrcPath, zshrcPath)
		if err := os.Rename(zshrcPath, zshrcPath+".bak"); err != nil {
			ui.Warning("备份.zshrc失败: %v", err)
		}
	}

	ui.Success(`卸载完成！请重新启动终端`)
	return nil
}

func init() {
	soft.Register("ohmyzsh", &OhMyzshManager{})
}
