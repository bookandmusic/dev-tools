package config

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type AnsibleConfig struct {
	BaseDir    string `yaml:"base-dir"`
	PythonDir  string `yaml:"python-dir"`
	AnsibleDir string `yaml:"ansible-dir"`
}

type LangConfig struct {
	BaseDir  string   `yaml:"base-dir"`
	Versions []string `yaml:"versions"`
	Global   string   `yaml:"global"`
}

type GoConfig struct {
	BaseDir  string   `yaml:"base-dir"`
	Versions []string `yaml:"versions"`
	Global   string   `yaml:"global"`
}

type CommonConfig struct {
	Debug       bool   `yaml:"debug"`
	RootDir     string `yaml:"root-dir"`
	CacheDir    string `yaml:"cache-dir"`
	GithubProxy string `yaml:"github-proxy"`
	HttpProxy   string `yaml:"http-proxy"`
}

type GlobalConfig struct {
	Common  *CommonConfig  `yaml:"common"`
	Ansible *AnsibleConfig `yaml:"ansible"`
	Python  *LangConfig    `yaml:"python"`
	Go      *LangConfig    `yaml:"go"`
}

func NewDefaultHomeDir() string {
	// 默认路径
	userHome, _ := os.UserHomeDir()

	return filepath.Join(userHome, ".tools")
}

func GenerateDefaultConfig(baseDir string) *GlobalConfig {
	return &GlobalConfig{
		Common: &CommonConfig{
			Debug:    false,
			RootDir:  baseDir,
			CacheDir: filepath.Join(baseDir, "cache"),
		},
		Ansible: &AnsibleConfig{
			BaseDir:    filepath.Join(baseDir, "ansible"),
			PythonDir:  filepath.Join(baseDir, "ansible", "python"),
			AnsibleDir: filepath.Join(baseDir, "ansible", "ansible"),
		},

		Python: &LangConfig{
			BaseDir: filepath.Join(baseDir, "python"),
		},
		Go: &LangConfig{
			BaseDir: filepath.Join(baseDir, "go"),
		},
	}
}

func LoadConfigFile(debug bool, configFile string) (*GlobalConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	config.Common.Debug = debug
	return &config, nil
}
