package ohmyzsh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
)

func NewOhMyzshManager(ui ui.UI, cfg *config.OhMyzshConfig, global *config.CommonConfig) *OhMyzshManager {
	return &OhMyzshManager{
		ui:     ui,
		cfg:    cfg,
		global: global,
	}
}

// OhMyzshManager 实现OhMyZsh管理
type OhMyzshManager struct {
	ui     ui.UI
	global *config.CommonConfig
	cfg    *config.OhMyzshConfig
}

// proxyURL 添加GitHub代理前缀
func (o *OhMyzshManager) proxyURL(url string) string {
	if o.global.GithubProxy != "" {
		o.ui.Debug("使用代理访问: %s -> %s%s", url, o.global.GithubProxy, url)
		if !strings.HasSuffix(o.global.GithubProxy, "/") {
			o.global.GithubProxy += "/"
		}
		return fmt.Sprintf("%s%s", o.global.GithubProxy, url)
	}
	return url
}

// runCmd 执行命令并处理输出
func (o *OhMyzshManager) runCmd(name string, args ...string) error {
	// 显示简化命令信息
	displayCmd := name
	if len(args) > 0 {
		displayCmd += " " + args[0]
		if len(args) > 1 {
			displayCmd += " ..."
		}
	}
	o.ui.Info("执行命令: %s", displayCmd)

	// 在调试模式下显示完整命令
	o.ui.Debug("完整命令: %s %s", name, strings.Join(args, " "))

	// 执行命令并捕获输出
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	// 在调试模式下显示完整输出
	if o.global.Debug {
		if stdoutBuf.Len() > 0 {
			o.ui.Debug("标准输出:\n%s", stdoutBuf.String())
		}
		if stderrBuf.Len() > 0 {
			o.ui.Debug("标准错误:\n%s", stderrBuf.String())
		}
	}

	if err != nil {
		// 错误时显示关键错误信息
		o.ui.Error("命令执行失败: %s", displayCmd)
		if stderrBuf.Len() > 0 {
			// 截取关键错误信息
			errorLines := strings.Split(stderrBuf.String(), "\n")
			if len(errorLines) > 0 {
				firstError := strings.TrimSpace(errorLines[0])
				if firstError != "" {
					o.ui.Error("错误信息: %s", firstError)
				}
			}
		}

		// 调试模式下显示完整错误
		if o.global.Debug {
			o.ui.Debug("完整错误: %v", err)
		}
		return fmt.Errorf("%s 命令失败: %w", name, err)
	}

	return nil
}

// isGitRepo 检查目录是否是有效的git仓库
func (o *OhMyzshManager) isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		o.ui.Debug("路径 %s 不是Git仓库: 缺少.git目录", path)
		return false
	}

	// 验证git仓库状态
	err := o.runCmd("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		o.ui.Debug("路径 %s 不是有效的Git仓库: %v", path, err)
		return false
	}

	o.ui.Debug("路径 %s 是有效的Git仓库", path)
	return true
}

// installDependencies 安装系统依赖
func (o *OhMyzshManager) installDependencies() error {
	if runtime.GOOS == "darwin" {
		o.ui.Info("macOS系统，跳过依赖安装")
		return nil
	}

	if _, err := exec.LookPath("zsh"); err == nil {
		o.ui.Info("zsh已安装")
		return nil
	}

	o.ui.Info("安装zsh和依赖...")
	switch {
	case pathExists("/etc/debian_version"):
		return o.runCmd("sudo", "apt", "update", "-y")
	case pathExists("/etc/redhat-release"):
		return o.runCmd("sudo", "yum", "install", "-y", "zsh", "git", "curl")
	case pathExists("/etc/arch-release"):
		return o.runCmd("sudo", "pacman", "-S", "--noconfirm", "zsh", "git", "curl")
	default:
		return fmt.Errorf("未知或不支持的Linux发行版")
	}
}

// setDefaultShell 设置zsh为默认shell
func (o *OhMyzshManager) setDefaultShell() error {
	if runtime.GOOS == "darwin" {
		o.ui.Info("macOS系统，跳过修改默认shell")
		return nil
	}

	currentShell := filepath.Base(os.Getenv("SHELL"))
	if currentShell == "zsh" {
		o.ui.Info("默认shell已是zsh")
		return nil
	}

	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		return fmt.Errorf("找不到zsh: %w", err)
	}

	o.ui.Info("修改默认shell为zsh")
	return o.runCmd("chsh", "-s", zshPath)
}

// cloneRepository 克隆OhMyZsh仓库
func (o *OhMyzshManager) cloneRepository() error {
	if o.isGitRepo(o.cfg.InstallDir) {
		o.ui.Info("OhMyZsh已安装: %s", o.cfg.InstallDir)
		return nil
	}

	// 目录存在但不是git仓库
	if pathExists(o.cfg.InstallDir) {
		o.ui.Warning("目录 %s 存在但不是git仓库，尝试备份", o.cfg.InstallDir)
		backupDir := o.cfg.InstallDir + ".bak"
		if err := os.Rename(o.cfg.InstallDir, backupDir); err != nil {
			return fmt.Errorf("备份目录失败: %w", err)
		}
		o.ui.Info("已备份到: %s", backupDir)
	}

	o.ui.Info("克隆OhMyZsh到: %s", o.cfg.InstallDir)
	url := o.proxyURL("https://github.com/ohmyzsh/ohmyzsh.git")
	return o.runCmd("git", "clone", "--depth=1", url, o.cfg.InstallDir)
}

// setupZshrc 配置.zshrc文件
func (o *OhMyzshManager) setupZshrc() error {
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")

	if pathExists(zshrcPath) {
		o.ui.Info(".zshrc已存在，跳过生成")
		return nil
	}

	templatePath := filepath.Join(o.cfg.InstallDir, "templates", "zshrc.zsh-template")
	o.ui.Info("生成.zshrc配置文件")
	return sh.Copy(zshrcPath, templatePath)
}

// installPlugins 安装所有插件
func (o *OhMyzshManager) installPlugins() error {
	pluginsDir := filepath.Join(o.cfg.InstallDir, "custom", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return err
	}

	o.ui.Info("安装插件...")
	for _, plugin := range o.cfg.Plugins {
		pluginName := filepath.Base(plugin.Name)
		pluginPath := filepath.Join(pluginsDir, pluginName)

		if o.isGitRepo(pluginPath) {
			o.ui.Info("插件 %s 已安装", pluginName)
			continue
		}

		// 目录存在但不是git仓库
		if pathExists(pluginPath) {
			o.ui.Warning("插件目录 %s 存在但不是git仓库，删除", pluginPath)
			if err := os.RemoveAll(pluginPath); err != nil {
				return err
			}
		}

		o.ui.Info("安装插件: %s", plugin.Name)
		url := o.proxyURL(fmt.Sprintf("https://github.com/%s.git", plugin.Repo))
		if err := o.runCmd("git", "clone", "--depth=1", url, pluginPath); err != nil {
			return err
		}
	}
	return nil
}

// configureZshrc 配置.zshrc文件
func (o *OhMyzshManager) configureZshrc() error {
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")

	// 配置ZSH环境变量
	zshLine := fmt.Sprintf(`export ZSH=%q`, o.cfg.InstallDir)
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

	// 设置主题
	themeLine := fmt.Sprintf("ZSH_THEME=%q", o.cfg.Theme)
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
	for _, plugin := range o.cfg.Plugins {
		plugins += fmt.Sprintf(" %s", plugin.Name)
	}
	// 设置插件
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

	// 写入更新后的配置
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(zshrcPath, []byte(newContent), 0o644); err != nil {
		return err
	}

	o.ui.Info("已配置主题: %s", o.cfg.Theme)
	o.ui.Info("已启用插件: %v", plugins)
	return nil
}

// Install 实现SoftManage接口的安装方法
func (o *OhMyzshManager) Install(installDir string) error {
	o.ui.Info("开始安装Oh My Zsh...")
	if installDir != "" {
		o.cfg.InstallDir = installDir
	}
	steps := []func() error{
		o.installDependencies,
		o.setDefaultShell,
		o.cloneRepository,
		o.setupZshrc,
		o.installPlugins,
		o.configureZshrc,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			o.ui.Error("安装失败: %v", err)
			return fmt.Errorf("安装失败: %w", err)
		}
	}

	o.ui.Info(`
✅ 安装完成！请执行以下操作：
  1. 重新打开终端
  2. 或运行: source ~/.zshrc`)
	return nil
}

// Update 实现SoftManage接口的更新方法
func (o *OhMyzshManager) Update() error {
	o.ui.Info("更新Oh My Zsh...")

	if !o.isGitRepo(o.cfg.InstallDir) {
		return fmt.Errorf("oh my zsh未安装或安装损坏")
	}

	// 更新主仓库
	o.ui.Info("更新主仓库...")
	if err := o.runCmd("git", "-C", o.cfg.InstallDir, "pull"); err != nil {
		return err
	}

	// 更新所有插件
	o.ui.Info("更新插件...")
	pluginsDir := filepath.Join(o.cfg.InstallDir, "custom", "plugins")
	for _, plugin := range o.cfg.Plugins {
		pluginName := filepath.Base(plugin.Name)
		pluginPath := filepath.Join(pluginsDir, pluginName)

		if o.isGitRepo(pluginPath) {
			o.ui.Info("更新插件: %s", pluginName)
			if err := o.runCmd("git", "-C", pluginPath, "pull"); err != nil {
				o.ui.Warning("插件 %s 更新失败: %v", pluginName, err)
			}
		} else {
			o.ui.Warning("插件 %s 未安装或安装损坏", pluginName)
		}
	}

	o.ui.Info(`
✅ 更新完成！请运行: source ~/.zshrc`)
	return nil
}

// Uninstall 实现SoftManage接口的卸载方法
func (o *OhMyzshManager) Uninstall() error {
	o.ui.Info("卸载Oh My Zsh...")

	// 删除安装目录
	if pathExists(o.cfg.InstallDir) {
		o.ui.Info("删除目录: %s", o.cfg.InstallDir)
		if err := os.RemoveAll(o.cfg.InstallDir); err != nil {
			return err
		}
	}

	// 备份.zshrc
	zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
	if pathExists(zshrcPath) {
		o.ui.Info("备份.zshrc: %s -> %s.bak", zshrcPath, zshrcPath)
		if err := os.Rename(zshrcPath, zshrcPath+".bak"); err != nil {
			o.ui.Warning("备份.zshrc失败: %v", err)
		}
	}

	o.ui.Info(`
✅ 卸载完成！请重新启动终端`)
	return nil
}

// pathExists 检查路径是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
