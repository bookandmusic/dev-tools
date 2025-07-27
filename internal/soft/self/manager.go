package self

import (
	"os"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/bookandmusic/dev-tools/internal/config"
	"github.com/bookandmusic/dev-tools/internal/ui"
	"github.com/bookandmusic/dev-tools/internal/utils"
)

type InstallParams struct {
	InstallDir string
}

type SelfManager struct {
	Config        *config.GlobalConfig
	UI            ui.UI
	InstallParams *InstallParams
}

func (s *SelfManager) Uninstall() error {
	return nil
}

func (s *SelfManager) Install() error {
	var installDir string
	if s.InstallParams == nil {
		installDir = "~/.tools"
	} else if s.InstallParams.InstallDir == "" {
		installDir = "~/.tools"
	} else {
		installDir = s.InstallParams.InstallDir
	}

	rootAbsDir := utils.ExpandAbsDir(installDir)
	pluginDir := filepath.Join(rootAbsDir, "plugins")
	binDir := filepath.Join(rootAbsDir, "bin")

	s.UI.Info("Installing dev-tools to: %s", rootAbsDir)

	// 1️⃣ 创建安装目录
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		s.UI.Error("failed to create install directory: %w", err)
		return err
	}
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		s.UI.Error("failed to create plugins directory: %s", err)
		return err
	}

	// 2️⃣ 生成默认 YAML 配置文件
	configFile := filepath.Join(rootAbsDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		data, err := yaml.Marshal(config.GenerateDefaultConfig(installDir))
		if err != nil {
			s.UI.Error("failed to marshal config: %s", err)
			return err
		}
		if err := os.WriteFile(configFile, data, 0o600); err != nil { // Changed to 0600
			s.UI.Error("failed to create config file: %s", err)
			return err
		}
		s.UI.Info("Generated default config file: %s", configFile)
	} else {
		s.UI.Warning("Config file already exists, skipped generation: %s", configFile)
	}

	// 3️⃣ 将当前执行文件复制到 bin 目录
	execPath, err := os.Executable()
	if err != nil {
		s.UI.Error("failed to get current executable path: %s", err)
		return err
	}
	targetPath := filepath.Join(binDir, "dtl")

	if err := utils.CopyFile(execPath, targetPath); err != nil {
		s.UI.Error("failed to copy executable to bin: %s", err)
		return err
	}
	s.UI.Info("Copied executable to: %s", targetPath)

	// 4️⃣ 复制插件文件
	// 判断当前路径是否有 plugins 目录
	srcPlugins := filepath.Join(s.Config.Common.RootDir, "plugins")
	if info, err := os.Stat(srcPlugins); err == nil && info.IsDir() {
		if err := utils.CopyDirWithProgress(srcPlugins, pluginDir, s.UI); err != nil {
			s.UI.Error("Failed to copy plugins: %s", err)
			return err
		}
	} else {
		s.UI.Warning("No plugins directory found:%s", err)
		return err
	}

	// 5️⃣ 设置环境变量 & PATH（使用 bin 目录）
	var (
		currentUser *user.User
		homeDir     string
		shell       string
	)
	if currentUser, err = utils.GetCurrentUser(); err != nil {
		s.UI.Warning("Failed to get current user: %s", err)
		return err
	}
	if homeDir, shell, err = utils.GetUserHomeAndShell(currentUser.Username); err != nil {
		s.UI.Warning("Failed to get user home and shell: %s", err)
		return err

	}
	envFile := utils.GetProfilePath(shell, homeDir)
	err = utils.AddEnvToEnvFile(map[string]string{
		"DEV_TOOLS_HOME": rootAbsDir,
	}, envFile)
	if err != nil {
		s.UI.Error("Failed to setup environment variables: %v", err)
		return err
	}

	s.UI.Info("Installation complete.\nPlease restart your shell or run 'source %s' to apply changes.\n", envFile)
	return nil
}
