package config

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/bookandmusic/dev-tools/internal/utils"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) DetermineWorkDir(cfgRootDir string, rootDir string, rootDirChange bool) string {
	// 1. 配置文件中的 rootDir (最高优先级)
	if cfgRootDir != "" {
		return cfgRootDir
	}

	// 2. --root-dir 命令行参数
	if rootDirChange && rootDir != "" {
		return rootDir
	}

	// 3. 环境变量 DEV_TOOLS_HOME
	if envDir := os.Getenv("DEV_TOOLS_HOME"); envDir != "" {
		return envDir
	}

	// 4. 当前目录 (最低优先级)
	return "."
}

func (m *Manager) DetermineRootDir(
	cfgRootDir string,
	rootDir string,
	rootDirChange bool,
) string {
	if rootDirChange && rootDir != "" {
		return rootDir
	}
	if cfgRootDir != "" {
		return cfgRootDir
	}
	if rootDir != "" {
		return rootDir
	}
	return "~/.tools"
}

func (m *Manager) DetermineConfigPath(
	configFile, rootDir string,
	configChange, rootDirChange bool,
) string {
	// 1. --config 指定的配置文件 (最高优先级)
	if configFile != "" {
		return configFile
	}

	// 2. --root-dir 目录下的 config.yml
	if rootDir != "" {
		return filepath.Join(rootDir, "config.yml")
	}

	// 3. 环境变量 DEV_TOOLS_HOME 下的 config.yml
	if envDir := os.Getenv("DEV_TOOLS_HOME"); envDir != "" {
		return filepath.Join(envDir, "config.yml")
	}

	// 4. 当前目录下的 config.yml (最低优先级)
	if currentDir, err := os.Getwd(); err == nil {
		return filepath.Join(currentDir, "config.yml")
	}

	return ""
}

// LoadConfigWithFallback 按优先级顺序尝试从多个路径加载配置
func (m *Manager) LoadConfigWithFallback(
	userConfigFile, userRootDir string,
	configChanged, rootDirChanged bool,
) (*GlobalConfig, error) {
	// 定义尝试路径的顺序
	tryPaths := []struct {
		path    string
		enabled bool
	}{
		// 1. 用户指定的配置文件 (最高优先级)
		{path: userConfigFile, enabled: configChanged && userConfigFile != ""},
		// 2. 用户指定的root目录下的config.yml
		{path: filepath.Join(userRootDir, "config.yml"), enabled: rootDirChanged && userRootDir != ""},
		// 3. 环境变量 DEV_TOOLS_HOME 下的config.yml
		{path: filepath.Join(os.Getenv("DEV_TOOLS_HOME"), "config.yml"), enabled: os.Getenv("DEV_TOOLS_HOME") != ""},
		// 4. 默认的配置文件路径
		{path: userConfigFile, enabled: userConfigFile != ""},
		// 5. 默认的root目录下的config.yml
		{path: filepath.Join(userRootDir, "config.yml"), enabled: userRootDir != ""},
		// 6. 当前目录下的config.yml (最低优先级)
		{path: filepath.Join(m.getWorkingDir(), "config.yml"), enabled: true},
	}

	var lastErr error

	// 按优先级顺序尝试加载配置
	for _, pathInfo := range tryPaths {
		if !pathInfo.enabled {
			continue
		}

		cfg, err := m.LoadConfig(pathInfo.path)
		if err != nil {
			// 记录错误，继续尝试下一个路径
			lastErr = err
			continue
		}

		// 成功加载配置
		return cfg, nil
	}

	// 所有路径尝试失败，返回最后一个错误或空配置
	if lastErr != nil {
		return nil, lastErr
	}

	// 没有任何配置文件存在，返回空配置
	return m.LoadConfig("")
}

// getWorkingDir 获取当前工作目录
func (m *Manager) getWorkingDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "."
}

func (m *Manager) LoadConfig(
	cfgPath string,
) (*GlobalConfig, error) {
	var cfg GlobalConfig
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (m *Manager) SetDefaults(cfg *GlobalConfig, rootDir string) {
	if rootDir == "" {
		rootDir = "~/.config"
	}
	rootDir = utils.ExpandAbsDir(rootDir)

	// 设置 Common 默认值
	m.setCommonDefaults(cfg, rootDir)

	// 设置 Ansible 默认值
	m.setAnsibleDefaults(cfg)

	// 设置 Python 默认值
	m.setPythonDefaults(cfg)

	// 设置 Go 默认值
	m.setGoDefaults(cfg)

	// 设置 OhMyZsh 默认值
	m.setOhMyzshDefaults(cfg)

	// 设置 Docker 默认值
	m.setDockerDefaults(cfg)
}

func (m *Manager) setCommonDefaults(cfg *GlobalConfig, rootDir string) {
	// 确保 Common 配置存在
	if cfg.Common == nil {
		cfg.Common = &CommonConfig{}
	}

	// 设置 Common 默认值
	if cfg.Common.RootDir == "" {
		cfg.Common.RootDir = rootDir
	}

	if cfg.Common.CacheDir == "" {
		cfg.Common.CacheDir = filepath.Join(cfg.Common.RootDir, "cache")
	}
}

func (m *Manager) setAnsibleDefaults(cfg *GlobalConfig) {
	// 设置 Ansible 默认值
	if cfg.Ansible == nil {
		cfg.Ansible = &AnsibleConfig{}
	}
	if cfg.Ansible.BaseDir == "" {
		cfg.Ansible.BaseDir = filepath.Join(cfg.Common.RootDir, "ansible")
	}
	if cfg.Ansible.PythonDir == "" {
		cfg.Ansible.PythonDir = filepath.Join(cfg.Ansible.BaseDir, "python")
	}
	if cfg.Ansible.AnsibleDir == "" {
		cfg.Ansible.AnsibleDir = filepath.Join(cfg.Ansible.BaseDir, "ansible")
	}
}

func (m *Manager) setPythonDefaults(cfg *GlobalConfig) {
	// 设置 Python 默认值
	if cfg.Python == nil {
		cfg.Python = &LangConfig{}
	}
	if cfg.Python.BaseDir == "" {
		cfg.Python.BaseDir = filepath.Join(cfg.Common.RootDir, "python")
	}
}

func (m *Manager) setGoDefaults(cfg *GlobalConfig) {
	// 设置 Go 默认值
	if cfg.Go == nil {
		cfg.Go = &LangConfig{}
	}
	if cfg.Go.BaseDir == "" {
		cfg.Go.BaseDir = filepath.Join(cfg.Common.RootDir, "go")
	}
}

func (m *Manager) setOhMyzshDefaults(cfg *GlobalConfig) {
	// 设置 OhMyZsh 默认值
	if cfg.OhMyzsh == nil {
		cfg.OhMyzsh = &OhMyzshConfig{}
	}
	if cfg.OhMyzsh.InstallDir == "" {
		cfg.OhMyzsh.InstallDir = filepath.Join(cfg.Common.RootDir, "oh-my-zsh")
	}
	if cfg.OhMyzsh.Theme == "" {
		cfg.OhMyzsh.Theme = "robbyrussell" // 默认主题
	}
	if len(cfg.OhMyzsh.Plugins) == 0 {
		cfg.OhMyzsh.Plugins = []*OhMyzshPlugin{
			{
				Name: "zsh-autosuggestions",
				Repo: "zsh-users/zsh-autosuggestions",
			},
			{
				Name: "zsh-history-substring-search",
				Repo: "zsh-users/zsh-history-substring-search",
			},
			{
				Name: "zsh-completions",
				Repo: "zsh-users/zsh-completions",
			},
			{
				Name: "zsh-syntax-highlighting",
				Repo: "zsh-users/zsh-syntax-highlighting",
			},
		}
	}
}

func (m *Manager) setDockerDefaults(cfg *GlobalConfig) {
	// 设置 Docker 默认值
	if cfg.Docker == nil {
		cfg.Docker = &DockerConfig{}
	}
	if cfg.Docker.Version == "" {
		cfg.Docker.Version = "26.1.0"
	}
	if cfg.Docker.InstallDir == "" {
		cfg.Docker.InstallDir = filepath.Join(cfg.Common.RootDir, "docker")
	}
	if cfg.Docker.HttpProxy == "" {
		cfg.Docker.HttpProxy = cfg.Common.HttpProxy
	}
}

func (m *Manager) UpdateRootDir(cfg *GlobalConfig, rootDir string) {
	rootDir = utils.ExpandAbsDir(rootDir)

	cfg.Common.RootDir = rootDir
	cfg.Common.CacheDir = filepath.Join(rootDir, "cache")

	cfg.Ansible.BaseDir = filepath.Join(cfg.Common.RootDir, "ansible")

	cfg.Ansible.PythonDir = filepath.Join(cfg.Ansible.BaseDir, "python")

	cfg.Ansible.AnsibleDir = filepath.Join(cfg.Ansible.BaseDir, "ansible")

	cfg.Python.BaseDir = filepath.Join(cfg.Common.RootDir, "python")

	cfg.Go.BaseDir = filepath.Join(cfg.Common.RootDir, "go")

	cfg.OhMyzsh.InstallDir = filepath.Join(cfg.Common.RootDir, "oh-my-zsh")
	cfg.Docker.InstallDir = filepath.Join(cfg.Common.RootDir, "docker")
}
